#!/usr/bin/env node
// sync-plugin-readmes.mjs — generate Docusaurus per-plugin pages from
// the canonical in-tree READMEs.
//
// Reads:
//   - internal/plugin/output/<name>/README.md      (built-in output)
//   - internal/plugin/input/<dir>/README.md        (built-in input)
//   - contrib/plugins/output/<name>/README.md      (in-repo external)
//
// Validates frontmatter:
//   - Mandatory fields per plugin.type (output/input)
//   - YAML parse must succeed
//   - Failure → skip + warn (per the rule in PLUGIN-README-STANDARD.md)
//
// Writes to:
//   - docs/plugins/output/<category>/<name>.md     (type=output)
//   - docs/plugins/input/<name>.md                 (type=input, flat)
//
// Cross-links rewritten:
//   `[x](../<name>/README.md)` and similar → the right Docusaurus path
//   based on the target plugin's resolved location.
//
// Always exits 0. Warn-only, by design — this runs as a prebuild hook
// and should not gate the docs build on a single stale README.

import {glob} from 'glob';
import matter from 'gray-matter';
import fs from 'fs/promises';
import path from 'path';
import {fileURLToPath} from 'url';

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const REPO_ROOT = path.resolve(__dirname, '..', '..');
const DOCS_ROOT = path.resolve(__dirname, '..');
const OUTPUT_BASE = path.join(DOCS_ROOT, 'docs', 'plugins');

const SOURCES = [
  'internal/plugin/output/*/README.md',
  'internal/plugin/input/*/README.md',
  'contrib/plugins/output/*/README.md',
  'contrib/plugins/input/*/README.md',
];

// Mandatory fields per plugin.type. Aligned with the "Docs site build
// behaviour" section of PLUGIN-README-STANDARD.md. A README missing any
// of these is skipped with a warning.
const MANDATORY = {
  output: ['type', 'name', 'category', 'source', 'app'],
  input: ['type', 'name', 'source', 'source_type', 'description'],
};

let warningCount = 0;
function warn(msg) {
  warningCount++;
  console.error(`WARN ${msg}`);
}

async function main() {
  const readmes = await collectReadmes();
  const nameIndex = buildNameIndex(readmes);
  await cleanGeneratedDirs();
  const written = await writeAll(readmes, nameIndex);

  console.error(
    `\ntinct-sync-plugin-readmes: ${written} pages written, ${warningCount} warning(s)`
  );
}

// collectReadmes walks every source glob, parses frontmatter, and
// returns the records that survive validation. Skipped READMEs are
// warned about and dropped.
async function collectReadmes() {
  const results = [];
  for (const pattern of SOURCES) {
    const files = await glob(pattern, {cwd: REPO_ROOT, absolute: true});
    for (const file of files) {
      const rel = path.relative(REPO_ROOT, file);
      const raw = await fs.readFile(file, 'utf8');

      let parsed;
      try {
        parsed = matter(raw);
      } catch (e) {
        warn(`${rel}: YAML frontmatter parse error: ${e.message}`);
        continue;
      }

      const plugin = parsed.data?.plugin;
      if (!plugin) {
        warn(`${rel}: missing 'plugin:' block in frontmatter`);
        continue;
      }

      const type = plugin.type;
      if (!type || !MANDATORY[type]) {
        warn(
          `${rel}: invalid or missing plugin.type=${JSON.stringify(type)}; expected 'output' or 'input'`
        );
        continue;
      }

      const missing = MANDATORY[type].filter((f) => {
        const v = plugin[f];
        return v === undefined || v === null || v === '';
      });
      if (missing.length > 0) {
        warn(`${rel}: missing mandatory field(s): ${missing.join(', ')}`);
        continue;
      }

      // Record the source directory basename so the cross-link rewriter
       // can resolve links written against the directory name (e.g.
       // `../googlegenai/`) as well as the plugin name
       // (`../google-genai/`). Input plugin dirs don't always match
       // their plugin name.
      const sourceDir = path.basename(path.dirname(file));
      results.push({
        sourceFile: rel,
        sourceDir,
        frontmatter: parsed.data,
        content: parsed.content,
      });
    }
  }
  return results;
}

// buildNameIndex maps both plugin name AND source directory basename →
// resolved location. Cross-links may be spelled either way; we resolve
// from whichever key matches first.
function buildNameIndex(readmes) {
  const index = new Map();
  for (const r of readmes) {
    const p = r.frontmatter.plugin;
    const target = {type: p.type, category: p.category ?? null, canonicalName: p.name};
    index.set(p.name, target);
    if (r.sourceDir && r.sourceDir !== p.name) {
      index.set(r.sourceDir, target);
    }
  }
  return index;
}

