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
		Part("header",
			style.Row(style.Space2),
		).
		Part("table",
			style.Stack(style.Space1),
		).
		Part("row",
			style.Stack(style.Space2),
			style.Pad(style.Space2),
			style.As(style.Page), // instead of generic Surface type
			style.Round(style.RadiusMd),
			style.Interactive(style.Page),
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
		Part("col-access").
		Part("col-permission").
		Part("col-roles").
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
