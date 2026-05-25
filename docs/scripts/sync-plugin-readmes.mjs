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
  const landings = await regenerateLandingTables(readmes);

  console.error(
    `\ntinct-sync-plugin-readmes: ${written} pages written, ${landings} landing table(s) refreshed, ${warningCount} warning(s)`
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

// regenerateLandingTables rewrites the plugin tables in
// docs/plugins/{input,output}/index.md between the AUTO-PLUGIN-TABLE
// markers. Curated content above and below the markers is preserved.
async function regenerateLandingTables(readmes) {
  const outputIndex = path.join(OUTPUT_BASE, 'output', 'index.md');
  const inputIndex = path.join(OUTPUT_BASE, 'input', 'index.md');
  let count = 0;
  if (await replaceMarkerBlock(outputIndex, renderOutputTables(readmes))) count++;
  if (await replaceMarkerBlock(inputIndex, renderInputTable(readmes))) count++;
  return count;
}

// replaceMarkerBlock rewrites the content between the `{/* BEGIN
// AUTO-PLUGIN-TABLE */}` and `{/* END AUTO-PLUGIN-TABLE */}` markers.
// Returns true if the file was updated, false if the markers are
// missing (in which case we leave the file alone and warn).
async function replaceMarkerBlock(file, body) {
  let content;
  try {
    content = await fs.readFile(file, 'utf8');
  } catch (e) {
    warn(`${file}: cannot read for landing-table regen: ${e.message}`);
    return false;
  }
  const beginMarker = '{/* BEGIN AUTO-PLUGIN-TABLE */}';
  const endMarker = '{/* END AUTO-PLUGIN-TABLE */}';
  const beginIdx = content.indexOf(beginMarker);
  const endIdx = content.indexOf(endMarker);
  if (beginIdx === -1 || endIdx === -1 || endIdx < beginIdx) {
    warn(`${file}: AUTO-PLUGIN-TABLE markers missing or out of order; skipping`);
    return false;
  }
  const before = content.slice(0, beginIdx + beginMarker.length);
  const after = content.slice(endIdx);
  const rewritten = `${before}\n\n${body.trimEnd()}\n\n${after}`;
  if (rewritten === content) return false;
  await fs.writeFile(file, rewritten);
  return true;
}

// Output landing: tables grouped by plugin.category. Category order
// and labels are explicit (not alphabetical) so the rendering reads
// the way humans expect.
const OUTPUT_CATEGORIES = [
  ['terminals', 'Terminals'],
  ['desktop', 'Desktop environments'],
  ['hyprland', 'Hyprland ecosystem'],
  ['bars-launchers', 'Bars and launchers'],
  ['editors', 'Editors and multiplexers'],
  ['special', 'Special purpose'],
];

function renderOutputTables(readmes) {
  const outputs = readmes
    .filter((r) => r.frontmatter.plugin?.type === 'output')
    .sort(byFrontmatterPosition);
  const sections = [];
  for (const [cat, label] of OUTPUT_CATEGORIES) {
    const items = outputs.filter((r) => r.frontmatter.plugin.category === cat);
    if (items.length === 0) continue;
    const rows = items
      .map((r) => {
        const p = r.frontmatter.plugin;
        const link = `./${cat}/${p.name}.md`;
        const desc = stripRedundantLead(
          firstParagraphSummary(r.content) || p.app || ''
        );
        return `| [${p.name}](${link}) | ${desc} |`;
      })
      .join('\n');
    sections.push(`## ${label}\n\n| Plugin | Description |\n|--------|-------------|\n${rows}`);
  }
  return sections.join('\n\n');
}

function renderInputTable(readmes) {
  const inputs = readmes
    .filter((r) => r.frontmatter.plugin?.type === 'input')
    .sort(byFrontmatterPosition);
  const rows = inputs
    .map((r) => {
      const p = r.frontmatter.plugin;
      const link = `./${p.name}.md`;
      const desc = p.description ?? firstParagraphSummary(r.content);
      const sourceType = p.source_type ?? '';
      const reqs = [];
      if (p.requires_network) reqs.push('network');
      if (Array.isArray(p.requires_credentials) && p.requires_credentials.length > 0) {
        reqs.push(`creds: ${p.requires_credentials.join(', ')}`);
      }
      const notes = reqs.join('; ');
      return `| [${p.name}](${link}) | ${desc} | ${sourceType} | ${notes} |`;
    })
    .join('\n');
  return `| Plugin | Description | Source type | Requires |\n|--------|-------------|-------------|----------|\n${rows}`;
}

// firstParagraphSummary extracts the first prose paragraph from a
// README body and reduces it to one cell-friendly line: markdown
// links/code stripped, whitespace normalised, truncated at the first
// sentence-ish boundary (≤140 chars). Falls back to empty string if
// the body has no usable prose (e.g. all headings + code blocks).
function firstParagraphSummary(body) {
  if (!body) return '';
  const lines = body.split('\n');
  let inFence = false;
  const collected = [];
  for (const raw of lines) {
    const line = raw.trim();
    if (line.startsWith('```')) {
      inFence = !inFence;
      continue;
    }
    if (inFence) continue;
    if (line.startsWith('#') || line.startsWith('---')) continue;
    if (line === '') {
      if (collected.length > 0) break;
      continue;
    }
    collected.push(line);
  }
  if (collected.length === 0) return '';

  let text = collected.join(' ');
  // [text](url) → text
  text = text.replace(/\[([^\]]+)\]\([^)]+\)/g, '$1');
  // **strong** → strong, *em* → em
  text = text.replace(/\*\*([^*]+)\*\*/g, '$1').replace(/\*([^*]+)\*/g, '$1');
  // `code` → code (keeps the literal token readable in a table cell)
  text = text.replace(/`([^`]+)`/g, '$1');
  // Collapse whitespace.
  text = text.replace(/\s+/g, ' ').trim();

  // Trim to the first sentence if it's reasonably short; otherwise
  // hard-truncate at 140. We walk candidate sentence ends in order and
  // skip ones whose last token is a common abbreviation (`i.e.`,
  // `e.g.`, `etc.`, etc.) so they don't masquerade as sentence ends
  // when the abbreviation is followed by a capitalised noun.
  const firstSentence = extractFirstSentence(text);
  if (firstSentence && firstSentence.length <= 140) return firstSentence;
  if (text.length > 140) return text.slice(0, 137).trimEnd() + '…';
  return text;
}

// stripRedundantLead removes leading verbs from a description that
// are visually redundant in a table where "output plugin" is already
// implied. The PLUGIN-README-STANDARD asks for "What the plugin
// generates, for which app", so almost every output README opens with
// "Generates …" — fine in prose, repetitive in a column of 30+ rows.
function stripRedundantLead(text) {
  const match = text.match(/^Generates\s+(?:an?\s+)?(.+)/);
  if (!match) return text;
  const rest = match[1];
  return rest.charAt(0).toUpperCase() + rest.slice(1);
}

// Common abbreviations whose trailing period should NOT count as a
// sentence end. Match is case-insensitive and bounded by word edges.
const ABBREVIATIONS = new Set([
  'i.e.', 'e.g.', 'etc.', 'cf.', 'vs.',
  'mr.', 'mrs.', 'ms.', 'dr.',
  'inc.', 'ltd.', 'co.', 'st.', 'no.',
  'fig.', 'eq.', 'ref.', 'ch.', 'sec.',
]);

// extractFirstSentence returns the longest leading substring of `text`
// ending in [.!?] followed by whitespace+uppercase (a real sentence
// boundary), skipping candidates whose last word is in the
// abbreviations set. Returns null if no qualifying boundary is found.
function extractFirstSentence(text) {
  let offset = 0;
  while (offset < text.length) {
    const rest = text.slice(offset);
    const match = rest.match(/[.!?](?=\s+[A-Z]|\s*$)/);
    if (!match) return null;
    const endIdx = offset + match.index + 1;
    const candidate = text.slice(0, endIdx);
    const lastWord = candidate.match(/\S+$/)?.[0]?.toLowerCase();
    if (lastWord && ABBREVIATIONS.has(lastWord)) {
      offset = endIdx;
      continue;
    }
    return candidate;
  }
  return null;
}

function byFrontmatterPosition(a, b) {
  const pa = a.frontmatter.sidebar_position ?? 99;
  const pb = b.frontmatter.sidebar_position ?? 99;
  return pa - pb;
}

main().catch((e) => {
  console.error(e);
  // Exit 0 even on a thrown error so we never gate the docs build on
  // this script. The error message above is the user's signal.
  process.exit(0);
});
