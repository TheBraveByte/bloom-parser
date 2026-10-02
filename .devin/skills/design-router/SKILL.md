---
name: design-router
description: Entry point for any design request. Use FIRST whenever the user asks to design, redesign, brand, lay out, mock up, illustrate or visualise anything - a website, app, logo, brand identity, poster, billboard, flyer, social post, packaging, deck, dashboard, wireframe, product shot, ad creative, generative art or print piece. Works out what kind of design job it is, which of the installed skills to use and in what order, and what the deliverable file should be. Also use when the user is not sure which skill applies, or says "design something" without naming a medium.
---

# Design router

Work out the job, then hand off. Do not design from this file.

## 1. Name the job

Ask only what you cannot infer. Usually you need three things:

- **Medium** - screen, print, social, deck, physical object, moving image
- **Deliverable** - a live site, a `.png`, a `.pdf`, a `.pptx`, a component, a spec
- **Brand** - is there an existing one? If the project has `BRAND.md`, read it first.
  If not and the job is brand-shaped, write one from `templates/BRAND.md`.

Two things to settle before routing anywhere:

- If the piece has to stand against an established field - a storefront, a SaaS site,
  a designer portfolio - run `competitor-experience-audit` first. It reads the leading
  sites as observable patterns and gives you the bar to beat. Cheaper now than after
  the build.
- If the aesthetic is not decided, `creative-direction` decides it. Four axes - tone
  register, aesthetic philosophy, audience relationship, sensory ambition - and its
  brief is what every skill downstream reads. Do not guess the style and call it a
  direction.

## 2. Route

### Screen - websites and web apps

`vue-web-builder` owns this. Vue 3, Bun, Vite, Tailwind v4. Never React.

Support skills, in order of use:
1. `impeccable` or `ui-ux-pro-max` - direction, palette, type pairing
2. `taste-skill` - reject the generic version before building it
3. `image-to-code` or `design-extractor` - if there is a screenshot or a reference URL
4. build with `vue-web-builder`
5. `animate`, then `review-animations`
6. `design-auditor` and `webapp-testing` - contrast, then the browser

### Screen - dashboards, product UI, app screens

Same as above, but `ui-ux-pro-max` over `impeccable` - it carries the industry
patterns. `reka-ui` or `shadcn-vue` as the component base.

### Print and static - poster, billboard, flyer, cover, packaging art

`canvas-design` owns this. It outputs `.png` and `.pdf`, which is what print needs.

1. `canvas-design` - it builds a design philosophy first, then expresses it
2. `theme-factory` if a palette is not settled
3. Generate imagery with Higgsfield - `vue-web-builder/references/assets.md`
4. Check the output at final size, not on screen at 40%

Billboard specifics: one idea, legible at distance, text no smaller than a tenth of
the height, and design at final aspect ratio from the start. `canvas-design` does not
know your substrate - state bleed and safe margins up front.

### Brand identity - logo, palette, type system, voice

1. `creative-direction` if the aesthetic is not settled. Nothing below works without it.
2. Write `BRAND.md` from `templates/BRAND.md`. This is the deliverable that everything
   else derives from.
3. `brand-identity` - the system: logo, colour, type, imagery, iconography, motion
4. `logo-design` - the mark on its own. Several architectures - wordmark, lockup,
   symbol, monogram - each with the application specs: 16px favicon, single colour,
   reverse, signage. `ai-graphic-design` generates and vectorises it, and covers IP.
5. `brand-voice` - how it writes. Before the copy, not after it.
6. `art-direction` - the photography and imagery brief. Lighting, grade, crop, fixed
   across the whole set so the pictures read as one brand.
7. `theme-factory` - palette and type as applicable tokens
8. `brand-style-guide`, then `canvas-design` for the guidelines document as a PDF
9. Higgsfield for product and environment mockups

Do not use the installed `brand-guidelines` skill from Anthropic if you see it -
that applies *Anthropic's* brand. This repo does not install it for that reason.

### Social - carousel, banner, ad creative, thumbnail

`banner-design` for the layout rules. `canvas-design` for the output.
Higgsfield for the imagery. Export at the platform's exact pixel size.

### Share cards - the image a link renders

`og-image-design`. Every link pasted into LinkedIn, X, Slack, WhatsApp or iMessage
renders a card. With no `og:image` it renders nothing, which is a blank space where a
piece of design should be.

1200x630, legible at the size a phone shows it, and the words in the image rather than
only in the meta tag. For a site, build it as a route and screenshot it so it stays in
sync; `canvas-design` for a one-off. Then paste the real URL somewhere and look at it.

### Decks and documents

`pptx` and `docx` are already built into Claude Code - use those, not a skill.
`pitch-deck-visuals` for an investor or pitch deck: the slide order and what each slide
has to prove, settled before any layout. `slides` for an HTML deck with charts.
`theme-factory` for the visual theme. `canvas-design` for any full-bleed slide art, and
for the one-page PDF you can send to someone who will not click a link.

### Email - newsletter, campaign, transactional

`email-design`. One column under 600px, inline styles, and a version that still reads
with images off. The subject line and preheader are part of the design, not something
written afterwards.

### App store listing

`app-store-screenshots`. Exact iOS and Play sizes, the gallery order, and the device
frame. The first two shots do all the work - the rest are read by almost nobody.

### Covers - book, report, editorial

`book-cover-design`. Genre convention first, then the thumbnail test: it has to hold at
100px before it is allowed to be good at full size. `canvas-design` for the output file.

### Generative and algorithmic art

`algorithmic-art` - p5.js, flow fields, particle systems, seeded randomness.
For particles inside a website, use `vue-web-builder/references/motion-libraries.md`
instead - that is tsParticles, not p5.

### Wireframes and low fidelity

`wireframe-sketch` if the optional group is installed. Otherwise `prototype` to put
several versions side by side and pick one.

### Motion and video

`animate` for UI motion. `motion-frames` for storyboard frames.
Higgsfield for generated video. Never animate a hero video that is decorative and
autoplaying without a poster frame and a pause control.

## 3. Rules that apply to every medium

- Direction before execution. A named philosophy, palette and type pairing before
  the first pixel or the first line.
- One idea per piece. If you cannot say what it is about in a sentence, it is not
  designed yet.
- Real content. Placeholder copy makes a layout look right when it is not.
- Look at the output at final size, in final context, before saying it is done.
- Generated imagery is licensed to you but not automatically safe as a trademark.
  For a logo, `ai-graphic-design-skill` covers the IP side - read it.

## 4. Deliverable by medium

| Medium | File | Notes |
| --- | --- | --- |
| website, app | running project | plus `DESIGN.md` at the root |
| poster, billboard, packaging | `.pdf` | CMYK, bleed, fonts outlined or embedded |
| social, banner, ad | `.png` | exact platform pixel size, 2x for retina |
| logo | `.svg` plus `.png` | vectorised, not an upscaled raster |
| brand identity | `BRAND.md` plus a `.pdf` guideline | tokens usable in code |
| deck | `.pptx` | |
| generative art | `.html` plus `.js` | seeded, so it reproduces |
| share card | `.png` at 1200x630 | plus `og:image` and `twitter:image` tags |
| investor deck | `.pptx` | one idea per slide |
| email | HTML, inline styles | 600px, reads with images off |
| app store listing | `.png` per required size | first two shots carry it |
| book or report cover | `.pdf` | checked at 100px as well as full size |

## Templates

- `templates/BRAND.md` - the brand spec everything else derives from
