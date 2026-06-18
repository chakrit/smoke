// Generate www/src/index.html from the canonical markdown in docs/guides/.
// The markdown is the single source of truth; this only styles and frames it.

import { readFileSync, writeFileSync } from 'node:fs';
import { fileURLToPath } from 'node:url';
import { dirname, join } from 'node:path';
import MarkdownIt from 'markdown-it';
import anchor from 'markdown-it-anchor';
import hljs from 'highlight.js';

import { lifecycleSvg, mergeSvg } from './diagrams.mjs';

const here = dirname(fileURLToPath(import.meta.url));
const guidePath = join(here, '..', 'docs', 'guides', 'index.md');
const templatePath = join(here, 'template.html');
const outPath = join(here, 'src', 'index.html');

// Highlight at build time: the page ships pre-coloured hljs-* spans, so
// highlight.js never reaches the browser — only the token theme in styles.css.
function highlight(code, lang) {
  const language = hljs.getLanguage(lang) ? lang : null;
  const inner = language
    ? hljs.highlight(code, { language }).value
    : md.utils.escapeHtml(code);
  return `<pre><code class="hljs">${inner}</code></pre>`;
}

const md = new MarkdownIt({ html: true, linkify: true, typographer: false, highlight })
  .use(anchor, { tabIndex: false });

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

// Sidebar nav is the list of H2 sections, kept in sync with the headings.
const sections = [];
for (let i = 0; i < tokens.length; i++) {
  const open = tokens[i];
  if (open.type === 'heading_open' && open.tag === 'h2') {
    sections.push({ id: open.attrGet('id'), title: headingText(tokens[i + 1]) });
  }
}
const navHtml = sections
  .map((s) => `<li><a href="#${s.id}">${s.title}</a></li>`)
  .join('\n      ');

const contentHtml = md.renderer
  .render(tokens, md.options, {})
  .replace('<!--DIAGRAM:lifecycle-->', lifecycleSvg)
  .replace('<!--DIAGRAM:merge-->', mergeSvg);

const page = readFileSync(templatePath, 'utf8')
  .replace('{{NAV}}', navHtml)
  .replace('{{CONTENT}}', contentHtml);

writeFileSync(outPath, page);
console.log(`rendered ${sections.length} sections -> ${outPath}`);
