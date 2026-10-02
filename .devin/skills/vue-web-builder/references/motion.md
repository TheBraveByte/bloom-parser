# Motion

Bad motion is the most common reason an otherwise good site feels cheap.

## Budget

- Micro interactions - 120 to 200ms
- Element enters and exits - 250 to 400ms
- Page or section transitions - 400 to 600ms
- Anything above 600ms needs a reason

## Easing

- Enter - `cubic-bezier(0.22, 1, 0.36, 1)`. Fast start, soft settle.
- Exit - `cubic-bezier(0.4, 0, 1, 1)`. Leave quickly.
- Never `linear` except for continuous loops. Never `ease-in-out` on an enter.

## What to animate

Only `transform` and `opacity`. Animating `height`, `width`, `top` or `box-shadow`
drops frames. For height, animate `grid-template-rows` from `0fr` to `1fr`, or use
`interpolate-size: allow-keywords`.

## Vue specifics

- `<Transition>` and `<TransitionGroup>` for enter and leave. Built in, no dependency.
- `motion-v` when you need spring physics, layout transitions or scroll-linked motion.
- Scroll reveals via `useIntersectionObserver` from `@vueuse/core`. Never a scroll
  listener. Reveal once, then unobserve - re-animating on scroll back up is annoying.
- View Transitions API for route changes, behind a support check.

## Never `scroll-behavior: smooth` in CSS

`html { scroll-behavior: smooth }` also smooths *programmatic* `scrollTo()` calls.
Stepped screenshot scrolls then animate between targets, and IntersectionObserver
can skip elements that whip past mid-animation - scroll-reveal content never fires
and renders as blank space in shots and for users on jump navigation. Twice hit
in production (sable, crossfade). Do smooth anchor scrolling in JS instead:

```ts
document.addEventListener('click', (e) => {
  const a = (e.target as HTMLElement).closest?.('a[href^="#"]')
  const id = a?.getAttribute('href')?.slice(1)
  const target = id && document.getElementById(id)
  if (!target) return
  e.preventDefault()
  const reduce = matchMedia('(prefers-reduced-motion: reduce)').matches
  target.scrollIntoView({ behavior: reduce ? 'auto' : 'smooth' })
})
```

## Non-negotiable

```css
@media (prefers-reduced-motion: reduce) {
  *, *::before, *::after {
    animation-duration: 0.01ms !important;
    transition-duration: 0.01ms !important;
  }
}
```

## Restraint

Stagger a group by 40 to 80ms per item, and cap it - a 20 item stagger means the
last item arrives a second and a half late. Do not animate everything on a page.
Pick the two or three moments that carry meaning.
