# CATsoup design system

Version 1, implemented 2026-11-04. The owner requested reusable standards matching
the paw icon, open layouts instead of cards, and no colored selection rail.

## Direction and layout rules

CATsoup is a local bilingual video workspace. Its character is warm, focused and
approachable. The video supplies imagery; the paw and paired language control
supply identity. The palette follows the owner-supplied artwork without tinting
the video itself. Use spacing, heading hierarchy and occasional thin rules to
group information. No card grids, nested panel boxes, decorative shadows or
colored vertical selection markers. Selected navigation uses a soft background
and accessible current-state attributes. Outlines are appropriate for controls;
the player keeps a neutral dark surface and its own bounded media area.

## Palette

| Role | Token | Value | Use |
|---|---|---|---|
| Canvas | `--page` | #FFFCF7 | Main workspace |
| Cream | `--sidebar` | #FBF3E8 | Navigation |
| Cocoa | `--text` | #4F281B | Titles and body |
| Secondary | `--secondary` | #705446 | Supporting copy |
| Muted | `--muted` | #786052 | Metadata; never lighter for small text |
| Orange | `--accent` | #FF793F | Primary actions and active language |
| Orange hover / pressed | `--accent-hover` / `--accent-pressed` | #F36B32 / #E95D25 | Pointer feedback |
| Text on orange | `--on-accent` | #361608 | Dark text, not white |
| Soft selection | `--accent-soft` | #F9E4D3 | Selected navigation |
| Separator | `--line` | #E4D7C8 | Noninteractive quiet rules |
| Control boundary | `--control-border` | #A48C7D | Inputs and secondary buttons |
| Focus / accent text | `--accent-ink` | #9D3E15 | Keyboard outline; orange text alternative |
| Success | `--success` | #386247 | Confirmed success/approval only |
| Attention | `--warning` | #805A19 | Review required |
| Error | `--danger` | #A4352B | Failures and destructive actions |

Orange does not mean approved. Status always includes text, not color alone.
The tested normal text/background pairs meet at least 4.5:1 contrast, including
hover/pressed primary states and metadata on selection. This is not a claim of
complete accessibility certification. Disabled controls are visibly inactive.

## Typography

Instrument Sans is bundled locally with its SIL OFL 1.1 license in
`frontend/public/fonts/`. Source: Google Fonts' official `ofl/instrumentsans`
directory. No runtime font requests. Segoe UI is fallback; Consolas is reserved
for IDs and technical measurements. Weights: 400 body, 500 headings/actions,
600 short labels. Sentence case, no decorative uppercase tracking.

| Role | Size / line height | Shared class |
|---|---|---|
| Page title | 28 / 36 | `.type-title` or h1 |
| Section title | 20 / 28 | `.type-section` or h2 |
| Subtitle | 16 / 24 | `.type-subtitle` |
| Body | 14 / 24 | `.type-body` |
| Label | 14 / 20, weight 600 | `.type-label` or h3/h4 |
| Supporting controls/context | 13 / inherited | `--text-sm` |
| Caption | 12 / 16 | `.type-caption` |
| Technical metadata | 12, tabular | `.mono` |

## Controls and states

Use `UiButton.vue` instead of inventing button styles per screen:

- `variant`: primary, secondary (default), quiet, danger.
- `size`: default (40px minimum height), compact (32px).
- `type`: button by default; submit/reset only when needed.
- `disabled` and `busy`: both prevent activation; busy exposes `aria-busy`.
- Keep descriptive text while working. At most one primary action per task.
- Hover changes surface color; pressed is stronger; keyboard focus is a 2px outline
  with 3px offset. Focus is not the removed decorative navigation rail.

`UiSegmented.vue` uses those same buttons, a labeled group and `aria-pressed`.
It accepts `modelValue`, labeled options and emits `update:modelValue`.
Use native inputs/selects with `.ui-field`, visible labels, `aria-invalid` and
an associated error description when necessary. Status primitives use `.ui-status`
and a semantic modifier. Do not turn statuses into decorative pills.

## Spacing and geometry

4px base: 4, 8, 12, 16, 20, 24, 32, 40, 48. Inside groups use 8–16; between
sections use 24–32. Controls have 8px corners; grouped inner buttons 4px;
media 12px. No general card radius or shadow token is provided deliberately.
Transition only control colors/borders over 140ms, with reduced-motion support.

## Desktop shell and modules

Use `brand/catsoup-rounded.png` for the paw icon: the original 150px artwork with
a 24px antialiased, transparent corner mask. Do not redraw or recolor the paw.
The source `catsoup.png` stays untouched. The rounded asset is precomputed, not
processed at runtime; the same asset supplies native icon resources and the UI.

The custom 40px title bar is part of CATsoup's design and replaces the Windows
title bar via Wails frameless mode. Its background is cream; the content region
retains the canvas color. The drag region excludes controls. Use Wails runtime
methods for minimize, maximize/restore and close, with accessible names and keyboard
focus. Window state is event-driven; never poll it. Native behavior remains native,
including OS drag/resize mechanics and the system keyboard close shortcut.

The sidebar starts with Modules. Publisher is available; Trends is explicitly
disabled/planned. Episodes form a separate group within Publisher. Delivery review
is a Publisher task, not the application's top-level identity. This navigation does
not merge the independent Trends backend or introduce an unimplemented module.

Keep the shell lightweight: no UI framework additions, icon library, animation
library, router or background module loading for this layout. The three window
symbols are inline SVG paths. The design catalog is a separate entry and is not
loaded by the Publisher page.

## Source of truth and preview

- `frontend/src/design/tokens.css`: colors, typography, spacing, dimensions, motion.
- `frontend/src/design/primitives.css`: shared styles and interaction states.
- `frontend/src/components/UiButton.vue`, `UiSegmented.vue`: reusable behavior.
- `frontend/src/style.css`: Publisher-specific layout only.
- `frontend/src/design/DesignSystem.vue`: live catalog using the actual components.
- `frontend/design-system.html`: separate catalog entry, built alongside the app;
  not an extra navigation item or a production publishing feature.

Run `node apps/publisher/support/check_design.mjs` after the visual test prerequisites
in README. It checks token contrasts, disabled/busy controls, keyboard focus, language
selection and font availability; it renders `work/publisher-ui-check/design-system.png`.
Run `check_ui.mjs` for delivery playback regression and responsive app screenshots.
The original native playback was owner-verified; browser fixtures are not native QA.

New screens must use these foundations. Add a token or variant only for a repeated
semantic need, not to accommodate arbitrary one-off styling. A broad brand change
belongs in these shared files and this document, not a cascade of screen overrides.

Design process informed by [Unslop](https://github.com/heydamir/unslop), by Damir
Sokolovsky, CC BY 4.0. CATsoup's paw artwork was supplied separately by the owner.
