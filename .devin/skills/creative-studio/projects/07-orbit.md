# 07 - Orbit

Self-initiated concept project. Build last - it reuses everything the earlier six
established.

**Proves:** motion design, product film, storyboarding, sound, art direction over time.

## The fiction

Orbit is an AI infrastructure platform. The deliverable is a 45 to 60 second product
film - the kind that sits at the top of a homepage or opens a conference talk.

You can reuse TraceAI's dashboard as Orbit's product if that saves building a second
one. Say so in the case study.

## Why this project

Everything else in the portfolio is static or interactive. This proves you can make
software feel cinematic, which is the capability behind fintech customer stories, product
sizzle reels and SaaS marketing video.

## Sequence

A starting structure, to be argued with in phase 3, not followed blindly:

| Time | Beat |
| --- | --- |
| 0-5s | the problem, stated without words if possible |
| 5-12s | the system - scale made visible |
| 12-20s | product reveal |
| 20-30s | the dashboard doing something real |
| 30-40s | the AI workflow, one concrete task start to finish |
| 40-50s | result, with a number |
| 50-60s | wordmark and one line |

Sixty seconds is short. Most product films fail by trying to cover nine features. Pick
one thing the product does and show it properly.

## Craft that separates it

- **Typography in motion** - type that arrives with intent, held long enough to read.
  Most AI-made motion graphics fail here: text appears for 800ms and cannot be read.
- **UI animation that is honest** - real interface, real data, real timing. A dashboard
  that populates in 200ms is a lie the viewer feels even if they cannot name it.
- **One transition idea** used consistently, not eight.
- **Sound.** Motion without sound design is half a film. A bed, a few marked moments, and
  silence used deliberately. If you cannot do sound, say so in the case study rather
  than shipping it silent and hoping.
- **Pacing.** Cut on the beat. Hold the reveal a beat longer than feels comfortable.

## Deliverables

- The film, 45-60s, with sound
- A silent 15s cut for social, 9:16 as well as 16:9
- Storyboard
- Frame set - the key frames as stills
- The web page it lives on, with poster frame, no autoplay with sound, and a pause
  control
- Case study

## Notes on execution

`motion-frames` if the optional group is installed, for storyboard frames.
Higgsfield for generated video and for environment plates - it covers Sora, Veo and Kling
class models. `animate` for UI motion inside the film. `algorithmic-art` if the system
visualisation is better generated than animated by hand.

Captions, not just for accessibility - most of this is watched muted.

On the page: a decorative autoplaying hero video needs a poster frame, `muted`,
`playsinline`, a pause control, and a reduced-motion path that shows the poster instead.
Weight matters - a 12MB hero video undoes the performance work from every other project.

## Skills for this project

| Phase | Skill | Why |
| --- | --- | --- |
| 3 direction | `graphic-design-styles` | futuristic, disciplined by Swiss typography |
| 5 structure | storyboard before anything moves | |
| 5 structure | `motion-frames` (optional group) | storyboard frames |
| 6 build | `algorithmic-art` | the system visualisation, if generated beats hand-animated |
| 6 build | `vue-web-builder` | the page the film lives on |
| 7 motion | `animate` | UI motion inside the film |
| 7 motion | `review-animations` | |
| 7 motion | `animation-vocabulary` | naming the transitions for the case study |
| 8 gates | performance, gate 7 | a 12MB hero video undoes six projects of work |
| 8 gates | `webapp-testing` | poster frame, pause control, reduced-motion path |

Higgsfield for generated video and environment plates - Sora, Veo and Kling class models.

Reuse TraceAI's dashboard as the product rather than building a second one. Say so in
the case study.

**Not used:** brand group, `canvas-design`, `ux-audit`, `data-visualization`.

## Gates that matter most here

Motion (gate 8) and performance (gate 7). Also judgement - watch it the next day with
fresh eyes, and watch it muted.
