//go:build !wasm

package apiexplorer

import (
	"webtyp.com/icons/terminal"
	"webtyp.com/svg/sprite"
)

// IconSvg ships the explorer's nav glyph (the shared command-prompt icon) to
// the page sprite, so the menu entry Icon() points at a symbol that exists.
func (e *Explorer) IconSvg() *sprite.Sprite {
	return sprite.NewSprite(terminal.Def())
}
