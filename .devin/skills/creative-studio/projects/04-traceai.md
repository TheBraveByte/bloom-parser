# 04 - TraceAI

Self-initiated concept project.

**Proves:** complex software design, data visualisation, information architecture,
design system at scale.

## The fiction

TraceAI is an AI infrastructure platform for engineering teams. The piece is its usage
and cost dashboard - where teams find out what their AI spend is doing and why.

This is the project that shows you design software, not marketing sites.

## The real problem

The user is not browsing. They arrived because a bill was higher than expected, or a
latency alarm fired. The dashboard has one job: get them from "something is wrong" to
"here is what changed" in under a minute.

Design for that arrival, not for a tour.

## What the default bad version looks like

Sidebar, four stat cards, one line chart, one table. Everything the same size. No answer
to any question, just a surface for numbers.

Refuse it. The design problem is hierarchy - one number that matters, its movement, then
the breakdown that explains it, then the detail.

## Dimensions

Usage, cost, models, requests, latency, errors, teams, projects, environments,
providers, API keys, budgets, alerts.

Filterable by date, model, team, project, environment, provider. Filter state has to be
visible, shareable as a URL, and resettable.

## Data visualisation

This is where the project earns its place. Read the `dataviz` skill before writing the
first chart.

- Cost over time with a comparison period, not a bare line
- Spend by model as a ranked comparison, not a pie
- Latency as a distribution, because the p99 is the story and an average hides it
- Errors against volume, so a spike in one is readable against the other
- A table that is actually good - sortable, sticky header, aligned numerals, units in
  the header not in every cell

Tabular numerals throughout. Currency aligned on the decimal. A number that cannot be
compared by eye is a number nobody uses.

## Also design

- Budget and alert configuration, which is a form problem and usually where products
  give up
- The alert itself - in-app and email
- Over-budget state, which is the most important screen in the product
- Empty state for a new account with no data yet
- Loading that does not shift layout when the real numbers arrive
- Permission-denied, for a team member without billing access

## Deliverables

- Running Vue 3 dashboard - `vue-web-builder`, `reka-ui` or `shadcn-vue` as the base
- `DESIGN.md` with the full token set
- Component inventory
- Case study

## Notes on execution

Realistic data with realistic shape - spend is spiky, not a smooth upward line. One
model dominating cost. A weekend dip. An incident visible in the errors.

Dark mode is worth doing here because engineers will ask, but only if it is done fully.

## Skills for this project

| Phase | Skill | Why |
| --- | --- | --- |
| 1 brief | `jtbd-framing` | the user arrived because a bill was wrong |
| 3 direction | `graphic-design-styles` | minimal with editorial data display. Refuse aurora. |
| 4 system | `design-system` | a real component inventory, this is the densest project |
| 5 structure | `information-architecture` | hierarchy is the whole design problem |
| 6 build | `vue-web-builder` | build, on `reka-ui` or `shadcn-vue` |
| 6 build | built-in `dataviz`, then `data-visualization` | read `dataviz` before the first chart |
| 6 build | `pick-ui-library` | table, virtualisation and date-range choices |
| 8 gates | `design-auditor` | chart colours and disabled controls fail contrast most |
| 8 gates | `webapp-testing`, `ux-audit` | |
| 8 gates | `accessibility-audit` | |

**Not used:** `canvas-design`, brand group, `apple-design`, `algorithmic-art`.

## Gates that matter most here

Content (gate 3) and states (gate 4). Also accessibility - dashboards fail contrast in
chart colours and in disabled controls more than anywhere else.
