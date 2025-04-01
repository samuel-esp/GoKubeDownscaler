package scalable

import (
	"fmt"
	"log/slog"
	"slices"
	"strconv"
	"strings"

	"github.com/caas-team/gokubedownscaler/internal/pkg/util"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
)

const annotationOriginalReplicas = "downscaler/original-replicas"

// FilterExcluded filters the workloads to match the includeLabels, excludedNamespaces and excludedWorkloads.
func FilterExcluded(
	workloads []Workload,
	includeLabels,
	excludedNamespaces,
	excludedWorkloads util.RegexList,
	includedResources map[string]struct{},
	workloadsUIDToWorkload map[types.UID]*Workload,
) []Workload {
	externallyScaled := getExternallyScaled(workloads)

	results := make([]Workload, 0, len(workloads))

	for _, workload := range workloads {
		if !isMatchingLabels(workload, includeLabels) {
			slog.Debug(
				"workload is not matching any of the specified labels, excluding it from being scanned",
				"workload", workload.GetName(),
				"namespace", workload.GetNamespace(),
			)

			continue
		}

		if isNamespaceExcluded(workload, excludedNamespaces) {
			slog.Debug(
				"the workloads namespace is excluded, excluding it from being scanned",
				"workload", workload.GetName(),
				"namespace", workload.GetNamespace(),
			)

			continue
		}

		if isWorkloadExcluded(workload, excludedWorkloads, includedResources, workloadsUIDToWorkload) {
			slog.Debug(
				"the workloads name is excluded, excluding it from being scanned",
				"workload", workload.GetName(),
				"namespace", workload.GetNamespace(),
			)

			continue
		}

		if isExternallyScaled(workload, externallyScaled) {
			slog.Debug(
				"the workload is scaled externally, excluding it from being scanned",
				"workload", workload.GetName(),
				"namespace", workload.GetNamespace(),
			)

			continue
		}

		results = append(results, workload)
	}

	return slices.Clip(results)
}

type workloadIdentifier struct {
	gvk       schema.GroupVersionKind
	name      string
	namespace string
}

// getExternallyScaled returns identifiers for workloads which are being scaled externally and should therefore be excluded.
func getExternallyScaled(workloads []Workload) []workloadIdentifier {
	externallyScaled := make([]workloadIdentifier, 0, len(workloads))

	for _, workload := range workloads {
		scaledobject := getWorkloadAsScaledObject(workload)
		if scaledobject == nil {
			continue
		}

		externallyScaled = append(externallyScaled, workloadIdentifier{
			gvk: schema.GroupVersionKind{
				Kind:    scaledobject.Spec.ScaleTargetRef.Kind,
				Group:   strings.Split(scaledobject.Spec.ScaleTargetRef.APIVersion, "/")[0],
				Version: strings.Split(scaledobject.Spec.ScaleTargetRef.APIVersion, "/")[1],
			},
			name:      scaledobject.Spec.ScaleTargetRef.Name,
			namespace: scaledobject.Namespace,
		})
	}

	return slices.Clip(externallyScaled)
}

// isExternallyScaled checks if the workload matches any of the given workload identifiers.
func isExternallyScaled(workload Workload, externallyScaled []workloadIdentifier) bool {
	for _, wid := range externallyScaled {
		if wid.name != workload.GetName() {
			continue
		}

		if wid.namespace != workload.GetNamespace() {
			continue
		}

		if !(wid.gvk.Group == "" || wid.gvk.Group == workload.GroupVersionKind().Group) {
			continue
		}

		if !(wid.gvk.Version == "" || wid.gvk.Version == workload.GroupVersionKind().Version) {
			continue
		}

		if !(wid.gvk.Kind == "" || wid.gvk.Kind == workload.GroupVersionKind().Kind) {
			continue
		}

		return true
	}

	return false
}

// getWorkloadAsScaledObject tries to get the given workload as a scaled object.
func getWorkloadAsScaledObject(workload Workload) *scaledObject {
	replicaScaled, isReplicaScaled := workload.(*replicaScaledWorkload)
	if !isReplicaScaled {
		return nil
	}

	scaledObject, isScaledObject := replicaScaled.replicaScaledResource.(*scaledObject)
	if !isScaledObject {
		return nil
	}

	return scaledObject
}

// isMatchingLabels check if the workload is matching any of the specified labels.
func isMatchingLabels(workload Workload, includeLabels util.RegexList) bool {
	if includeLabels == nil {
		return true
	}

	for label, value := range workload.GetLabels() {
		if !includeLabels.CheckMatchesAny(fmt.Sprintf("%s=%s", label, value)) {
			continue
		}

		return true
	}

	return false
}

