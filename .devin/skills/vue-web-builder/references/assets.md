# Assets - Higgsfield

Higgsfield generates images and video through one async API covering many models
(Soul 2, Soul Cinema, Recraft, Ideogram, Qwen, plus video models). Use it for hero
imagery, textures and short background video instead of stock photos.

Docs: https://docs.higgsfield.ai/docs
Model catalogue: https://console.higgsfield.ai
LLM-readable index: https://docs.higgsfield.ai/docs/llms.txt - fetch this first,
model IDs and parameters change.

## Two ways in

- **MCP** - `bin/mcp.sh higgsfield` adds `https://mcp.higgsfield.ai/mcp` at user scope.
  OAuth opens a browser on the first tool call. Quickest for one-off generation while
  designing. That endpoint is not listed in Higgsfield's own docs, so if it breaks,
  fall back to the SDK.
- **SDK** - the documented path, and the only one to use inside a project's own scripts.
  Everything below covers this.

## Credentials

A credential is a key ID and a secret, created in the console.

```bash
export HF_CREDENTIALS="your-key-id:your-key-secret"    # TypeScript SDK
export HF_KEY="your-key-id:your-key-secret"            # Python SDK
```

Server-side only. The v2 TypeScript client refuses to run in a browser for exactly
this reason. Never put these in a Vite `VITE_` variable - those are compiled into
the client bundle and are public.

Generation is billed per request from a prepaid balance. Do not loop over prompts
without a cap.

## Node usage

```bash
bun add @higgsfield/client
```

```ts
import { config, higgsfield } from '@higgsfield/client/v2'

config({ credentials: process.env.HF_CREDENTIALS })

const result = await higgsfield.subscribe('higgsfield-ai/soul/v2/standard', {
  input: { prompt: 'Editorial portrait in soft daylight' },
  withPolling: true,
})

if (result.status === 'completed') {
  console.log(result.images?.[0]?.url)
}
```

Use `subscribe` for scripts. Use `submit` plus a webhook for anything in production,
keeping polling as the recovery path.

## Workflow for a site

1. Write the image direction into `DESIGN.md` first - subject, lighting, palette,
   crop, mood. Prompts come from that, not the other way round.
2. Generate into `assets/generated/` with a script under `scripts/`, not by hand.
3. Download the outputs. Generated URLs expire - do not hotlink them from the site.
4. Convert to AVIF and WebP, generate the sizes you need, commit the results.
5. Reference locally with explicit `width`, `height` and `loading="lazy"` for
   anything below the fold.

## Prompting for web imagery

- Name the lens and light, not the emotion. "85mm, window light from camera left,
  shallow depth of field" beats "beautiful and premium".
- Ask for negative space where text will sit.
- Generate at the aspect ratio you will use. Cropping a square into a 21:9 banner
  wastes most of the image.
- Keep a consistent lighting direction and palette across a set, or the page looks
  like a collage.
