---
name: creative-studio
description: Run a design brief end to end as a studio would - research, positioning, creative direction, brand or design system, build, motion, QA, then a written case study. Use when the work is a whole project rather than a task: a new brand, a site or product from nothing, a portfolio or concept piece, a client brief, a pitch, or anything where the answer starts with understanding a business rather than opening an editor. Also use when asked to build one of the seven portfolio projects by name - AstraForge, Fixly, NexaGrid, TraceAI, Flowstate, NOMA, Orbit.
---

# Creative studio

One rule above all others:

> Never start designing before you can say, in one sentence, what this business or
> product is trying to communicate and to whom.

If you cannot, you are researching, not designing. Say so and keep researching.

## Phases

Each phase has an exit condition. Do not start the next until it is met. State which
phase you are in when you report progress.

### 1. Brief

Read `references/brief-intake.md`. Produce a written brief even if the user gave you
one - theirs is the request, yours is the understanding.

**Exit:** a one-sentence positioning statement, a named audience, and the single thing
this piece has to make someone feel or do.

### 2. Research

Read `references/research.md`. Understand the category, not just the client. For a
portfolio project, invent the client but research the real category.

`competitor-experience-audit` over the leading sites in the category. It reads them as
observable patterns - layout density, primary task, merchandising, register - and gives
you the experience bar this build has to meet or beat. Set the bar before building
against it, not after.

**Exit:** three named reference points with what is being taken from each, and a written
list of the category's clichés you are refusing.

### 3. Creative direction

Read `references/creative-direction.md`. A named direction, not a mood board.

`creative-direction` for the four axes - tone register, aesthetic philosophy, audience
relationship, sensory ambition. Its brief is the input every skill below reads, so it
decides the aesthetic rather than any one of them guessing.
`graphic-design-styles` to pick and commit to one style, and to say what it refuses.

**Exit:** a direction with a name, a one-line rationale, and a reason it beats the
obvious alternative.

### 4. System

- Brand-shaped work - `brand-identity` for the system, `logo-design` for the mark and
  its application specs, `brand-voice` before any copy is written, `art-direction` for
  the imagery brief. Record all of it in `BRAND.md` from
  `../design-router/templates/BRAND.md`, and `brand-style-guide` if it has to survive
  a handoff.
- One site or app - write `DESIGN.md` from
  `../vue-web-builder/templates/DESIGN.md`.
- Both, if the project has an identity and a product.

**Exit:** every token has a value. No section still in prose.

### 5. Structure

Information architecture and the narrative order before any visual work. For a site,
the section sequence and what each one has to land. For a product, the flows and states.

**Exit:** a written outline a stranger could follow, and every state named - empty,
loading, error, success, over-limit.

### 6. Build

Hand off:
- websites, web apps, dashboards - `vue-web-builder`
- posters, print, packaging art - `canvas-design`
- logos and identity assets - `ai-graphic-design`
- decks - `pitch-deck-visuals` for the order and the argument, then built-in `pptx`
  plus `theme-factory`
- generative pieces - `algorithmic-art`, seeded so it reproduces
- share cards for anything that gets linked - `og-image-design`
- email and campaigns - `email-design`
- app store listings - `app-store-screenshots`
- covers and editorial print - `book-cover-design`
- a one-page PDF for someone who will not click a link - `canvas-design`
- Webflow deliverables - the `webflow` skills

**Exit:** real content in place, responsive, every state built.

### 7. Motion

`animate` to build it, `review-animations` to judge it. One signature moment per
piece. See `../vue-web-builder/references/motion.md`.

**Exit:** reduced-motion honoured, nothing over 600ms without a reason.

### 8. Quality gates

Read `references/quality-gates.md`. This is the definition of done, not a suggestion.
Nothing ships on "it compiles".

**Exit:** every gate passed or explicitly waived in writing with a reason.

### 9. Case study

Read `references/case-study.md`. The project is not finished until it is written up.

**Exit:** a case study that shows the thinking, not just the screenshots.

## Honesty rules

- Concept projects are labelled as concept projects. "Self-initiated concept project"
  is credible; implying a client that does not exist is not.
- Invented companies get invented names, never a real company's name or marks.
- No claims you cannot support. This matters most for anything health, financial or
  scientific - a supplement brand can have beautiful packaging and say nothing about
  what the product does to a body.
- Generated imagery is yours to use but is not automatically safe as a trademark.
  For a logo, read `ai-graphic-design` on the IP side.
- If a phase went badly, the case study says so. A project with a recorded wrong turn
  reads as real work; one with none reads as a render.

## Stack

Web work is Vue 3, Bun, Vite, Tailwind v4. If a brief or another skill says Next.js,
React or Framer Motion, translate it - `../vue-web-builder/references/react-to-vue.md`.

Webflow is different. It is a client deliverable platform, not a substitute for the
build stack. Use the `webflow` skills when the deliverable is a Webflow site, and
`vue-web-builder` when it is code.

## Portfolio projects

Seven briefs live in `projects/`. Each covers several capabilities at once, which is
the point - seven strong pieces beat twenty thin ones.

| Brief | Proves |
| --- | --- |
| `projects/01-astraforge.md` | deep-tech web, brand, 3D, frontend |
| `projects/02-fixly.md` | AI product, marketplace UX, payments |
| `projects/03-nexagrid.md` | B2B rebrand, identity system, site |
| `projects/04-traceai.md` | complex software, data visualisation |
| `projects/05-flowstate.md` | mobile and desktop interaction design |
| `projects/06-noma.md` | consumer identity, packaging, art direction |
| `projects/07-orbit.md` | motion, product film, sound |

Build them in that order. Each one establishes a capability the next can assume.

Read the brief file, then run the nine phases against it. The brief is the input to
phase 1, not a replacement for it.

## References

- `references/brief-intake.md` - what to establish before anything else
- `references/research.md` - how to research a category you do not know
- `references/creative-direction.md` - arriving at a direction with a reason
- `references/quality-gates.md` - the definition of done
- `references/case-study.md` - writing the project up
