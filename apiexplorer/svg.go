//go:build !wasm

package apiexplorer

import (
	"webtyp.com/icons/code"
	"webtyp.com/icons/terminal"
	"webtyp.com/svg/sprite"
)

// IconSvg ships the explorer's nav glyph (code) and the glyph of the «Developer tools» group
// layout/app.BuildMenu wraps it in (terminal), so both menu entries point at symbols that exist.
func (e *Explorer) IconSvg() *sprite.Sprite {
	return sprite.NewSprite(code.Def(), terminal.Def())
}
