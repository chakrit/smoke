// Inline SVG diagrams injected into the rendered guide at <!--DIAGRAM:*--> markers.
// Colours follow the visualise design-system ramps (light-mode picks: 50 fill,
// 600 stroke, 800 text). Flat fills only — no gradients or shadows.

const FONT = 'font-family: ui-sans-serif, system-ui, -apple-system, sans-serif';

const ARROW = `
  <defs>
    <marker id="arrow" viewBox="0 0 10 10" refX="8" refY="5"
      markerWidth="7" markerHeight="7" orient="auto-start-reverse">
      <path d="M2 1L8 5L2 9" fill="none" stroke="context-stroke"
        stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/>
    </marker>
  </defs>`;

function box(x, y, w, h, fill, stroke) {
  return `<rect x="${x}" y="${y}" width="${w}" height="${h}" rx="8"
    fill="${fill}" stroke="${stroke}" stroke-width="1"/>`;
}

const C = {
  amber: { fill: '#FAEEDA', stroke: '#854F0B', text: '#633806' },
  green: { fill: '#EAF3DE', stroke: '#3B6D11', text: '#27500A' },
  red: { fill: '#FCEBEB', stroke: '#A32D2D', text: '#791F1F' },
  gray: { fill: '#F1EFE8', stroke: '#888780', text: '#444441' },
  teal: { fill: '#E1F5EE', stroke: '#0F6E56', text: '#085041' },
};
const LABEL = '#5F5E5A';

export const lifecycleSvg = `<figure class="diagram">
<svg width="100%" viewBox="0 0 680 190" role="img"
  aria-label="Lifecycle: NEW becomes UNCHANGED after review and commit; UNCHANGED drifts to CHANGED; CHANGED returns to UNCHANGED after review and commit."
  style="${FONT}">
  ${ARROW}
  ${box(46, 44, 148, 56, C.amber.fill, C.amber.stroke)}
  ${box(266, 44, 148, 56, C.green.fill, C.green.stroke)}
  ${box(486, 44, 148, 56, C.red.fill, C.red.stroke)}

  <g text-anchor="middle">
    <text x="120" y="66" font-size="14" font-weight="500" fill="${C.amber.text}">NEW</text>
    <text x="120" y="84" font-size="12" fill="${C.amber.text}">exit 3</text>
    <text x="340" y="66" font-size="14" font-weight="500" fill="${C.green.text}">UNCHANGED</text>
    <text x="340" y="84" font-size="12" fill="${C.green.text}">exit 0</text>
    <text x="560" y="66" font-size="14" font-weight="500" fill="${C.red.text}">CHANGED</text>
    <text x="560" y="84" font-size="12" fill="${C.red.text}">exit 1</text>
  </g>

  <line x1="200" y1="72" x2="260" y2="72" stroke="${LABEL}" stroke-width="1" marker-end="url(#arrow)"/>
  <line x1="420" y1="72" x2="480" y2="72" stroke="${LABEL}" stroke-width="1" marker-end="url(#arrow)"/>
  <path d="M524 100 C 524 152, 356 152, 356 102" fill="none" stroke="${LABEL}"
    stroke-width="1" marker-end="url(#arrow)"/>

  <g text-anchor="middle" font-size="12" fill="${LABEL}">
    <text x="230" y="34">review &#183; commit</text>
    <text x="450" y="34">output drifts</text>
    <text x="440" y="148">review &#183; commit</text>
  </g>
</svg>
</figure>`;

function row(yc, leftFill, leftStroke, rightFill, rightStroke, leftText, rightText, leftLabel, rightLabel, arrowColor, arrowLabel) {
  const top = yc - 17;
  return `
  ${box(46, top, 150, 34, leftFill, leftStroke)}
  ${box(484, top, 150, 34, rightFill, rightStroke)}
  <text x="121" y="${yc + 4}" text-anchor="middle" font-size="14" font-weight="500" fill="${leftText}">${leftLabel}</text>
  <text x="559" y="${yc + 4}" text-anchor="middle" font-size="14" font-weight="500" fill="${rightText}">${rightLabel}</text>
  <line x1="200" y1="${yc}" x2="480" y2="${yc}" stroke="${arrowColor}" stroke-width="1" marker-end="url(#arrow)"/>
  <text x="340" y="${yc - 8}" text-anchor="middle" font-size="12" fill="${arrowColor}">${arrowLabel}</text>`;
}

export const mergeSvg = `<figure class="diagram">
<svg width="100%" viewBox="0 0 680 210" role="img"
  aria-label="Partial commit merges only the filtered test onto the lock: tests A and C are kept, test B is replaced by its new output."
  style="${FONT}">
  ${ARROW}
  <text x="340" y="24" text-anchor="middle" font-size="12" fill="${C.teal.text}"
    style="font-family: ui-monospace, monospace">smoke --include B --commit</text>
  <text x="121" y="50" text-anchor="middle" font-size="12" fill="${LABEL}">existing lock</text>
  <text x="559" y="50" text-anchor="middle" font-size="12" fill="${LABEL}">after the commit</text>

  ${row(82, C.gray.fill, C.gray.stroke, C.gray.fill, C.gray.stroke, C.gray.text, C.gray.text, 'A', 'A', C.gray.stroke, 'kept')}
  ${row(124, C.gray.fill, C.gray.stroke, C.teal.fill, C.teal.stroke, C.gray.text, C.teal.text, 'B', 'B′', C.teal.stroke, 'replaced')}
  ${row(166, C.gray.fill, C.gray.stroke, C.gray.fill, C.gray.stroke, C.gray.text, C.gray.text, 'C', 'C', C.gray.stroke, 'kept')}
</svg>
</figure>`;
