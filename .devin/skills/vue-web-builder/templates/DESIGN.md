# DESIGN.md

Visual spec for this project. Fill every section. Delete the guidance in brackets.

## Intent

[One sentence: what this site is and who it is for.]
[One sentence: the feeling on first load.]

## Reference

[Two or three sites this should sit next to, and what specifically is being taken
from each. "Stripe" is not an answer. "Stripe's restraint with colour - one accent,
everything else neutral" is.]

## Not this

[What to avoid. Be specific about the generic version of this site, so it is clear
what is being rejected.]

## Type

| Role | Family | Weight | Size token | Tracking | Leading |
| --- | --- | --- | --- | --- | --- |
| Hero | | | `--text-hero` | | |
| Section heading | | | | | |
| Body | | | | | |
| Small / label | | | | | |

Scale ratio: [1.25 / 1.333 / 1.5]
Measure: [ch]

## Colour

All values in `oklch`.

| Token | Value | Used for |
| --- | --- | --- |
| `--color-ink` | | primary text |
| `--color-paper` | | page background |
| `--color-muted` | | secondary text |
| `--color-line` | | borders |
| `--color-accent` | | one thing only - name it |
| `--color-surface` | | raised surfaces |

Dark mode: [yes with a toggle / no]

## Space

Base unit: [rem]
Section rhythm: [values]
Gutter: [mobile / desktop]
Max content width: [rem]

## Shape and depth

Radius: [card / button / input]
Borders: [width, colour token]
Shadows: [how many levels, light direction, or none]

## Motion

Enter: [duration, easing]
Exit: [duration, easing]
Page transition: [what happens, or none]
Scroll behaviour: [reveals and where, or none]
Signature moment: [the one piece of motion people should remember]

## Imagery

Subject: []
Treatment: [lighting, grade, grain]
Crops: [aspect ratios in use]
Source: [Higgsfield model / photography / illustration]

## Layout

[Grid, section order, what is full-bleed and what is contained. Sketch it in text
if that is faster.]

## Components

[List the components this project needs, and any that deliberately do not exist.]
