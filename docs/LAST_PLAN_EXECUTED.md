---
PLAN: "fix(schemaview): the diagram takes the room, findings in a side rail"
EXECUTOR: local
REVIEWER: none
---

# Plan — `layout/schemaview`: the diagram was 2 px tall

## Why

The findings list was `KeepSize()` (never shrinks) above a `Scroll()` canvas: with 30 findings the canvas measured
2 px and the diagram was invisible. A rail beside the diagram keeps a clicked finding's table in view.

## Design gate

No exported symbol changes: a new part `body` laid out with `style.Sidebar(SideEnd, RailWide, Space3)`; canvas and
findings each `Scroll()`. Recipes only (**widget-styling**).

| Stage | File | Action |
|---|---|---|
| 1 | `schemaview/schemaview.go` | `body` wraps canvas + findings |
| 2 | `schemaview/css.go` | `Sidebar` body; findings `Scroll()` |

## Additions (review with the user, 2026-10-28)

- Findings rail reads as a list: a summary («0 errores · 7 avisos», keys `errors`, `warnings`), then one card per
  finding (`As(Panel)`, table bold, reason `As(Subtle)` smaller; errors `DangerWash`).
- The band word `erd.ModuleLabel` is translated here before the SVG is inserted (the server has no dictionary).
- The canvas has no frame of its own (`Scroll()` only): the bands are the frame, aligned with the search bar.
- The search says what it filters: «Search warnings by table or column» (replaces «Search tables»).
