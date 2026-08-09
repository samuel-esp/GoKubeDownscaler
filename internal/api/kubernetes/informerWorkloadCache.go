package kubernetes

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/caas-team/gokubedownscaler/internal/pkg/scalable"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/cache"
)

type informerResource struct {
	gvr      schema.GroupVersionResource
	resource string
}

// These GVRs mirror the resource types accepted by the downscaler. Informers
// are created only for resources selected by the user.
var informerResources = map[string]informerResource{
	"deployments":              {schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "deployments"}, "deployment"},
	"statefulsets":             {schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "statefulsets"}, "statefulset"},
	"daemonsets":               {schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "daemonsets"}, "daemonset"},
	"cronjobs":                 {schema.GroupVersionResource{Group: "batch", Version: "v1", Resource: "cronjobs"}, "cronjob"},
	"jobs":                     {schema.GroupVersionResource{Group: "batch", Version: "v1", Resource: "jobs"}, "job"},
	"poddisruptionbudgets":     {schema.GroupVersionResource{Group: "policy", Version: "v1", Resource: "poddisruptionbudgets"}, "poddisruptionbudget"},
	"horizontalpodautoscalers": {schema.GroupVersionResource{Group: "autoscaling", Version: "v2", Resource: "horizontalpodautoscalers"}, "horizontalpodautoscaler"},
	"scaledobjects":            {schema.GroupVersionResource{Group: "keda.sh", Version: "v1alpha1", Resource: "scaledobjects"}, "scaledobject"},
	"rollouts":                 {schema.GroupVersionResource{Group: "argoproj.io", Version: "v1alpha1", Resource: "rollouts"}, "rollout"},
	"stacks":                   {schema.GroupVersionResource{Group: "zalando.org", Version: "v1", Resource: "stacks"}, "stack"},
	"prometheuses":             {schema.GroupVersionResource{Group: "monitoring.coreos.com", Version: "v1", Resource: "prometheuses"}, "prometheus"},
	"autoscalingrunnersets":    {schema.GroupVersionResource{Group: "actions.github.com", Version: "v1alpha1", Resource: "autoscalingrunnersets"}, "autoscalingrunnerset"},
	"services":                 {schema.GroupVersionResource{Version: "v1", Resource: "services"}, "service"},
	"awsnlbservices":           {schema.GroupVersionResource{Version: "v1", Resource: "services"}, "service"},
	"awselbservices":           {schema.GroupVersionResource{Version: "v1", Resource: "services"}, "service"},
	"ingresses":                {schema.GroupVersionResource{Group: "networking.k8s.io", Version: "v1", Resource: "ingresses"}, "ingress"},
	"gateways":                 {schema.GroupVersionResource{Group: "gateway.networking.k8s.io", Version: "v1", Resource: "gateways"}, "gateway"},
	"postgresqls":              {schema.GroupVersionResource{Group: "acid.zalan.do", Version: "v1", Resource: "postgresqls"}, "postgresql"},
	"kafkaconnects":            {schema.GroupVersionResource{Group: "kafka.strimzi.io", Version: "v1", Resource: "kafkaconnects"}, "kafkaconnect"},
	"kafkamirrormaker2s":       {schema.GroupVersionResource{Group: "kafka.strimzi.io", Version: "v1", Resource: "kafkamirrormaker2s"}, "kafkamirrormaker2"},
	"kafkabridges":             {schema.GroupVersionResource{Group: "kafka.strimzi.io", Version: "v1", Resource: "kafkabridges"}, "kafkabridge"},
}

type informerWorkloadCache struct {
	informers map[string]cache.SharedIndexInformer
	resources map[string]informerResource
	mu        sync.RWMutex
}

func newNamespaceInformer(client dynamic.Interface) (cache.SharedIndexInformer, error) {
	if client == nil {
		return nil, errors.New("dynamic Kubernetes client is nil")
	}

	resourceInterface := client.Resource(schema.GroupVersionResource{Version: "v1", Resource: "namespaces"}).Namespace(metav1.NamespaceAll)
	listWatch := &cache.ListWatch{
		ListWithContextFunc: func(ctx context.Context, options metav1.ListOptions) (runtime.Object, error) {
			return resourceInterface.List(ctx, options)
		},
		WatchFuncWithContext: func(ctx context.Context, options metav1.ListOptions) (watch.Interface, error) {
			return resourceInterface.Watch(ctx, options)
		},
	}

	informer := cache.NewSharedIndexInformer(
		listWatch,
		&unstructured.Unstructured{},
		10*time.Minute,
		cache.Indexers{},
	)
	if err := addInformerDebugLogging("namespaces", informer); err != nil {
		return nil, fmt.Errorf("failed to add namespace informer event handler: %w", err)
	}

	slog.Debug("created Kubernetes informer", "resource", "namespaces")

	return informer, nil
}

func newInformerWorkloadCache(client dynamic.Interface, resourceTypes []string) (*informerWorkloadCache, error) {
	if client == nil {
		return nil, errors.New("dynamic Kubernetes client is nil")
	}

	c := &informerWorkloadCache{
		informers: make(map[string]cache.SharedIndexInformer),
		resources: make(map[string]informerResource),
	}

	for _, resourceType := range resourceTypes {
		resourceType = strings.ToLower(resourceType)

		resource, ok := informerResources[resourceType]
		if !ok {
			return nil, fmt.Errorf("no informer mapping for resource %q", resourceType)
		}

		if _, exists := c.informers[resourceType]; exists {
			continue
		}

		resourceInterface := client.Resource(resource.gvr).Namespace(metav1.NamespaceAll)
		listWatch := &cache.ListWatch{
			ListWithContextFunc: func(ctx context.Context, options metav1.ListOptions) (runtime.Object, error) {
				return resourceInterface.List(ctx, options)
			},
			WatchFuncWithContext: func(ctx context.Context, options metav1.ListOptions) (watch.Interface, error) {
				return resourceInterface.Watch(ctx, options)
			},
		}

		informer := cache.NewSharedIndexInformer(
			listWatch,
			&unstructured.Unstructured{},
			10*time.Minute,
			cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc},
		)
		if err := addInformerDebugLogging(resourceType, informer); err != nil {
			return nil, fmt.Errorf("failed to add informer event handler for %q: %w", resourceType, err)
		}

		c.informers[resourceType] = informer
		c.resources[resourceType] = resource
		slog.Debug("created Kubernetes informer", "resource", resourceType, "gvr", resource.gvr.String())
	}

	return c, nil
}

