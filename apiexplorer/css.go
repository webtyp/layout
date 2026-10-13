//go:build !wasm

package apiexplorer

import (
	"webtyp.com/css"
	"webtyp.com/widget/style"
)

func (e *Explorer) RenderSheet() *style.Sheet {
	return style.For(e).
		Root(
			style.Fill(),
			style.Stack(style.Space3),
			style.Pad(style.Space3),
		).
		// Stack, not Row: the search bar is the header's only child and a Stack
		// gives it the full width; a Row left it at its intrinsic 120px.
		// KeepSize: the header stays put while the table scrolls under it.
		Part("header",
			style.Stack(style.Space2),
			style.KeepSize(),
		).
		// The table is the scroll region. The root fills the stage, so without
		// a scroller the rows had nowhere to overflow but into each other.
		Part("table",
			style.Stack(style.Space1),
			style.Scroll(),
		).
		// KeepSize: a Stack emits min-height: 0, so inside the fixed-height
		// table a row shrank below its content and the text overlapped.
		Part("row",
			style.Stack(style.Space2),
			style.KeepSize(),
			style.Pad(style.Space2),
			style.As(style.Page), // instead of generic Surface type
			style.Round(style.RadiusMd),
			style.Interactive(style.Page),
		).
		Part("row-header",
			style.Row(style.Space3),
		).
		Part("row-expanded",
			style.As(style.Highlight),
		).
		Part("orphan",
			style.As(style.Danger),
		).
		Part("public",
			style.As(style.Secondary),
		).
		Part("col-method", style.FontWeight(style.WeightBold)).
		Part("col-path", style.Grow()).
		Part("error",
			style.As(style.Danger),
		).
		Part("try-form",
			style.Stack(style.Space3),
			style.Pad(style.Space3),
			style.As(style.Panel),
			style.Round(style.RadiusMd),
		).
		Part("path-form",
			style.Stack(style.Space2),
		).
		Part("body-form",
			style.Stack(style.Space2),
		).
		Part("result-panel",
			style.Stack(style.Space2),
			style.Pad(style.Space2),
			style.As(style.Page),
			style.Round(style.RadiusSm),
			style.HideOverflow(),
		)
}

func (e *Explorer) RenderCSS() *css.Stylesheet {
	return e.RenderSheet().Stylesheet()
}
