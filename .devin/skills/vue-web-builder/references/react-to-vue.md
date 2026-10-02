# Reading React advice as Vue

Almost every good design skill and component source is written for React. The design
thinking transfers whole. The code does not. Translate, do not copy, and never install
a React package to satisfy a snippet.

## Package mapping

| React | Vue | Notes |
| --- | --- | --- |
| framer-motion / motion | `motion-v` | same API shape, `<motion.div>` becomes `<motion.div>` |
| react-spring | `motion-v` springs, or `@vueuse/motion` | |
| sonner | `vue-sonner` | same author's API, ported |
| radix-ui | `reka-ui` | headless primitives, same anatomy |
| shadcn/ui | `shadcn-vue` | same CLI and component names |
| lucide-react | `lucide-vue-next` | |
| react-router-dom | `vue-router` | v5 now, not v4 |
| zustand / jotai | `pinia` | |
| tanstack/react-query | `@tanstack/vue-query` | |
| react-hook-form | `vee-validate` + `zod` | |
| cmdk | `reka-ui` combobox, or `vue-command-palette` | |
| react-three-fiber | `trois`, or plain `three` | plain `three` is usually less friction |
| Aceternity / Magic UI | `inspira-ui` | Vue port of the same effects |
| React Bits | `vue-bits` | same author, same components |
| Origin UI, 21st.dev | no Vue port | read the markup, rewrite the component |

## API mapping

| React | Vue |
| --- | --- |
| `useState` | `ref` |
| `useMemo` / derived state | `computed` |
| `useEffect` | `watch`, `watchEffect`, `onMounted` |
| `useRef` for a DOM node | `useTemplateRef`, or `ref` plus a matching `ref=` attribute |
| `useCallback` | not needed |
| `React.memo` | not needed |
| `children` | `<slot>` |
| render prop | scoped slot |
| `className` | `class` |
| `onClick` | `@click` |
| `<>...</>` | multiple roots, allowed |
| context provider | `provide` / `inject` |
| `AnimatePresence` | `motion-v`'s `AnimatePresence`, or `<Transition>` |
| CSS-in-JS | Tailwind utilities |

## Translating a component from a React source

1. Read what the component does and what it looks like. Ignore the hooks.
2. Take the markup and the Tailwind classes verbatim. Those transfer unchanged.
3. Rewrite state as `ref` and `computed`.
4. Rewrite motion in `motion-v` or a `<Transition>`.
5. Replace every token with a `@theme` token from this project's `DESIGN.md`. A ported
   component carrying its source's colours is the fastest way to a page that looks
   like three different sites.
6. Check the keyboard and focus behaviour yourself. Ported components lose it quietly.

## When a skill tells you to do something React-shaped

Skills like `impeccable`, `emil-design-eng`, `animate` and `ui-ux-pro-max` give
excellent judgement in React syntax. Take the judgement - the curve, the duration, the
hierarchy call, the critique - and write it in Vue. If a skill's instruction is
specifically "install framer-motion" or "use Next.js Image", substitute the Vue
equivalent above and say that you did.

`ask-sonner` is about Sonner specifically. Its API guidance holds for `vue-sonner`,
which is a direct port. Its React setup instructions do not.
