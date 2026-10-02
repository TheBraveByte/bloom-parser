# Accessibility

The checks that actually get missed. Not a full WCAG walkthrough.

## Always

- Contrast: 4.5:1 body text, 3:1 for text above 24px and for UI borders. Check the
  states too - a disabled button and placeholder text usually fail.
- `:focus-visible` has a visible ring with 3:1 contrast against both the element and
  the page. Never `outline: none` without a replacement.
- One `<h1>` per page. Heading levels do not skip.
- Landmarks: `<header>`, `<nav>`, `<main>`, `<footer>`. One `<main>`.
- Every image has `alt`. Decorative images get `alt=""`, not a missing attribute.
- Form inputs have a real `<label>`, not a placeholder standing in for one.
- Errors are announced - `aria-invalid` plus `aria-describedby` pointing at the message.
- Buttons that only contain an icon get `aria-label`.
- Keyboard: tab through the whole page. Everything reachable, nothing trapped,
  order matches visual order.
- Modals trap focus, close on Escape, and return focus to the trigger.
- `prefers-reduced-motion` honoured - see motion.md.

## Vue specifics

- `<div @click>` is not a button. Use `<button>`. If you truly cannot, add
  `role="button"`, `tabindex="0"` and a keydown handler - but you almost always can.
- Route changes do not move focus. After navigation, focus the new page heading, or
  screen reader users stay where they were.
- `v-if` removing a focused element drops focus to `<body>`. Move it deliberately.
- Live regions for async updates: `aria-live="polite"` on a wrapper that exists
  before the content does.

## Checking

`bunx @axe-core/cli http://localhost:5173` catches the mechanical failures.
It does not catch bad focus order, bad alt text, or a hover-only menu. Tab through
it yourself.
