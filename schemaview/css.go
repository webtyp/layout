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
		Part("body", style.Sidebar(style.SideEnd, style.RailWide, style.Space3), style.Fill()).
		Part("findings", style.Stack(style.Space2), style.Scroll()).
		Part("summary", style.FontWeight(style.WeightBold), style.KeepSize()).
		Part("finding", style.Stack(style.Space1), style.Pad(style.Space2), style.As(style.Panel), style.Round(style.RadiusMd), style.Interactive(style.Panel), style.KeepSize()).
		Part("finding-error", style.As(style.DangerWash)).
		Part("finding-where", style.FontWeight(style.WeightBold)).
		Part("finding-why", style.FontSize(style.TextSm), style.As(style.Subtle)).
		Part("canvas", style.Scroll())
}

func (v *View) RenderCSS() *css.Stylesheet { return v.RenderSheet().Stylesheet() }
