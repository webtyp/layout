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
			style.Cover(),
		).
		Part(partHeader,
			style.Row(style.Space2),
			style.Pad(style.Space2),
			style.KeepSize(),
		).
		Part(partWork,
			style.Stack(style.Space1),
			style.Fill(),
		).
		Part(partThread,
			style.Grow(),
			style.Scroll(),
		)
}

// RenderCSS implements the visual contract for chatview using the style DSL.
func (v *ChatView) RenderCSS() *css.Stylesheet {
	return v.RenderSheet().Stylesheet()
}
