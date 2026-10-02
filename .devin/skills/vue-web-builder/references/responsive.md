# Responsive

Mobile-first. Write the small-screen style unprefixed, then add `sm:` `md:` `lg:`.
If you find yourself writing `max-lg:` a lot, you built desktop-first by accident.

## Breakpoints

Tailwind defaults are fine: 40rem, 48rem, 64rem, 80rem, 96rem. Add a token in
`@theme` only if the design genuinely breaks at another width.

The real widths to check: 320, 390, 768, 1024, 1440, 1920.
320 is the one that breaks. Test it.

## Things that break on mobile and get shipped anyway

- Horizontal overflow from a fixed width, a long unbroken string, or a grid whose
  columns do not collapse. Find it with `document.documentElement.scrollWidth > window.innerWidth`.
- Hero text set with `vw` units and no `clamp()` floor - unreadable at 320.
- Hover-only interactions. Touch devices have no hover. Every hover reveal needs a
  tap equivalent.
- Tap targets under 44px.
- `100vh` on mobile Safari. Use `100dvh`.
- Fixed headers eating a third of a short viewport.
- Sticky elements inside an `overflow: hidden` ancestor - silently does nothing.
- Inputs with `font-size` under 16px - iOS zooms the page on focus.
- Bottom-fixed bars sitting under the home indicator. Use
  `padding-bottom: env(safe-area-inset-bottom)`.

## Layout

Prefer intrinsic sizing over breakpoints where you can:

```html
<div class="grid grid-cols-[repeat(auto-fit,minmax(18rem,1fr))] gap-6">
```

That collapses on its own and needs no media query. Use it for card grids.

Container queries with `@container` when a component must react to its own width
rather than the viewport - a sidebar card, a widget that appears in two places.
