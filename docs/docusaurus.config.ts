import {themes as prismThemes} from 'prism-react-renderer';
import type {Config} from '@docusaurus/types';
import type * as Preset from '@docusaurus/preset-classic';

// Check if we have any released versions
const versions: string[] = require('./versions.json');
const hasReleasedVersions = versions.length > 0;

const config: Config = {
  title: 'tinct',
  tagline: 'An extensible colour palette generator and theme manager for unified theming across your entire environment',
  favicon: 'img/favicon.ico',

  // GitHub Pages deployment
  url: 'https://jmylchreest.github.io',
  baseUrl: '/tinct/',
  organizationName: 'jmylchreest',
  projectName: 'tinct',
  deploymentBranch: 'gh-pages',
  trailingSlash: false,

  onBrokenLinks: 'throw',

  markdown: {
    hooks: {
      onBrokenMarkdownLinks: 'warn',
    },
  },

  i18n: {
    defaultLocale: 'en',
    locales: ['en'],
  },

  presets: [
    [
      'classic',
      {
        docs: {
          sidebarPath: './sidebars.ts',
          editUrl: 'https://github.com/jmylchreest/tinct/tree/main/docs/',
          // Versioning configuration - only enabled when versions exist
          // When versions.json is empty, current version is served at /docs/
          // When versions exist, released versions are at /docs/ and current at /docs/next/
          includeCurrentVersion: true,
          ...(hasReleasedVersions ? {
            versions: {
              current: {
                label: 'main',
                path: 'next',
                banner: 'unreleased',
              },
            },
            lastVersion: versions[0],
          } : {}),
        },
        blog: false, // Disable blog
        theme: {
          customCss: './src/css/custom.css',
        },
      } satisfies Preset.Options,
    ],
  ],

  // Local search plugin (no external service dependency)
  plugins: [
    [
      '@cmfcmf/docusaurus-search-local',
      {
        indexDocs: true,
        indexBlog: false,
        indexPages: true,
        language: 'en',
        maxSearchResults: 8,
      },
    ],
  ],

  themeConfig: {
    // Color mode configuration
    colorMode: {
      defaultMode: 'dark',
      disableSwitch: false,
      respectPrefersColorScheme: true,
    },

    // Navigation bar
    navbar: {
      title: 'tinct',
      logo: {
        alt: 'tinct Logo',
        src: 'img/logo.svg',
      },
      items: [
        {
          type: 'docSidebar',
          sidebarId: 'docs',
          position: 'left',
          label: 'Documentation',
        },
        // Only show version dropdown when versions exist
        ...(hasReleasedVersions ? [{
          type: 'docsVersionDropdown' as const,
          position: 'right' as const,
          dropdownActiveClassDisabled: true,
        }] : []),
        {
          href: 'https://github.com/jmylchreest/tinct',
          label: 'GitHub',
          position: 'right',
        },
      ],
    },

    // Footer
    footer: {
      style: 'dark',
      links: [
        {
          title: 'Documentation',
          items: [
            {label: 'Getting Started', to: '/docs/quickstart/'},
            {label: 'Commands', to: '/docs/commands/'},
            {label: 'Plugins', to: '/docs/plugins/input/'},
          ],
        },
        {
          title: 'Community',
          items: [
            {label: 'GitHub', href: 'https://github.com/jmylchreest/tinct'},
            {label: 'Issues', href: 'https://github.com/jmylchreest/tinct/issues'},
          ],
        },
        {
          title: 'Related Projects',
          items: [
            {label: 'histui', href: 'https://jmylchreest.github.io/histui/'},
            {label: 'tinct-plugins', href: 'https://github.com/jmylchreest/tinct-plugins'},
          ],
        },
      ],
      copyright: `Copyright ${new Date().getFullYear()} tinct. Built with Docusaurus.`,
    },

    // Prism syntax highlighting
    prism: {
      theme: prismThemes.github,
      darkTheme: prismThemes.dracula,
      additionalLanguages: ['bash', 'toml', 'css', 'go', 'json'],
    },
  } satisfies Preset.ThemeConfig,
};

export default config;
