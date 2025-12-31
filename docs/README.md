# Tinct Documentation

This directory contains the documentation website for tinct, built with [Docusaurus](https://docusaurus.io/).

## Development

### Prerequisites

- Node.js >= 20.0
- npm or yarn

### Local Development

```bash
cd docs
npm install
npm run start
```

This starts a local development server at http://localhost:3000/ with hot reloading.

### Build

```bash
npm run build
```

Generates static content in the `build` directory.

### Deployment

Documentation is automatically deployed to GitHub Pages when changes are pushed to the `main` branch.

## Documentation Guidelines

### Writing Style

- Use sentence case for headings
- Include code examples for all CLI commands
- Specify language for code blocks (`bash`, `toml`, `css`, `go`, `json`)
- Use admonitions sparingly (`note`, `warning`, `tip`)
- Keep pages focused; split large topics into multiple pages
- Link to related documentation where applicable

### File Organization

- `/docs/` - Markdown documentation content
- `/src/` - React components and custom CSS
- `/static/` - Static assets (images, fonts)

### Versioning

To create a new version:

```bash
npm run docusaurus docs:version X.Y.Z
```

This snapshots the current documentation into `versioned_docs/version-X.Y.Z/`.

## Theme Configuration

The theme uses a warm amber/orange accent color scheme. To change the accent color, edit the CSS variables at the top of `src/css/custom.css`:

```css
/* Current: Amber/Orange */
--ifm-color-primary: #f59e0b;

/* Alternative: Teal (matching histui) */
/* --ifm-color-primary: #14b8a6; */
```
