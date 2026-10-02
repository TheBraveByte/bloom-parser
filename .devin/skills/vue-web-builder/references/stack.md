# Stack

## New project

```bash
bun create vite@latest my-app -- --template vue-ts
cd my-app
bun install
bun add vue-router
bun add -d @tailwindcss/vite tailwindcss vue-tsc
bun add -d vitest @vitest/browser playwright
bun run dev
```

Add Pinia only when two unrelated components need the same mutable state:

```bash
bun add pinia
```

## Versions verified 2026-09-26

Treat these as the floor, not a pin. Run `bun outdated` and move up unless a
major bump is known to break something.

| Package | Version |
| --- | --- |
| vue | 3.5.43 |
| vite | 8.3.1 |
| @vitejs/plugin-vue | 6.0.9 |
| vue-router | 5.3.1 |
| pinia | 4.0.3 |
| tailwindcss | 4.3.3 |
| @tailwindcss/vite | 4.3.3 |
| typescript | 7.0.2 |
| vue-tsc | 3.3.11 |
| vitest | 5.0.2 |
| @vueuse/core | 15.0.0 |
| motion-v | 2.4.4 |
| playwright | 1.63.0 |
| eslint | 10.11.0 |
| prettier | 3.9.9 |

Vue Router 5 and Pinia 4 are both majors past what most training data and most
blog posts cover. Check the current docs before copying a pattern from memory.

## vite.config.ts

```ts
import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import tailwindcss from '@tailwindcss/vite'
import { fileURLToPath, URL } from 'node:url'

export default defineConfig({
  plugins: [vue(), tailwindcss()],
  resolve: {
    alias: { '@': fileURLToPath(new URL('./src', import.meta.url)) },
  },
})
```

## package.json scripts

```json
{
  "scripts": {
    "dev": "vite",
    "build": "vue-tsc --build && vite build",
    "preview": "vite preview",
    "typecheck": "vue-tsc --noEmit",
    "lint": "eslint . --fix",
    "test": "vitest run"
  }
}
```

## Bun notes

- `bun install` not `npm install`. There is one lockfile, `bun.lock`, and it is committed.
- `bunx` not `npx`.
- `bun --bun run dev` forces Vite onto Bun's runtime. Useful, but drop it the moment
  a plugin misbehaves - Vite on Node is the supported path.
- `bun outdated` to see drift. `bun update --latest` to take majors, then run the doctor.
