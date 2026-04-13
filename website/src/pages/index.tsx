import useDocusaurusContext from "@docusaurus/useDocusaurusContext";
import Layout from "@theme/Layout";
import { SupportedResources } from "@site/src/components/Homepage/SupportedResources";
import ProjectDescription from "@site/src/components/Homepage/ProjectDescription";
import KubeDownscalerFeatures from "@site/src/components/Homepage/KubeDownscalerFeatures";
import { Button } from "../components/Basic/Button";
import * as KubedownscalerSVG from "@site/static/img/kubedownscaler.svg";
import Heading from "@theme/Heading";
import Head from "@docusaurus/Head";

function HomepageHeader() {
  const { siteConfig } = useDocusaurusContext();
  return (
    <div className="relative overflow-x-hidden overflow-y-visible">
      <div className="transform bg-magenta -skew-y-6 xl:hidden h-full w-full absolute top-0 origin-top-left" />
      <header className="select-none text-white bg-magenta items-center flex pt-10 pb-24 px-8 overflow-hidden relative text-center">
        <div className="px-4 w-full flex flex-col items-center justify-center gap-6">
          {/* Logo */}
          <KubedownscalerSVG.default className="animate-fade-down h-28 sm:h-36 md:h-44" />
          {/* Name */}
          <Heading
            as="h1"
            className="animate-fade-down text-[clamp(1.75rem,6vw,3.5rem)] font-bold m-0"
          >
            {siteConfig.title}
          </Heading>
          {/* Subtitle */}
          <p className="animate-fade-down text-lg sm:text-xl md:text-2xl lg:text-3xl max-w-4xl m-0">
            Saving Cloud Costs By Scaling Workloads Down After Hours
          </p>
          {/* CTA buttons */}
          <div className="animate-fade-down flex justify-center gap-3 flex-col sm:flex-row">
            <Button name="Get Started" to="/guides/getting-started" className="w-52" />
            <Button name="Documentation" to="/docs" className="w-52" />
          </div>
        </div>
      </header>
    </div>
  );
}

export default function Home(): JSX.Element {
  const { siteConfig } = useDocusaurusContext();
  return (
    <Layout title={siteConfig.title} description={siteConfig.tagline}>
      <Head>
        <title>
          {siteConfig.title}: {siteConfig.tagline}
        </title>
      </Head>
      <HomepageHeader />
      <main>
        <ProjectDescription />
        <KubeDownscalerFeatures />
        <SupportedResources />
      </main>
    </Layout>
  );
}
