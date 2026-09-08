//go:build !wasm

package platformd

import "webtyp.com/svg/sprite"

// IconSvg registers the default platform chrome glyphs: the identity fallback,
// the brand fallback and the menu button. The full page sprite is the MERGE of
// every module's IconSvg() (platformd + crudview + components + the demo app,
// webtyp.com/app-demo), assembled by webtyp/ssr and injected inline
// in <body>.
//
// Content icons (what a module calls itself) live with their module, in the
// consuming application's repository: the chassis has no business drawing "home"
// or "devices".
func (p *Platform) IconSvg() *sprite.Sprite {
	return sprite.NewSprite(
		// The default identity glyph. A consumer may return its own from
		// Identity.UserIcon(); this is what the demo and any platform without a
		// bespoke avatar falls back to.
		sprite.Define(IconUser, "0 0 448 512", sprite.Path("M224 256c70.7 0 128-57.3 128-128S294.7 0 224 0 96 57.3 96 128s57.3 128 128 128zm89.6 32h-16.7c-22.2 10.2-46.9 16-72.9 16s-50.6-5.8-72.9-16h-16.7C60.2 288 0 348.2 0 422.4V464c0 26.5 21.5 48 48 48h352c26.5 0 48-21.5 48-48v-41.6c0-74.2-60.2-134.4-134.4-134.4z")),
		// The default brand mark. A consumer may return its own from
		// Brand.BrandMark(); this is what any platform without a bespoke logo
		// falls back to, the same way IconUser backs a missing avatar.
		sprite.Define(IconBrand, "0 0 576 512", sprite.Path("M259.3 17.8L194 150.2 47.9 171.5c-26.2 3.8-36.7 36.1-17.7 54.6l105.7 103-25 145.5c-4.5 26.3 23.2 46 46.4 33.7L288 439.6l130.7 68.7c23.2 12.2 50.9-7.4 46.4-33.7l-25-145.5 105.7-103c19-18.5 8.5-50.8-17.7-54.6L382 150.2 316.7 17.8c-11.7-23.6-45.6-23.9-57.4 0z")),
		sprite.Define(iconMenu, "0 0 448 512", sprite.Path("M16 132h416c8.837 0 16-7.163 16-16V76c0-8.837-7.163-16-16-16H16C7.163 60 0 67.163 0 76v40c0 8.837 7.163 16 16 16zm0 160h416c8.837 0 16-7.163 16-16v-40c0-8.837-7.163-16-16-16H16c-8.837 0-16 7.163-16 16v40c0 8.837 7.163 16 16 16zm0 160h416c8.837 0 16-7.163 16-16v-40c0-8.837-7.163-16-16-16H16c-8.837 0-16 7.163-16 16v40c0 8.837 7.163 16 16 16z")),
	)
}
