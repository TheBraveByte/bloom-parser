# Tailwind v4

v4 has no `tailwind.config.js`. Configuration lives in CSS. If you find yourself
writing a JS config, you are following a v3 tutorial.

## src/style.css

```css
@import 'tailwindcss';

@theme {
  --font-display: 'Instrument Serif', serif;
  --font-sans: 'Inter', system-ui, sans-serif;

  --color-ink: oklch(0.18 0.01 260);
  --color-paper: oklch(0.98 0.005 90);
  --color-accent: oklch(0.62 0.19 28);

  --text-hero: clamp(2.75rem, 7vw, 6rem);

  --spacing-gutter: 1.5rem;
  --radius-card: 1rem;

  --ease-out-soft: cubic-bezier(0.22, 1, 0.36, 1);
}
```

Every token in `@theme` becomes a utility. `--color-ink` gives you `text-ink`,
`bg-ink`, `border-ink`. That is the point - define once, use as a class.

## Rules

- Colours in `oklch`. Lightness is perceptual, so a ramp built by stepping L
  actually looks evenly spaced.
- No arbitrary values for anything that repeats. `text-[13px]` in three files means
  a missing token.
- No `@apply` for component styling. If a pattern repeats, it is a Vue component,
  not a CSS class.
- Dark mode with `@variant dark`, driven by a `class` on `<html>`, not `prefers-color-scheme`
  alone - users want a toggle.
- Long class lists are fine and expected. Do not "clean them up" into custom CSS.

## Common v3 habits to drop

| v3 | v4 |
| --- | --- |
| `tailwind.config.js` theme extend | `@theme` in CSS |
| `@tailwind base/components/utilities` | `@import 'tailwindcss'` |
| `postcss.config.js` with the tailwind plugin | `@tailwindcss/vite` |
| `content: [...]` globs | automatic detection |
| `darkMode: 'class'` | `@variant dark (&:where(.dark, .dark *))` |
