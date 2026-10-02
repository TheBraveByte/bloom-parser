# Typography

Type is the fastest way a site reads as considered or as generic.

## Scale

Pick a ratio and stay on it. 1.25 for dense interfaces, 1.333 or 1.5 for editorial.
Define every step as a token. Never a size that is not on the scale.

Hero sizes use `clamp()` so they scale with the viewport instead of stepping at
breakpoints:

```css
--text-hero: clamp(2.75rem, 7vw, 6rem);
```

## Pairing

Two families maximum. Usually one display face and one workhorse sans. A third
face needs a reason you can say out loud.

Combinations that hold up:

- Instrument Serif display + Inter body - editorial, warm
- Geist display + Geist body, weight contrast only - product, neutral
- Fraunces display + Inter body - characterful without being loud
- One grotesque at 3 weights - hardest to do well, cleanest when it works

## Settings that matter

- Measure: 60 to 75 characters for body. `max-w-[65ch]`.
- Line height: tighter as size grows. `leading-[0.95]` on a hero, `leading-relaxed` on body.
- Letter spacing: negative on large display text, `tracking-tight` or tighter. Positive
  only on small uppercase labels.
- Optical alignment: large display text needs a small negative left offset to look
  aligned with body text below it.
- `text-wrap: balance` on headings, `text-wrap: pretty` on paragraphs.
- Load fonts with `font-display: swap` and preload the display face only.

## Hierarchy

Three levels of emphasis on a screen, not six. Size, weight and colour are the
levers - use one or two per level, not all three.
