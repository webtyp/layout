//go:build !wasm

package schemaview

import (
	"webtyp.com/icons/database"
	"webtyp.com/svg/sprite"
)

// IconSvg ships the screen's nav glyph (database) to the page sprite.
func (v *View) IconSvg() *sprite.Sprite { return sprite.NewSprite(database.Def()) }
