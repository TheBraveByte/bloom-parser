# Component libraries

Pick one primary source per project. Mixing three libraries gives you three visual
languages and three sets of focus behaviour.

## Vue-native, safe to install

| Library | Version | Install | Use for |
| --- | --- | --- | --- |
| `reka-ui` | 2.10.5 | `bun add reka-ui` | headless primitives - dialog, popover, combobox, select. Unstyled, accessible, keyboard correct. The right default for product UI. |
| `shadcn-vue` | 2.8.2 | `bunx shadcn-vue@latest init` | copies styled components built on Reka into your repo. You own the code. Best starting point for a dashboard or app. |
| `daisyui` | 5.7.46 | `bun add -d daisyui` then `@plugin "daisyui";` in CSS | pure CSS classes on top of Tailwind. No JS, framework-agnostic, works in Vue with zero bindings. Good for fast marketing pages and prototypes. |
| `inspira-ui` | copy in | https://inspira-ui.com | the Aceternity and Magic UI effects, ported to Vue. Marketing-page eye candy - spotlight cards, beams, marquees, text effects. |
| `vue-bits` | copy in | https://vue-bits.dev | React Bits by the same author, in Vue. Animated and interactive show pieces. |
| `vue-sonner` | 2.0.9 | `bun add vue-sonner` | toasts. Port of Sonner. |
| `lucide-vue-next` | | `bun add lucide-vue-next` | icons. One icon set per project. |

DaisyUI note: it is a v4-and-v5 Tailwind plugin loaded with `@plugin`, not a JS config
entry. Its semantic class names (`btn`, `card`, `badge`) coexist with utilities, but pick
a lane per component - half DaisyUI and half hand-built looks it.

## React-only, use as reference and port

These have no Vue build. Read the markup and the Tailwind classes, rewrite the logic.
See `react-to-vue.md`.

- **Origin UI** - https://originui.com - Cal.com's design system. Best source for
  forms, inputs and dense product UI. The markup is plain Tailwind and ports cleanly.
- **21st.dev** - https://21st.dev - community component registry, 10,000+ React and
  Tailwind components. Good for finding a pattern quickly. Its MCP server generates
  React, so use the site as a visual and markup reference rather than the MCP.
- **React Bits** - https://reactbits.dev - check `vue-bits` first, it covers most of it.
- **Aceternity, Magic UI** - check `inspira-ui` first.

## Rules

- Never install a React package to use one component. Port it.
- Components you copy in get your `@theme` tokens before they get merged. No exceptions.
- Headless plus your own styling beats a styled library you then fight.
- A library's default spacing and radius are its opinion, not yours. Override to your scale.
