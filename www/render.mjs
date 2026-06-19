// Generate www/src/index.html from the canonical markdown in docs/guides/.
// The markdown is the single source of truth; this only styles and frames it.

import { readFileSync, writeFileSync } from 'node:fs';
import { fileURLToPath } from 'node:url';
import { dirname, join } from 'node:path';
import MarkdownIt from 'markdown-it';
import anchor from 'markdown-it-anchor';
import hljs from 'highlight.js';

import { lifecycleSvg } from './diagrams.mjs';

const here = dirname(fileURLToPath(import.meta.url));
const guidePath = join(here, '..', 'docs', 'guides', 'index.md');
const templatePath = join(here, 'template.html');
const outPath = join(here, 'src', 'index.html');

// Highlight at build time: the page ships pre-coloured hljs-* spans, so
// highlight.js never reaches the browser — only the token theme in styles.css.
function highlight(code, lang) {
  const resolved = lang === 'jsonl' ? 'json' : lang;   // each JSONL line is JSON
  const language = hljs.getLanguage(resolved) ? resolved : null;
  const inner = language
    ? hljs.highlight(code, { language }).value
    : md.utils.escapeHtml(code);
  return `<pre><code class="hljs">${inner}</code></pre>`;
}

// GitHub-compatible heading slugs. The guide is canonical markdown that also
// renders on GitHub, so site anchors must match GitHub's (strip punctuation),
// not markdown-it-anchor's default (percent-encode it) — otherwise an intra-doc
// link like [Advanced](#advanced-cue-json-and-jsonl-specs) resolves on only one.
function slugify(text) {
  return text
    .trim()
    .toLowerCase()
    .replace(/[^\w\s-]/g, '')
    .replace(/\s+/g, '-');
}

const md = new MarkdownIt({ html: true, linkify: true, typographer: false, highlight })
  .use(anchor, { tabIndex: false, slugify });

const source = readFileSync(guidePath, 'utf8');
const tokens = md.parse(source, {});

// Plain text of a heading's inline token — read its text children rather than
// .content, which markdown-it leaves empty once the inline has been tokenized.
function headingText(inline) {
  return (inline.children ?? [])
    .filter((c) => c.type === 'text' || c.type === 'code_inline')
    .map((c) => c.content)
    .join('');
}

// Sidebar nav mirrors the H2 headings. A designated H2 renders as a labelled,
// non-clickable section whose H3 subsections become its nav children, instead
// of a plain link.
const NAV_SECTIONS = { 'advanced-spec-formats': 'Advanced' };

const nav = [];
let section = null;
for (let i = 0; i < tokens.length; i++) {
  const open = tokens[i];
  if (open.type !== 'heading_open') continue;

  const id = open.attrGet('id');
  const title = headingText(tokens[i + 1]);

  if (open.tag === 'h2' && id in NAV_SECTIONS) {
    section = { label: NAV_SECTIONS[id], children: [] };
    nav.push({ section });
  } else if (open.tag === 'h2') {
    section = null;
    nav.push({ item: { id, title } });
  } else if (open.tag === 'h3' && section) {
    section.children.push({ id, title: title.replace(/ \(.*\)$/, '') });
  }
}

function navLink({ id, title }) {
  return `<li><a href="#${id}">${title}</a></li>`;
}

function navEntry(entry) {
  if (entry.item) return navLink(entry.item);
  const subs = entry.section.children.map(navLink).join('\n          ');
  return `<li class="nav-section">
        <span class="nav-section-label">${entry.section.label}</span>
        <ul class="nav-sub">
          ${subs}
        </ul>
      </li>`;
}

const navHtml = nav.map(navEntry).join('\n      ');

const contentHtml = md.renderer
  .render(tokens, md.options, {})
  .replace('<!--DIAGRAM:lifecycle-->', lifecycleSvg);

const page = readFileSync(templatePath, 'utf8')
  .replace('{{NAV}}', navHtml)
  .replace('{{CONTENT}}', contentHtml);

writeFileSync(outPath, page);
console.log(`rendered ${nav.length} nav entries -> ${outPath}`);
