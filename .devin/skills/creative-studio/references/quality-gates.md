# Quality gates

The definition of done. Every gate is pass or explicitly waived in writing with a
reason. "It compiles" is not a gate.

Run them in this order - cheap and mechanical first, judgement last.

## Gate 1 - the brief

- Output matches the positioning sentence
- Nothing on the clichés list appears in the output
- The signature moment exists and is the only one

Fails here mean going back to direction, not tweaking spacing.

## Gate 2 - stack

```bash
bin/doctor.sh <project>
```

Zero failures. Warnings each need a reason.

## Gate 3 - content

- Real or realistic content everywhere. No Lorem ipsum, no "Feature one".
- Every number is plausible and internally consistent
- Copy read out loud does not sound like a language model wrote it
- Names, dates and prices consistent across every screen

## Gate 4 - states

Every one of these exists and is designed, not defaulted:

empty, loading, partial, error, offline if relevant, success, over-limit or
over-quota, permission denied, first-run.

Most portfolio work fails here. It is also the fastest way to look like someone who has
shipped software.

## Gate 5 - accessibility

- `design-auditor` for contrast and tokens
- `bunx @axe-core/cli http://localhost:5173` for the mechanical failures
- Tab through it yourself. Axe does not catch focus order, bad alt text or a
  hover-only menu.
- `references/accessibility.md` in `vue-web-builder` for what gets missed

Contrast passes in disabled and placeholder states too, not just body text.

## Gate 6 - the browser

Not inference. Actually look.

```bash
bunx playwright screenshot --viewport-size=390,844  --full-page http://localhost:5173 shots/mobile.png
bunx playwright screenshot --viewport-size=768,1024 --full-page http://localhost:5173 shots/tablet.png
bunx playwright screenshot --viewport-size=1440,900 --full-page http://localhost:5173 shots/desktop.png
```

`webapp-testing` or the `chrome-devtools` MCP for anything interactive. Open every
screenshot. Every route, not just the home page.

Then the share card. Paste the real URL into a chat and look at what renders - a missing
or default `og:image` is a blank rectangle on every link anyone ever sends.
`og-image-design`.

Full checklist: `vue-web-builder/references/visual-qa.md`.

## Gate 7 - performance

- Lighthouse or `chrome-devtools` trace on the built site, not the dev server
- No layout shift on load. A hero that jumps is the first thing anyone notices.
- Fonts preloaded, images sized, `three` and particles lazy-loaded
- Test on a throttled connection once

## Gate 8 - motion

- `review-animations` on what you built
- Reduced motion honoured
- Nothing over 600ms without a reason
- No animation that fires on every scroll pass

## Gate 9 - print, if the deliverable is print

- Final trim size, bleed and safe margins correct
- CMYK, not RGB
- Fonts embedded or outlined
- Viewed at 100% on screen, and printed once if it is going to be printed
- Billboard or large format: legible at distance, one idea, text no smaller than a
  tenth of the height

## Gate 10 - judgement

Last, because it only means something once the mechanical gates pass.

- `impeccable` critique, asked to be harsh
- `taste-skill` on whether it went generic
- `ux-audit` if installed, for a structured pass with evidence and severity

Then your own read. Look at it cold, ideally the next day. The question is not "is
anything wrong" - it is "would a studio put this out".

## Reporting

State which gates passed, which were waived and why. If a gate failed and you did not
fix it, say that plainly rather than reporting a pass.
