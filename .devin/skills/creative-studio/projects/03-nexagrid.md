# 03 - NexaGrid

Self-initiated concept project, framed as a rebrand.

**Proves:** brand strategy, identity system, guidelines, applied across web and product.

## The fiction

NexaGrid is a B2B infrastructure observability platform. It has been around long enough
to have a real customer base and a brand that was assembled rather than designed.

A rebrand is a stronger portfolio piece than a new brand because it forces a before and
an after, and the before is where you show judgement.

## Design the "before" honestly

Make a credible bad brand first - the kind a good engineering team ships when nobody owns
design. Generic geometric sans, a blue that came from a template, an abstract node-graph
mark, four accent colours with no system, screenshots at an angle.

Then the after has something to argue with. Do not make a strawman - if the before is
obviously terrible, the after proves nothing.

## What to establish

- Positioning, and what is being given up to own it
- Archetype and voice, with a sample sentence and the same sentence written badly
- Wordmark, and a symbol only if it earns its place
- Colour with a stated ratio in use, `oklch` for code and CMYK for print
- Type system with licensing noted for both web and print
- Iconography rules
- Diagram and illustration language - this is an observability product, so how systems
  are drawn is a brand decision, not an afterthought
- Photography or its deliberate absence
- Motion principles
- Clear space, minimum sizes, and the misuse rules

## Deliverables

- `BRAND.md`
- Brand guidelines as a `.pdf` - `canvas-design`
- Wordmark as `.svg` plus raster exports
- Marketing site, running - `vue-web-builder`
- Product UI applying the system
- Social and deck templates
- Mockups in context
- Case study with before and after side by side

## Skills for this project

This one is brand-led, so the brand group carries it and the build is the application.

| Phase | Skill | Why |
| --- | --- | --- |
| 1 brief | `brand-discovery`, `creative-brief` | |
| 2 research | `competitor-experience-audit` | the before has to be credible, not a strawman |
| 3 direction | `brand-archetype-system` | positioning frame |
| 3 direction | `graphic-design-styles` | Swiss plus a diagram language |
| 4 system | `brand-identity`, `logo-design`, `brand-voice` | the identity |
| 4 system | `brand-style-guide` | the guidelines document, a real deliverable |
| 4 system | `ai-graphic-design` | vectorisation and the IP side |
| 4 system | `art-direction` | photography or its deliberate absence |
| 6 build | `canvas-design` | the guidelines PDF |
| 6 build | `vue-web-builder` | site and product UI |
| 8 gates | `design-auditor`, `webapp-testing` | |

**Not used:** the motion audit set, `algorithmic-art`, `mobile-native`.

## Category clichés to refuse

Node-graph or constellation marks, purple-to-blue gradient, a dark dashboard screenshot
floating at 15 degrees, "observability for the modern stack", a logo wall above the fold.

## Notes on execution

`brand-identity`, `brand-style-guide`, `brand-voice` and `logo-design` from the brand
group. `ai-graphic-design` for vectorisation and IP.

The guidelines document is a real deliverable, designed as carefully as the site. Most
brand case studies show a logo on coloured squares; a guidelines PDF that a developer
could work from is more convincing.

Tokens have to be real. Every colour in `BRAND.md` appears in the site's `@theme`, with
the same name. That link is the point of the whole exercise.
