//go:build !wasm

package chatview

import (
	"webtyp.com/css"
	"webtyp.com/widget/style"
)

// RenderSheet returns the style Sheet containing the rules for chatview.
func (v *ChatView) RenderSheet() *style.Sheet {
	return style.For(v).
		Root(
			style.Fill(),
			style.HideOverflow(),
		).
		Part(partHeader,
			style.Row(style.Space2),
			style.Pad(style.Space2),
			style.KeepSize(),
			style.As(style.Panel),
			style.DividerBelow(),
			style.FontWeight(style.WeightBold),
		).
		Part(partWork,
			style.Stack(style.SpaceNone),
			style.Fill(),
			style.HideOverflow(),
		).
		Part(partThread,
			style.Grow(),
			style.Fill(),
			style.HideOverflow(),
		)
}

// RenderCSS implements the visual contract for chatview using the style DSL.
func (v *ChatView) RenderCSS() *css.Stylesheet {
	return v.RenderSheet().Stylesheet()
}
