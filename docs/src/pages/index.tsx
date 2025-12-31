import type {ReactNode} from 'react';
import Link from '@docusaurus/Link';
import useDocusaurusContext from '@docusaurus/useDocusaurusContext';
import Layout from '@theme/Layout';

import styles from './index.module.css';

function HeroSection() {
  return (
    <section className={styles.hero}>
      <div className={styles.heroInner}>
        <h1 className={styles.title}>
          t<span className={styles.titleAccent}>inct</span>
        </h1>
        <p className={styles.pronunciation}>/tiŋ(k)t/</p>
        <p className={styles.tagline}>
          An extensible colour palette generator and theme manager for unified
          theming across your entire environment.
        </p>

        <div className={styles.buttons}>
          <Link className={styles.primaryBtn} to="/docs">
            Get Started
          </Link>
          <Link className={styles.secondaryBtn} to="https://github.com/jmylchreest/tinct">
            View on GitHub
          </Link>
        </div>

        <div className={styles.codePreview}>
          <div className={styles.codeHeader}>
            <span className={`${styles.codeDot} ${styles.codeDotRed}`}></span>
            <span className={`${styles.codeDot} ${styles.codeDotYellow}`}></span>
            <span className={`${styles.codeDot} ${styles.codeDotGreen}`}></span>
            <span className={styles.codeTitle}>terminal</span>
          </div>
          <div className={styles.codeContent}>
            <div className={styles.codeLine}>
              <span className={styles.codePrompt}>$</span>
              <span className={styles.codeCommand}>tinct generate -i image -p ~/wallpaper.jpg -o all</span>
            </div>
            <div className={styles.codeOutput}>
              <div>Extracting palette from image...</div>
              <div>Generated 49 semantic colour roles</div>
              <div>Applied to: hyprland, kitty, waybar, dunst...</div>
            </div>
            <div className={styles.codeLine} style={{marginTop: '1rem'}}>
              <span className={styles.codePrompt}>$</span>
              <span className={styles.codeCommand}>tinct plugins list</span>
            </div>
            <div className={styles.codeOutput}>
              <div>25+ output plugins available</div>
            </div>
          </div>
        </div>
      </div>
    </section>
  );
}

type FeatureItem = {
  icon: string;
  title: string;
  description: string;
};

const features: FeatureItem[] = [
  {
    icon: '\u{1F3A8}',
    title: 'Multiple Input Sources',
    description: 'Extract colours from images, generate with AI (Google Gemini, OpenRouter), fetch remote themes (JSON/CSS), or specify manually.',
  },
  {
    icon: '\u{1F4D0}',
    title: 'Material Design 3',
    description: 'Automatic semantic colour role assignment with 49+ roles. WCAG contrast compliance and intelligent light/dark theme detection.',
  },
  {
    icon: '\u{1F5A5}',
    title: '25+ Applications',
    description: 'Apply consistent themes across terminals, desktop environments, window managers, status bars, editors, and more.',
  },
  {
    icon: '\u{1F4A1}',
    title: 'External Device Support',
    description: 'Control LED strips (WLED), smart lights, and RGB peripherals (OpenRGB). Ambient edge/corner colour extraction for monitor sync.',
  },
  {
    icon: '\u{1F9E9}',
    title: 'Plugin Architecture',
    description: 'Extend with custom plugins in any language. JSON-stdio protocol for simple scripts, go-plugin for high-performance solutions.',
  },
  {
    icon: '\u{1F4BE}',
    title: 'Theme Portability',
    description: 'Save complete themes to markdown files with embedded wallpapers. Share, version control, and restore your exact setup anywhere.',
  },
];

function FeaturesSection() {
  return (
    <section className={styles.features}>
      <div className={styles.featuresInner}>
        <h2 className={styles.featuresTitle}>Unified Theming for Your Desktop</h2>
        <div className={styles.featuresGrid}>
          {features.map((feature, idx) => (
            <div key={idx} className={styles.featureCard}>
              <div className={styles.featureIcon}>{feature.icon}</div>
              <h3 className={styles.featureTitle}>{feature.title}</h3>
              <p className={styles.featureDesc}>{feature.description}</p>
            </div>
          ))}
        </div>
      </div>
    </section>
  );
}

function InstallSection() {
  return (
    <section className={styles.install}>
      <div className={styles.installInner}>
        <h2 className={styles.installTitle}>Quick Install (Arch Linux)</h2>
        <div className={styles.installCode}>
          <span>
            <span className={styles.installPrompt}>$ </span>
            yay -S tinct-bin
          </span>
        </div>
        <p style={{marginTop: '1rem', color: '#666', fontSize: '0.875rem'}}>
          See <Link to="/docs/installation">Installation Guide</Link> for Go install, source builds, and other methods
        </p>
      </div>
    </section>
  );
}

export default function Home(): ReactNode {
  const {siteConfig} = useDocusaurusContext();
  return (
    <Layout
      title="Home"
      description={siteConfig.tagline}>
      <HeroSection />
      <FeaturesSection />
      <InstallSection />
    </Layout>
  );
}
