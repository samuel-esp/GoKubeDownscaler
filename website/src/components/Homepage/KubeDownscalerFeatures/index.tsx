import React, { ReactNode } from "react";
import clsx from "clsx";
import Heading from "@theme/Heading";
import styles from "./styles.module.css";

/* ── Inline SVG icons ── */

function ClockSvg(props: React.ComponentProps<"svg">) {
  return (
    <svg {...props} viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth={1.5} strokeLinecap="round" strokeLinejoin="round">
      <circle cx="12" cy="12" r="10" />
      <polyline points="12 6 12 12 16 14" />
    </svg>
  );
}

function MoneySvg(props: React.ComponentProps<"svg">) {
  return (
    <svg {...props} viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth={1.5} strokeLinecap="round" strokeLinejoin="round">
      <rect x="2" y="7" width="20" height="14" rx="2" ry="2" />
      <path d="M16 3H8a2 2 0 0 0-2 2v2h12V5a2 2 0 0 0-2-2z" />
      <line x1="12" y1="12" x2="12" y2="16" />
      <line x1="10" y1="14" x2="14" y2="14" />
    </svg>
  );
}

function LayersSvg(props: React.ComponentProps<"svg">) {
  return (
    <svg {...props} viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth={1.5} strokeLinecap="round" strokeLinejoin="round">
      <polygon points="12 2 2 7 12 12 22 7 12 2" />
      <polyline points="2 17 12 22 22 17" />
      <polyline points="2 12 12 17 22 12" />
    </svg>
  );
}

function SliderSvg(props: React.ComponentProps<"svg">) {
  return (
    <svg {...props} viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth={1.5} strokeLinecap="round" strokeLinejoin="round">
      <line x1="4" y1="21" x2="4" y2="14" />
      <line x1="4" y1="10" x2="4" y2="3" />
      <line x1="12" y1="21" x2="12" y2="12" />
      <line x1="12" y1="8" x2="12" y2="3" />
      <line x1="20" y1="21" x2="20" y2="16" />
      <line x1="20" y1="12" x2="20" y2="3" />
      <line x1="1" y1="14" x2="7" y2="14" />
      <line x1="9" y1="8" x2="15" y2="8" />
      <line x1="17" y1="16" x2="23" y2="16" />
    </svg>
  );
}

function ShieldSvg(props: React.ComponentProps<"svg">) {
  return (
    <svg {...props} viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth={1.5} strokeLinecap="round" strokeLinejoin="round">
      <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z" />
    </svg>
  );
}

function PuzzleSvg(props: React.ComponentProps<"svg">) {
  return (
    <svg {...props} viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth={1.5} strokeLinecap="round" strokeLinejoin="round">
      <path d="M20.59 13.41l-7.17 7.17a2 2 0 0 1-2.83 0L2 12V2h10l8.59 8.59a2 2 0 0 1 0 2.82z" />
      <line x1="7" y1="7" x2="7.01" y2="7" />
    </svg>
  );
}

/* ── Feature list ── */

type FeatureItem = {
  title: string;
  Svg: React.ComponentType<React.ComponentProps<"svg">>;
  description: ReactNode;
};

const FeatureList: FeatureItem[] = [
  {
    title: "Works Everywhere",
        Svg: PuzzleSvg,
    description: (
      <>
          Built for any Kubernetes distribution across AWS, GCP, Azure, and on-premises environments.
          Fully supports all Kubernetes resources and popular CRDs like KEDA, Prometheus, and Argo.
      </>
  ),
  },
  {
    title: "Enterprise Grade Scheduling",
    Svg: SliderSvg,
    description: (
      <>
          Control scheduling globally through CLI flags and environment variables, or via annotations
          at namespace or workload level. Supports flexible scheduling for teams across timezones in multi-tenant clusters.
      </>
    ),
  },
  {
      title: "Time Driven Scaling",
          Svg: ClockSvg,
      description: (
      <>
          Define scaling windows via RFC3339 timestamps, weekly schedules, or always/never rules.
          Reduce costs with time as a native scaling dimension aligned to real infrastructure usage.
      </>
  ),
  },
//{
//    title: "Time Native Scaling",
//        Svg: ClockSvg,
//    description: (
//    <>
//        Define downscale and upscale windows using RFC3339 timestamps,
//        recurring weekly schedules, or always and never rules.
//        Apply them per workload, namespace, or globally.
//    </>
//),
//},
];

//{
//    title: "Reduce Cloud Costs",
//        Svg: MoneySvg,
//    description: (
//    <>
//        Stop paying for idle replicas at night and on weekends. Scale workloads
//        to zero during off-hours and bring them back automatically when needed.
//    </>
//),
//},
//{
//    title: "Broad Resource Support",
//        Svg: LayersSvg,
//    description: (
//    <>
//        Works with Deployments, StatefulSets, DaemonSets, CronJobs, HPAs, PDBs,
//        Argo Rollouts, KEDA ScaledObjects, Zalando Stacks and more.
//    </>
//),
//},
//{
//    title: "Ecosystem Integration",
//        Svg: PuzzleSvg,
//    description: (
//    <>
//        Native support for Prometheus monitoring, KEDA event-driven scaling,
//        Argo Rollouts, GitHub Actions runners, and Zalando Stacksets.
//    </>
//),
//},
//{
//    title: "Time-based Scheduling",
//        Svg: ClockSvg,
//    description: (
//    <>
//        Define downscale and upscale windows using RFC3339 timestamps,
//        recurring weekly schedules, or always and never rules.
//        Apply them per workload, namespace, or globally.
//    </>
//),
//},
//{
//    title: "Safe Operations",
//        Svg: ShieldSvg,
//    description: (
//    <>
//        Original state is preserved so workloads always return to the
//        right state at the right time. Grace periods protect freshly deployed workloads from
//        premature downscaling.
//    </>
//),
//},

/* ── Sub-components ── */

function Feature({ title, Svg, description }: FeatureItem) {
  return (
    <div className={clsx("col col--4", styles.col)}>
      <div className="text--center">
        <div className={styles.iconWrap}>
          <Svg className={styles.featureSvg} role="img" aria-label={title} />
        </div>
      </div>
      <div className="text--center padding-horiz--md">
        <Heading as="h3">{title}</Heading>
        <p>{description}</p>
      </div>
    </div>
  );
}

/* ── Section ── */

export default function KubeDownscalerFeatures(): ReactNode {
  return (
    <section className={styles.features}>
      <div className="container">
        <div className="row">
          {FeatureList.map((props, idx) => (
            <Feature key={idx} {...props} />
          ))}
        </div>
      </div>
    </section>
  );
}