// cleanGeneratedDirs removes generated pages so stale ones (from a
// previous run with a plugin that's since been removed or renamed)
// don't linger. We preserve the per-type/category index.md landing
// pages — those are hand-written and listed in sidebars.ts directly.
async function cleanGeneratedDirs() {
  const preservedNames = new Set(['index.md', 'index.mdx']);
  const targets = [
    path.join(OUTPUT_BASE, 'input'),
    path.join(OUTPUT_BASE, 'output'),
  ];
  for (const dir of targets) {
    try {
      await rmGeneratedRecursive(dir, preservedNames);
    } catch (e) {
      if (e.code !== 'ENOENT') throw e;
    }
  }
}

async function rmGeneratedRecursive(dir, preserved) {
  const entries = await fs.readdir(dir, {withFileTypes: true});
  for (const entry of entries) {
    const full = path.join(dir, entry.name);
    if (entry.isDirectory()) {
      await rmGeneratedRecursive(full, preserved);
      // Remove the directory if now empty (categories with no plugins).
      const remaining = await fs.readdir(full);
      if (remaining.length === 0) {
        await fs.rmdir(full);
      }
      continue;
    }
    if (preserved.has(entry.name)) continue;
    if (entry.name.endsWith('.md') || entry.name.endsWith('.mdx')) {
      await fs.unlink(full);
    }
  }
}

// writeAll renders each surviving README into its Docusaurus location,
// rewriting cross-links along the way.
async function writeAll(readmes, nameIndex) {
  let written = 0;
  for (const r of readmes) {
    const p = r.frontmatter.plugin;
    const outPath =
      p.type === 'output'
        ? path.join(OUTPUT_BASE, 'output', p.category, `${p.name}.md`)
        : path.join(OUTPUT_BASE, 'input', `${p.name}.md`);

    const rewritten = rewriteCrossLinks(r.content, p, nameIndex);
    const stringified = matter.stringify(rewritten, r.frontmatter);

    await fs.mkdir(path.dirname(outPath), {recursive: true});
    await fs.writeFile(outPath, stringified);
    written++;
  }
  return written;
}

// rewriteCrossLinks finds Markdown links pointing at another plugin's
// README and rewrites them to the Docusaurus-resolved path of that
// plugin.
//
// Two patterns accepted (handles existing in-tree README conventions):
//   - `](anything/<key>/README.md)` — explicit README link
//   - `](anything/<key>/)` or `](anything/<key>)` — directory link
//
// `<key>` is matched against the name-or-directory index, so links
// written against either the plugin name (`../google-genai/`) or the
// source directory (`../googlegenai/`) both resolve.
function rewriteCrossLinks(content, selfRaw, nameIndex) {
  // Normalise self.category to null so it compares cleanly against the
  // target.category we recorded in the name index (which is also null
  // for input plugins).
  const self = {
    name: selfRaw.name,
    type: selfRaw.type,
    category: selfRaw.category ?? null,
  };

  const replacer = (full, _prefix, key, suffix) => {
    const target = nameIndex.get(key);
    if (!target) {
      // Unknown — leave alone for onBrokenLinks to surface.
      return full;
    }
    const targetName = target.canonicalName;
    // Self-reference → same dir.
    if (targetName === self.name) {
      return `](./${targetName}.md)`;
    }
    // Same type and same category → relative sibling.
    if (target.type === self.type && target.category === self.category) {
      return `](./${targetName}.md)`;
    }
    // Both output, different categories → cross-category relative.
    if (target.type === 'output' && self.type === 'output') {
      return `](../${target.category}/${targetName}.md)`;
    }
    // Crossing input ↔ output → use absolute Docusaurus URL.
    if (target.type === 'input') {
      return `](/docs/plugins/input/${targetName})`;
    }
    return `](/docs/plugins/output/${target.category}/${targetName})`;
  };

  return content
    // `../<key>/README.md` form.
    .replace(/\]\(([^)]*\/)?([\w-]+)\/README\.md\)/g, replacer)
    // `../<key>/` or `../<key>` form. Prefix must start with `.` or
    // `..` and end with `/`, optionally including intermediate segments.
    // The trailing slash on the key is optional. README.md links are
    // handled by the previous .replace, so they won't reach here.
    .replace(
      /\]\((\.\.?(?:\/[^/)]+)*\/)([\w-]+)\/?\)/g,
      (full, prefix, key) => replacer(full, prefix, key, '')
    );
}

main().catch((e) => {
  console.error(e);
  // Exit 0 even on a thrown error so we never gate the docs build on
  // this script. The error message above is the user's signal.
  process.exit(0);
});
