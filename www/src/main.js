// Highlight the sidebar entry for the section currently in view, and close the
// mobile nav after a jump. No dependencies, no state outside the DOM.

const links = new Map();
for (const a of document.querySelectorAll('.nav a')) {
  links.set(a.getAttribute('href').slice(1), a);
}

const sections = [...links.keys()]
  .map((id) => document.getElementById(id))
  .filter(Boolean);

let active = null;
function setActive(id) {
  if (id === active) return;
  if (active && links.has(active)) links.get(active).classList.remove('active');
  if (id && links.has(id)) links.get(id).classList.add('active');
  active = id;
}

const observer = new IntersectionObserver(
  (entries) => {
    const visible = entries
      .filter((e) => e.isIntersecting)
      .sort((a, b) => a.boundingClientRect.top - b.boundingClientRect.top);
    if (visible.length) setActive(visible[0].target.id);
  },
  { rootMargin: '0px 0px -70% 0px', threshold: 0 },
);
for (const section of sections) observer.observe(section);

const toggle = document.getElementById('nav-toggle');
for (const a of links.values()) {
  a.addEventListener('click', () => {
    if (toggle) toggle.checked = false;
  });
}
