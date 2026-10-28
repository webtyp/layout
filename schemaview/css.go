//go:build !wasm

package schemaview

import (
	"webtyp.com/css"
	"webtyp.com/widget/style"
)

func (v *View) RenderSheet() *style.Sheet {
	return style.For(v).
		Root(style.Fill(), style.Stack(style.Space3), style.Pad(style.Space3)).
		Part("header", style.Stack(style.Space2), style.KeepSize()).
		Part("error", style.As(style.Danger)).
		Part("findings", style.Stack(style.Space1), style.KeepSize()).
		Part("finding", style.Row(style.Space3), style.Pad(style.Space2), style.As(style.Page), style.Round(style.RadiusMd), style.Interactive(style.Page)).
		Part("finding-error", style.As(style.DangerWash)).
		Part("canvas", style.Scroll(), style.As(style.Panel), style.Round(style.RadiusMd))
}

func (v *View) RenderCSS() *css.Stylesheet { return v.RenderSheet().Stylesheet() }
