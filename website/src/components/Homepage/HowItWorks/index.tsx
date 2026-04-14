import React from "react";
import Heading from "@theme/Heading";
import useBaseUrl from "@docusaurus/useBaseUrl";
import { useColorMode } from "@docusaurus/theme-common";
import styles from "./styles.module.css";

function Terminal() {
  return (
    <div className={styles.terminal}>
      {/* Title bar */}
      <div className={styles.titleBar}>
        <span className={styles.dot} style={{ background: "#ff5f57" }} />
        <span className={styles.dot} style={{ background: "#febc2e" }} />
        <span className={styles.dot} style={{ background: "#28c840" }} />
        <span className={styles.fileName}>configmap.yaml</span>
      </div>
      {/* Code body */}
      <pre className={styles.code}>
        <code>
          <span className={styles.keyword}>apiVersion</span>
          <span className={styles.punct}>: </span>
          <span className={styles.value}>v1</span>
          {"\n"}
          <span className={styles.keyword}>kind</span>
          <span className={styles.punct}>: </span>
          <span className={styles.value}>ConfigMap</span>
          {"\n"}
          <span className={styles.keyword}>metadata</span>
          <span className={styles.punct}>:</span>
          {"\n"}
          {"  "}
          <span className={styles.keyword}>name</span>
          <span className={styles.punct}>: </span>
          <span className={styles.value}>kube-downscaler</span>
          {"\n"}
          {"  "}
          <span className={styles.keyword}>namespace</span>
          <span className={styles.punct}>: </span>
          <span className={styles.value}>kube-downscaler</span>
          {"\n"}
          <span className={styles.keyword}>data</span>
          <span className={styles.punct}>:</span>
          {"\n"}
          {"  "}
          <span className={styles.annotation}>DEFAULT_UPTIME</span>
          <span className={styles.punct}>: </span>
          <span className={styles.scheduleValue}>Mon-Fri 09:00-17:00 America/Los_Angeles</span>
          {"\n"}
          {"  "}
          <span className={styles.annotation}>EXCLUDE_NAMESPACES</span>
          <span className={styles.punct}>: </span>
          <span className={styles.value}>kube-system,cilium,kube-downscaler</span>
        </code>
      </pre>
    </div>
  );
}

export default function HowItWorks(): JSX.Element {
  const { colorMode } = useColorMode();
  const darkSrc  = useBaseUrl("/img/how-it-works-dark.png");
  const lightSrc = useBaseUrl("/img/how-it-works-light.png");
  const previewSrc = colorMode === "dark" ? darkSrc : lightSrc;
  return (
    <section className={styles.section}>
      <div className={styles.inner}>
        {/* Heading + description — always visible */}
        <div className={`${styles.textBlock} animate-fade-down animate-once animate-delay-0`}>
          <Heading as="h2" className={styles.headline}>
            How It Works
          </Heading>
          <p className={styles.body}>
            The most common way to use GoKubeDownscaler is to set a global
            schedule. The controller continuously read the
            desired configuration and scales down workloads when needed. Once the
            downscaling window ends, the controller brings the workload back
            at its original state.
          </p>
        </div>

        {/* Terminal — always visible */}
        <div className={`${styles.terminalBlock} animate-fade-down animate-once animate-delay-300`}>
          <Terminal />
        </div>

        {/* Image — hidden on small screens or when image is missing */}
        <div className={`${styles.visualBlock} animate-fade-down animate-once animate-delay-500`}>
          <img
            src={previewSrc}
            alt="GoKubeDownscaler in action"
            className={styles.previewImage}
            onError={(e) => {
              (e.currentTarget.parentElement as HTMLElement).style.display = "none";
            }}
          />
        </div>
      </div>
    </section>
  );
}