func addInformerDebugLogging(resource string, informer cache.SharedIndexInformer) error {
	_, err := informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj any) {
			logInformerEvent("added", resource, obj, false)
		},
		UpdateFunc: func(_, newObj any) {
			logInformerEvent("updated", resource, newObj, false)
		},
		DeleteFunc: func(obj any) {
			logInformerEvent("deleted", resource, obj, true)
		},
	})

	return err
}

func logInformerEvent(event, resource string, object any, deletion bool) {
	key, err := cache.MetaNamespaceKeyFunc(object)
	if deletion {
		key, err = cache.DeletionHandlingMetaNamespaceKeyFunc(object)
	}

	if err != nil {
		slog.Debug("Kubernetes informer event", "event", event, "resource", resource, "error", err)
		return
	}

	slog.Debug("Kubernetes informer event", "event", event, "resource", resource, "key", key)
}

func startAndWaitForInformers(ctx context.Context, informers ...cache.SharedIndexInformer) (context.CancelFunc, error) {
	runCtx, cancel := context.WithCancel(ctx)

	slog.Debug("starting Kubernetes informers", "count", len(informers))

	for index, informer := range informers {
		slog.Debug("starting Kubernetes informer", "index", index)

		go informer.Run(runCtx.Done())
	}

	deadline := time.NewTimer(15 * time.Second)
	defer deadline.Stop()

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		allSynced := true

		for _, informer := range informers {
			if !informer.HasSynced() {
				allSynced = false
				break
			}
		}

		if allSynced {
			slog.Debug("Kubernetes informers synchronized", "count", len(informers))
			return cancel, nil
		}

		select {
		case <-deadline.C:
			cancel()
			slog.Error("Kubernetes informers failed to synchronize", "count", len(informers), "timeout", "15s")

			return nil, errors.New("informer caches failed to sync within 15 seconds")
		case <-ctx.Done():
			cancel()
			slog.Debug("Kubernetes informer startup canceled", "count", len(informers), "error", ctx.Err())

			return nil, ctx.Err()
		case <-ticker.C:
		}
	}
}

func (c *informerWorkloadCache) GetWorkloads(namespaces, resourceTypes []string) ([]scalable.Workload, error) {
	namespaceSet := make(map[string]struct{}, len(namespaces))
	for _, namespace := range namespaces {
		namespaceSet[namespace] = struct{}{}
	}

	allNamespaces := namespaces == nil
	workloads := make([]scalable.Workload, 0)

	slog.Debug("reading workloads from informer cache", "resources", resourceTypes, "namespaces", namespaces)

	c.mu.RLock()
	defer c.mu.RUnlock()

	for _, resourceType := range resourceTypes {
		resourceType = strings.ToLower(resourceType)

		informer, ok := c.informers[resourceType]
		if !ok {
			return nil, fmt.Errorf("resource %q is not cached", resourceType)
		}

		resource := c.resources[resourceType]
		resourceCount := 0

		for _, obj := range informer.GetStore().List() {
			cachedObject, ok := obj.(*unstructured.Unstructured)
			if !ok {
				return nil, fmt.Errorf("cached resource %q has unexpected type %T", resourceType, obj)
			}

			if !allNamespaces {
				if _, included := namespaceSet[cachedObject.GetNamespace()]; !included {
					continue
				}
			}

			if resourceType == "awselbservices" || resourceType == "awsnlbservices" {
				value := strings.ToLower(cachedObject.GetAnnotations()[scalable.AWSLoadBalancerAnnotation])

				isNLB := value == "nlb"
				if (resourceType == "awsnlbservices") != isNLB {
					continue
				}
			}

			cachedCopy := cachedObject.DeepCopy()
			if err := setCachedType(cachedCopy, resource.gvr, resource.resource); err != nil {
				return nil, err
			}

			raw, err := json.Marshal(cachedCopy.Object)
			if err != nil {
				return nil, fmt.Errorf("failed to encode cached %s: %w", resourceType, err)
			}

			workload, err := scalable.ParseWorkloadFromRawObject(resource.resource, raw)
			if err != nil {
				return nil, fmt.Errorf("failed to parse cached %s: %w", resourceType, err)
			}

			workloads = append(workloads, workload)
			resourceCount++
		}

		slog.Debug("read workloads from informer cache", "resource", resourceType, "count", resourceCount)
	}

	slog.Debug("finished reading workloads from informer cache", "count", len(workloads))

	return workloads, nil
}

func setCachedType(obj *unstructured.Unstructured, gvr schema.GroupVersionResource, resource string) error {
	if obj == nil {
		return fmt.Errorf("cached %s is nil", resource)
	}

	if obj.GetAPIVersion() == "" {
		obj.SetAPIVersion(gvr.GroupVersion().String())
	}

	if obj.GetKind() == "" {
		return fmt.Errorf("cached %s has no kind", resource)
	}

	return nil
}
