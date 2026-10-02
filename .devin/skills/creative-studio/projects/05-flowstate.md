# 05 - Flowstate

Self-initiated concept project.

**Proves:** interaction design, mobile and desktop in one system, motion as function.

## The fiction

Flowstate combines a calendar with deep work sessions. Not another task list - a system
where planning and protecting focused time is the same act.

## The interaction problem

Every calendar app looks the same because the grid is a solved problem. The opportunity
is not the grid, it is the relationship between a plan and what happened.

Design that: the intended day, the actual day, and the gap. Shown without making the
user feel judged, which is a real design constraint and the reason most of these apps
fail.

## Two surfaces, one system

**Desktop.** Keyboard first. Every action reachable without the mouse. Command palette.
A dense day view that is still legible. Drag to reschedule with real physics.

**Mobile.** Thumb first. One-handed. Glanceable - the answer to "what now" visible
without scrolling. Focus mode is the primary screen on mobile, not a secondary one.

Same system, different priorities. Do not shrink the desktop layout and call it mobile.

## Focus mode

The signature moment. A session starts and the interface recedes - notifications
suppressed, calendar protected, one task named, time remaining.

Design the edges, which is where it becomes real:
- Interrupted mid-session, and how the session records that honestly
- Overrunning, and whether that is a success or a failure
- A meeting starting during a session
- Ending early
- What the record looks like afterwards

## Category clichés to refuse

Gradient timer rings, a tomato, "crush your goals", streak shaming, a dashboard of
productivity metrics that makes someone feel worse, animated confetti.

## Deliverables

- Running Vue 3 implementation, desktop and mobile layouts
- `DESIGN.md`
- Interaction specs for drag, focus transition, and the day transition
- Motion built and reviewed
- Case study, with the interaction shown as short recordings not stills

## Notes on execution

`apple-design` for the physics - springs, interruptible transitions, drag that feels
weighted. `animate` to build, `review-animations` to judge. `mobile-native` for the
things that separate a web app from an installed one.

Motion here is function, not polish. A drag that does not track the finger exactly is a
broken feature, not a rough edge.

Time zones, daylight saving and all-day events are where calendar work gets hard. Handle
at least one of them properly and say so - it is a credibility signal.

## Skills for this project

Motion-led, so the Emil Kowalski set carries it.

| Phase | Skill | Why |
| --- | --- | --- |
| 2 research | `ux-research` | the gap between plan and actual, without judging |
| 3 direction | `graphic-design-styles` | minimal surface, clay physics in motion |
| 4 system | `design-system` | one system, two priorities |
| 6 build | `vue-web-builder` | build |
| 6 build | `mobile-native` | one-handed, glanceable |
| 7 motion | `apple-design` | springs, interruptible transitions, weighted drag |
| 7 motion | `animate` | build it |
| 7 motion | `review-animations` | judge it |
| 7 motion | `animation-vocabulary` | naming the moves for the case study |
| 8 gates | `webapp-testing` | record it, this cannot be judged from stills |
| 8 gates | `design-auditor`, `ux-audit` | |

**Not used:** brand group, `canvas-design`, `data-visualization`, `pitch-deck-visuals`.

## Gates that matter most here

Motion (gate 8) and the browser (gate 6). This project cannot be evaluated from
screenshots; record it.
