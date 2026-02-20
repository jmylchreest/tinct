import type {SidebarsConfig} from '@docusaurus/plugin-content-docs';

/**
 * Sidebar configuration for tinct documentation.
 * Organized by outcome (what users want to accomplish) rather than just reference.
 */
const sidebars: SidebarsConfig = {
  docs: [
    'intro',
    'installation',
    'telemetry',
    {
      type: 'category',
      label: 'Getting Started',
      items: [
        'quickstart/index',
        'quickstart/first-theme',
        'quickstart/workflows',
      ],
    },
    {
      type: 'category',
      label: 'Concepts',
      items: [
        'concepts/color-extraction',
        'concepts/material-design',
        'concepts/color-roles',
        'concepts/theme-detection',
      ],
    },
    {
      type: 'category',
      label: 'Commands',
      link: {
        type: 'doc',
        id: 'commands/index',
      },
      items: [
        'commands/generate',
        'commands/extract',
        'commands/plugins',
        'commands/files',
      ],
    },
    {
      type: 'category',
      label: 'Plugins',
      link: {
        type: 'doc',
        id: 'plugins/overview',
      },
      items: [
        {
          type: 'category',
          label: 'Input Plugins',
          link: {
            type: 'doc',
            id: 'plugins/input/index',
          },
          items: [
            'plugins/input/image',
            'plugins/input/google-genai',
            'plugins/input/openrouter',
            'plugins/input/remote-json',
            'plugins/input/remote-css',
            'plugins/input/file',
            'plugins/input/markdown',
          ],
        },
        {
          type: 'category',
          label: 'Output Plugins',
          link: {
            type: 'doc',
            id: 'plugins/output/index',
          },
          items: [
            {
              type: 'category',
              label: 'Terminals',
              items: [
                'plugins/output/terminals/alacritty',
                'plugins/output/terminals/ghostty',
                'plugins/output/terminals/kitty',
                'plugins/output/terminals/konsole',
                'plugins/output/terminals/ptyxis',
              ],
            },
            {
              type: 'category',
              label: 'Desktop Environments',
              items: [
                'plugins/output/desktop/gnome-shell',
                'plugins/output/desktop/kde-plasma',
                'plugins/output/desktop/gtk3',
                'plugins/output/desktop/gtk4',
                'plugins/output/desktop/libadwaita',
                'plugins/output/desktop/qt5',
                'plugins/output/desktop/qt6',
              ],
            },
            {
              type: 'category',
              label: 'Hyprland Ecosystem',
              items: [
                'plugins/output/hyprland/hyprland',
                'plugins/output/hyprland/hyprlock',
                'plugins/output/hyprland/hyprpaper',
              ],
            },
            {
              type: 'category',
              label: 'Bars & Launchers',
              items: [
                'plugins/output/bars-launchers/waybar',
                'plugins/output/bars-launchers/dunst',
                'plugins/output/bars-launchers/swayosd',
                'plugins/output/bars-launchers/fuzzel',
                'plugins/output/bars-launchers/walker',
                'plugins/output/bars-launchers/wofi',
              ],
            },
            {
              type: 'category',
              label: 'Editors & Multiplexers',
              items: [
                'plugins/output/editors/neovim',
                'plugins/output/editors/zellij',
              ],
            },
            {
              type: 'category',
              label: 'Special Purpose',
              items: [
                'plugins/output/special/markdown',
                'plugins/output/special/template',
                'plugins/output/special/histui',
              ],
            },
          ],
        },
      ],
    },
    {
      type: 'category',
      label: 'Templating',
      link: {
        type: 'doc',
        id: 'templating/index',
      },
      items: [
        'templating/functions',
        'templating/color-access',
        'templating/format-conversion',
        'templating/versioned',
      ],
    },
    {
      type: 'category',
      label: 'Plugin Development',
      link: {
        type: 'doc',
        id: 'plugin-development/index',
      },
      items: [
        'plugin-development/protocols',
        'plugin-development/creating',
        'plugin-development/hooks',
        'plugin-development/publishing',
      ],
    },
    {
      type: 'category',
      label: 'Changelog',
      link: {
        type: 'doc',
        id: 'changelog/index',
      },
      items: [
        'changelog/unreleased',
        'changelog/v0.1.22',
        'changelog/v0.1.21',
        'changelog/v0.1.20',
        'changelog/v0.1.18',
        'changelog/v0.1.17',
        'changelog/v0.1.16',
        'changelog/v0.1.15',
        'changelog/v0.1.14',
        'changelog/v0.1.11',
        'changelog/v0.1.10',
        'changelog/v0.1.8',
        'changelog/v0.1.5',
        'changelog/v0.1.0',
      ],
    },
  ],
};

export default sidebars;
