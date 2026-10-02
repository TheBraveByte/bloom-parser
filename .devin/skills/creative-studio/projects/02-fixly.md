# 02 - Fixly

Self-initiated concept project.

**Proves:** AI product design, marketplace UX, two-sided flows, payments, mobile.

## The fiction

Fixly connects homeowners with local trade professionals. The entry point is a
conversational intake that turns a plain description of a problem into a structured job.

## The core interaction

Someone types:

> My bathroom ceiling has started leaking after heavy rain.

The product produces a structured request - category, problem, urgency, location,
appointment window, photos requested, estimated scope - and shows it back for
confirmation before anything is sent to a professional.

That confirmation step is the whole design problem. Get it wrong and the AI feels like
it is guessing on the user's behalf. The user has to see what was understood, correct it
in one tap, and know what happens next.

## Category clichés to refuse

A chat bubble as the entire interface, a sparkle icon on everything, "AI-powered" in the
hero, gradient mesh backgrounds, an isometric illustration of a house.

## Two sides to design

**Homeowner:** marketing site, intake, structured job confirmation, matching, provider
comparison, quote comparison, booking, payment, job status, messaging, review.

**Professional:** onboarding, verification, profile, incoming jobs, job detail, quote
submission, schedule, earnings, customer messaging.

## What the design has to carry

Trust is the product. Every screen answers one of: who is this person, what will it
cost, what if it goes wrong.

- Pricing transparency - ranges before commitment, no surprise line items
- Provider credibility - verification shown as evidence, not a badge
- Safety - what is checked, and what is not
- Payment states - authorised, held, released, refunded, disputed
- Cancellation and no-show, from both sides

## Deliverables

- Running Vue 3 implementation of the core flows, both sides
- Marketing site
- `DESIGN.md`, product brief, information architecture, user flows
- Case study

## Notes on execution

Realistic data. Real suburb names, plausible trade pricing for one named city, real
trade categories. Made-up data makes a marketplace look like a mockup instantly.

Mobile first and mobile real - homeowners photograph a leak standing in a bathroom.
`mobile-native` covers what separates a web app from something that feels installed.

The AI output needs a visible confidence boundary. When it is unsure of urgency or
category, it says so and asks. A product that is always confident is a product that is
sometimes confidently wrong.

## Skills for this project

| Phase | Skill | Why |
| --- | --- | --- |
| 1 brief | `jtbd-framing` | the job is "stop the leak", not "find a plumber" |
| 2 research | `ux-research`, `journey-mapping` | two-sided journeys, and where trust breaks |
| 3 direction | `graphic-design-styles` | minimalism, clay accents only |
| 4 system | `design-system` | it is two products sharing one system |
| 5 structure | `information-architecture` | flows and every state |
| 6 build | `vue-web-builder` | the build |
| 6 build | `mobile-native` | homeowners use this standing in a bathroom |
| 6 build | `ask-sonner` | confirmations and failures, via `vue-sonner` |
| 8 gates | `ux-audit`, `design-auditor`, `webapp-testing` | |
| 8 gates | `usability-testing` | walk the two journeys as each persona |

**Not used:** `canvas-design`, `algorithmic-art`, `pitch-deck-visuals`, motion audit set
beyond `animate`.

## Gates that matter most here

States (gate 4). A marketplace is mostly states - no providers available, quote expired,
payment failed, professional cancelled, job disputed. Design them.
