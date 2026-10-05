# DESIGN.md — Bloom Parser console

## Job

Internal test console for the bloom-parser service. Operators drop scanned
documents, pick an extraction pipeline, inspect structured output, and export
CSV. A tool, not a marketing page — density and clarity over decoration.

## Direction

Card-directory instrument, borrowing the prompts.chat visual language: clean
raised cards on a quiet surface, numbered steps, pill metadata, honest numbers.
Light and dark themes with a manual toggle (system preference is the default,
choice persists in localStorage).

- **Tone**: functional, engineering, directory-clean
- **Palette (light)**: paper background, white panels, hairline borders,
  indigo accent, amber flagged / red errors / green clean
- **Palette (dark)**: deep charcoal panels (the old lab-instrument palette),
  blue accent, same status semantics
- **Type**: system sans for UI, mono for filenames, stats and cell data
- **Motion**: 150–200 ms state transitions only; the drop zone tints while
  dragging
- **Icons**: Phosphor (`@phosphor-icons/vue`), one family

## Tokens

Declared in `src/style.css` under `@theme` (light) with a `[data-theme="dark"]`
override block. No ad-hoc values in components. Shared surfaces live in
`@layer components` (`.card`, `.opt`, `.pill`, `.field`, `.btn-*`).

## Layout

- Sticky header: mark + name + tagline, concurrency pill, theme toggle
- Hero strip: one-line pitch + capability pills
- Intake: card grid — Documents (drop zone + queue) beside Pipeline / Options /
  Run cards, stacked on small screens
- Results: KPI stat cards, quality bar, per-file pills, then a single facts
  card (search + copy + download in the card head, windowed table)
