# Motion, particles and 3D

Reach for these in order. Most sites need only the first two.

| Need | Tool | Version |
| --- | --- | --- |
| enter, leave, list transitions | Vue `<Transition>`, `<TransitionGroup>` | built in |
| springs, layout animation, gestures, scroll-linked motion | `motion-v` | 2.4.4 |
| timelines, scroll-driven sequences, SVG morphs, complex choreography | `gsap` | 3.15.0 |
| smooth scroll | `lenis` | 1.3.26 |
| particles | `@tsparticles/vue3` + `@tsparticles/slim` | 4.4.0 |
| 3D | `three` | 0.186.1 |

```bash
bun add motion-v                              # the usual answer
bun add gsap                                  # only for timelines
bun add lenis
bun add @tsparticles/vue3 @tsparticles/slim
bun add three
```

## motion-v

The Vue equivalent of Framer Motion, close enough that React motion advice transfers
directly. Use it for anything beyond a fade.

```vue
<script setup lang="ts">
import { motion } from 'motion-v'
</script>

<template>
  <motion.div
    :initial="{ opacity: 0, y: 24 }"
    :while-in-view="{ opacity: 1, y: 0 }"
    :in-view-options="{ once: true, margin: '-80px' }"
    :transition="{ duration: 0.5, ease: [0.22, 1, 0.36, 1] }"
  >
    <slot />
  </motion.div>
</template>
```

## GSAP

Only when `motion-v` cannot express it - a scroll-driven sequence with several
overlapping steps, or SVG path morphing. Always kill the timeline on unmount:

```ts
onUnmounted(() => { tl.kill(); ScrollTrigger.getAll().forEach(t => t.kill()) })
```

GSAP plus Lenis needs Lenis driving GSAP's ticker, or the two fight over scroll.

## Particles

`@tsparticles/slim`, not the full bundle - the full build is several hundred kilobytes
for effects you will not use.

Rules:
- Cap particle count. 60 to 120 for a hero. 300 kills a mid-range phone.
- `pauseOnOutsideViewport: true`.
- Disable entirely under `prefers-reduced-motion`, and skip on mobile unless it is the
  point of the page.
- Particles behind content need low contrast against the background or the text
  becomes unreadable as they drift past.
- A CSS gradient or a noise texture is usually the better answer. Reach for particles
  when motion is the message.

## 3D

`three` directly, driven by an `onMounted` setup and a `requestAnimationFrame` loop that
you stop in `onUnmounted`. Lazy-load the whole scene with a dynamic import - Three is
around 600kB and does not belong in the initial bundle.

Check the frame rate on a phone before committing to it.

## Reference for this kind of work

[Cuberto](https://cuberto.com) is the benchmark for particle, cursor and scroll-driven
motion done as art direction rather than decoration. Study the restraint: one
signature effect per page, everything else still. Copying three of their effects onto
one page gives you the opposite of their result.
