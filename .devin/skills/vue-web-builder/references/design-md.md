# DESIGN.md

`SKILL.md` says how to work. `DESIGN.md` says what this specific project looks like.
It lives at the project root, next to `CLAUDE.md`, and it is the thing that keeps
page five consistent with page one.

Write it before building screens. Get it agreed. Then treat it as the source of truth -
if a build decision contradicts it, either the build is wrong or `DESIGN.md` needs
updating first.

## Sources for a starting point

- https://getdesign.md - a catalogue of DESIGN.md files analysing real sites.
  Pick one close to the target and adapt it.
- `bergside/awesome-design-skills` - 67 named styles, each with a SKILL.md and
  DESIGN.md. Pick one per project. Do not let two styles mix.
- `Leonxlnx/taste-skill` `image-to-code-skill` - when there is a screenshot or a
  reference site, extract the system from it rather than eyeballing it.
- `billhector/design-skills` - extracts a design system from a live URL and audits
  an existing project, including Tailwind v4 theme output and contrast checks.

## Filling it in

Start from `templates/DESIGN.md`. Every value becomes a `@theme` token. If a section
is still prose by the time you start building components, it is not finished.