// isNamespaceExcluded checks if the workloads namespace is excluded.
func isNamespaceExcluded(workload Workload, excludedNamespaces util.RegexList) bool {
	if excludedNamespaces == nil {
		return false
	}

	return excludedNamespaces.CheckMatchesAny(workload.GetNamespace())
}

// isWorkloadExcluded checks if the workloads name is excluded.
func isWorkloadExcluded(
	workload Workload,
	excludedWorkloads util.RegexList,
	includedResources map[string]struct{},
	uidToWorkload map[types.UID]*Workload,
) bool {
	if excludedWorkloads == nil {
		return false
	}

	nonControllerOwnerReferences := getNonControllerOwnerReferences(&workload, uidToWorkload)
	if checkAllReferencesExcluded(&nonControllerOwnerReferences, includedResources) {
		return true
	}

	return excludedWorkloads.CheckMatchesAny(workload.GetName())
}

// checkAllReferencesExcluded checks if all owner references of the workload are excluded.
func checkAllReferencesExcluded(ownerReferences *[]Workload, includedResources map[string]struct{}) bool {
	if len(*ownerReferences) == 0 {
		return false
	}

	ownerReferencesExcluded := true

	for _, ownerReference := range *ownerReferences {
		if _, exists := includedResources[ownerReference.GroupVersionKind().String()]; exists {
			ownerReferencesExcluded = false
			break
		}
	}

	return ownerReferencesExcluded
}

// getOwnerReferenceParent recursively finds the root owner of the workload by following controller owner references.
//
//nolint:gocritic // this function should return a pointer
func getOwnerReferenceParent(workload *Workload, uidToWorkload map[types.UID]*Workload) *Workload {
	for _, ownerReference := range (*workload).GetOwnerReferences() {
		if ownerReference.Controller != nil && *ownerReference.Controller {
			if parentWorkload, exists := uidToWorkload[ownerReference.UID]; exists {
				return getOwnerReferenceParent(parentWorkload, uidToWorkload)
			}
		}
	}

	return workload
}

// getNonControllerOwnerReferences returns non-controller owner references of the workload.
//
//nolint:gocritic // this function should return a pointer
func getNonControllerOwnerReferences(
	workload *Workload,
	uidToWorkload map[types.UID]*Workload,
) []Workload {
	var nonControllerResources []Workload

	for _, ownerReference := range (*workload).GetOwnerReferences() {
		if ownerReference.Controller != nil && *ownerReference.Controller {
			if rootOwner := getOwnerReferenceParent(uidToWorkload[ownerReference.UID], uidToWorkload); rootOwner == workload {
				continue
			}
		} else {
			if resource, exists := uidToWorkload[ownerReference.UID]; exists {
				nonControllerResources = append(nonControllerResources, *resource)
			}
		}
	}

	return nonControllerResources
}

// setOriginalReplicas sets the original replicas annotation on the workload.
func setOriginalReplicas(originalReplicas int32, workload Workload) {
	annotations := workload.GetAnnotations()
	if annotations == nil {
		annotations = map[string]string{}
	}

	annotations[annotationOriginalReplicas] = strconv.Itoa(int(originalReplicas))
	workload.SetAnnotations(annotations)
}

// getOriginalReplicas gets the original replicas annotation on the workload. nil is undefined.
func getOriginalReplicas(workload Workload) (*int32, error) {
	annotations := workload.GetAnnotations()

	originalReplicasString, ok := annotations[annotationOriginalReplicas]
	if !ok {
		return nil, nil //nolint: nilnil // should get fixed along with https://github.com/caas-team/GoKubeDownscaler/issues/7
	}

	originalReplicas, err := strconv.ParseInt(originalReplicasString, 10, 32)
	if err != nil {
		return nil, fmt.Errorf("failed to parse original replicas annotation on workload: %w", err)
	}

	// #nosec G115
	result := int32(originalReplicas)

	return &result, nil
}

// removeOriginalReplicas removes the annotationOriginalReplicas from the workload.
func removeOriginalReplicas(workload Workload) {
	annotations := workload.GetAnnotations()
	delete(annotations, annotationOriginalReplicas)
	workload.SetAnnotations(annotations)
}

// GetIncludedKinds converts a comma-separated includedResources string list into a set of singular kinds.
func GetIncludedKinds(includedResources []string) (map[string]struct{}, error) {
	includedKinds := make(map[string]struct{}, len(includedResources))

	for _, resource := range includedResources {
		kind, err := GetKind(resource)
		if err != nil {
			return nil, err
		}

		includedKinds[kind] = struct{}{}
	}

	return includedKinds, nil
}
