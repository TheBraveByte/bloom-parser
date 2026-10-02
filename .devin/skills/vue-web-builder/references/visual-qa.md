# Visual QA

A build that compiles is not a build that is done. Run this before reporting
finished. Look at the result - do not infer it from the code.

## Setup

```bash
bun run dev
bunx playwright install chromium   # first time only
```

## Screenshot pass

```bash
bunx playwright screenshot --viewport-size=390,844  --full-page http://localhost:5173 shots/mobile.png
bunx playwright screenshot --viewport-size=768,1024 --full-page http://localhost:5173 shots/tablet.png
bunx playwright screenshot --viewport-size=1440,900 --full-page http://localhost:5173 shots/desktop.png
```

Open each one. Every route, not just the home page.

## Checklist

Layout
- No horizontal scroll at 320px wide
- Nothing clipped, nothing overlapping
- Consistent gutters down the page
- Images cropped sensibly at every width, subject not cut off
- No orphaned single item on its own grid row at an awkward width

Type
- Hero readable at 320 and not absurd at 1920
- Body measure under 75 characters
- No widows on headings
- Every size on the scale

Colour and depth
- Contrast passes, including disabled and placeholder states
- One light direction across all shadows
- Borders visible against their background

States
- Hover, focus-visible, active, disabled on every interactive element
- Loading state exists and does not shift layout when it resolves
- Empty state exists and says something useful
- Error state exists

Motion
- Nothing slower than 600ms
- No layout shift during animation
- Reduced motion respected

Console
- No errors, no Vue warnings, no 404s in the network tab

## Overflow check

Paste in the browser console:

```js
[...document.querySelectorAll('*')]
  .filter(el => el.scrollWidth > document.documentElement.clientWidth)
  .forEach(el => console.log(el.tagName, el.className, el.scrollWidth))
```

## Report honestly

If something is off and you are not fixing it, say so. Do not report a pass on a
page you did not open.
