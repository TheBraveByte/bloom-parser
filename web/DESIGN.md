# DESIGN.md — Bloom Parser console

## Job

Internal test console for the bloom-parser service. Operators drop scanned
documents, pick an extraction pipeline, inspect structured output, and export
CSV. A tool, not a marketing page — density and clarity over decoration.

## Direction

Dark technical console. The interface should feel like a lab instrument:
quiet chrome, visible state, honest numbers.

- **Tone**: functional, engineering
- **Palette**: deep charcoal panels, one accent blue, amber for flagged data,
  red for errors, green for success
- **Type**: system sans for UI, mono for filenames, stats and cell data
- **Motion**: none beyond a 150 ms state transitions; the only flourish is the
  drop-zone border glow while dragging

## Tokens

Declared in `src/style.css` under `@theme`. No ad-hoc values in components.

## Layout

- Header: product name + pipeline indicator
- Sidebar (360 px): drop zone, file list, pipeline controls, action button
- Result pane: status pills, per-file errors, fact table / document pages
