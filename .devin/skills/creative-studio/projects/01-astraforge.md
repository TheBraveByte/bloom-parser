# 01 - AstraForge Materials

Self-initiated concept project. Build first - it establishes the visual and technical
web capability the later projects assume.

**Proves:** deep-tech web design, brand basics, 3D or technical visualisation,
production frontend.

## The fiction

AstraForge develops lightweight thermal and structural materials for aerospace, electric
vehicles, robotics and high-performance computing. It sells to engineers who specify
materials, not to consumers.

Invented company. No real company's name, marks or claims.

## Positioning to arrive at

Do not use this - derive your own in phase 1. It is here as a calibration of specificity:

> For the design engineer choosing between a known material and a better one, AstraForge
> is the supplier that publishes its test data, because the decision is career risk
> before it is cost.

That is the shape. The buyer's fear is being the person who specified the part that
failed. Design against that.

## Category clichés to refuse

Dark navy, hexagon patterns, circuit-board textures, glowing blue wireframe globes,
generic laboratory stock photography, "innovative solutions", a carousel of industry
logos with no context.

## What has to be in it

- Cinematic opening that is about the material, not about space
- Positioning that a specifying engineer would not find insulting
- The technology, explained to someone technical
- Measurable performance, with real-looking numbers and units, compared to a baseline
- One memorable interactive technical visualisation - a cross-section, a thermal
  gradient, a load response
- Applications by industry
- Process and research
- Evidence - certifications, test standards, published data
- Contact that reads as a supplier, not a startup

## Deliverables

- Running Vue 3 site, responsive, production quality
- `BRAND.md` and `DESIGN.md`
- Generated material imagery, consistent across the set
- Case study

## Notes on execution

Stack is Vue 3, Vite, Bun, Tailwind v4. If a reference or another skill suggests
Next.js, translate it.

3D only if it genuinely explains something. A rotating abstract shape is decoration and
will read as such next to the rest of the piece. If you use `three`, lazy-load it and
check the frame rate on a phone. `vue-web-builder/references/motion-libraries.md`.

Typography and whitespace carry this one. Reach for scientific publishing and
industrial documentation before reaching for other tech sites.

Numbers need units and a baseline. "40% lighter" is marketing; "1.8 g/cm3 against
2.7 for 6061 aluminium" is engineering. The second one is the design decision.

## Skills for this project

Load these, not the whole library. Anything not listed is out of scope for this build.

| Phase | Skill | Why |
| --- | --- | --- |
| 2 research | `design-extractor` | pull real tokens off scientific publishing and instrumentation references |
| 3 direction | `graphic-design-styles` | Swiss with editorial. Refuse the futuristic default. |
| 3 direction | `impeccable` | critique the candidate direction, do not generate with it |
| 4 system | `brand-identity`, `brand-style-guide` | minimal identity, enough to hold the site |
| 4 system | `theme-factory` | palette if it is not settled after direction |
| 5 structure | `information-architecture` | section order and what each has to land |
| 6 build | `vue-web-builder` | the whole build |
| 6 build | `data-visualization` + built-in `dataviz` | the performance comparison charts |
| 7 motion | `animate`, then `review-animations` | the interactive visualisation |
| 8 gates | `design-auditor`, `webapp-testing`, `ux-audit` | contrast, browser, structured pass |
| 8 gates | `taste-skill` | did it go generic |

Higgsfield for material imagery. `bin/doctor.sh` before gate 3.

**Not used:** `ui-ux-pro-max` (marketing-site styles, wrong register here),
`canvas-design`, `apple-design`, the Webflow plugin.

## Gates that matter most here

Content (gate 3) and performance (gate 7). A deep-tech site with placeholder numbers or
a hero that shifts on load fails at the thing it is claiming to be.
