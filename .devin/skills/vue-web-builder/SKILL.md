---
name: vue-web-builder
description: Build and finish polished websites and web apps with Vue 3, Vite, Bun and Tailwind v4. Use for any new site or app, any "make this look better" or redesign request, any build from a screenshot or design reference, when porting a React component or snippet to Vue, when picking a component or animation library, and before calling any frontend work done. Covers stack setup, design direction via DESIGN.md, motion, particles and 3D, component library choice, responsive and accessibility rules, translating React advice into Vue, and a visual QA pass that must run before anything is reported as finished.
---

# Vue web builder

This skill covers websites and web apps. For a poster, billboard, logo, brand identity,
social creative, deck or generative art, `design-router` picks the right skill instead.

Fixed stack. Do not substitute:

- Runtime and package manager - Bun
- Build tool - Vite
- Framework - Vue 3, script setup, TypeScript
- Styling - Tailwind v4 via `@tailwindcss/vite`
- Router - vue-router
- State - Pinia, only when component state is genuinely not enough
- Tests - Vitest, Playwright for browser checks

Never React, Next.js, npm, yarn, pnpm, webpack or Tailwind v3 config files. If a
snippet, a library or another skill hands you React, port it - see
`references/react-to-vue.md`. Do not install a React package to use one component.

If the user explicitly asks for React, say the stack rule once and follow their call.

## Order of work

1. Direction before code. Read `DESIGN.md` at the project root. If there is none, write
   one from `templates/DESIGN.md` and get it agreed before building screens.
2. Scaffold or verify the stack - `references/stack.md`.
3. Choose at most one primary component source - `references/component-libraries.md`.
4. Build. Tokens first, then layout, then components, then motion.
5. Run the doctor - `bin/doctor.sh` in this repo.
6. Run the visual QA pass - `references/visual-qa.md`.

## Rules that decide quality

- One type scale, one spacing scale, one radius scale, one shadow scale. Declared in
  `@theme` in CSS, used everywhere. No ad-hoc `text-[13px]` or `p-[7px]`.
- One component source per project. One icon set. One motion library.
- Anything copied in from another library gets this project's `@theme` tokens before it
  gets merged.
- Real content or realistic placeholder content. Never "Lorem ipsum" in a deliverable.
- Layout gets a max width and a consistent gutter. Full-bleed is a deliberate choice.
- Motion is purposeful and fast - `references/motion.md`.
- One signature motion moment per page. Everything else is quiet.
- Every interactive element has hover, focus-visible, active and disabled states.
- Images have width, height, `loading`, and an art-directed crop.
- Dark mode only if asked. Half-done dark mode is worse than none.

## Working with the other installed skills

These are installed by `bin/install.sh` and carry the judgement this skill does not
duplicate. Take their reasoning, write the result in Vue.

| For | Use |
| --- | --- |
| design direction, critique, a word to steer with | `impeccable` |
| rejecting generic output | `taste-skill`, `frontend-design` |
| styles, palettes, type pairings by industry | `ui-ux-pro-max`, `design-system`, `ui-styling`, `brand` |
| building a specific animation | `animate` |
| judging animation already written | `review-animations` |
| auditing motion across a codebase | `improve-animations`, `find-animation-opportunities` |
| naming an effect you can only describe | `animation-vocabulary` |
| spring, gesture and sheet interactions | `apple-design` |
| the details that make UI feel finished | `emil-design-eng` |
| making a web app feel native on a phone | `mobile-native` |
| screenshot or reference site to code | `image-to-code` |
| pulling a system out of a live URL | `design-extractor` |
| contrast and token audit | `design-auditor` |
| trying several versions side by side | `prototype` |
| driving a real browser to check the result | `webapp-testing` |
| a palette when none is settled | `theme-factory` |

All of them assume React. `references/react-to-vue.md` is the bridge - read it once and
apply it silently from then on.

## Reference files

- `references/stack.md` - scaffold commands, pinned versions, config files
- `references/tailwind.md` - v4 token setup and the v3 habits to drop
- `references/react-to-vue.md` - package and API mapping, how to port a component
- `references/component-libraries.md` - what to install, what to only read
- `references/motion-libraries.md` - motion-v, GSAP, Lenis, particles, 3D
- `references/typography.md` - type scale, pairing, measure
- `references/motion.md` - timing, easing, what is safe to animate
- `references/responsive.md` - breakpoints, what actually breaks on mobile
- `references/accessibility.md` - the checks that get missed
- `references/visual-qa.md` - the pass that runs before you report done
- `references/assets.md` - Higgsfield image and video generation, server-side only
- `references/design-md.md` - what DESIGN.md is and where to source one

## Templates

- `templates/DESIGN.md` - the visual spec to fill in per project
