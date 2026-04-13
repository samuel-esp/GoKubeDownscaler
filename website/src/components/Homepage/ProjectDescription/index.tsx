import React from "react";
import Heading from "@theme/Heading";
import styles from "./styles.module.css";

export default function ProjectDescription(): JSX.Element {
  return (
    <section className={styles.section}>
      <div className={styles.inner}>
        <Heading as="h2" className={styles.headline}>
            Smart Kubernetes Downscaling Powered By Schedules
        </Heading>
        <p className={styles.body}>
            GoKubeDownscaler acts as a horizontal autoscaler that reduces cloud costs by
            keeping workloads running only when they are needed, and automatically scaling
            them down at night, on weekends, or during planned company holidays when they
            are not in use.
        </p>
      </div>
    </section>
  );
}
