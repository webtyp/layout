package platformd

// Identity is what the platform needs to know about whoever is logged in.
//
// It is a READ contract, not a store: platformd renders it and never mutates
// it. Whatever owns authentication — webtyp.com/user in a real
// application — supplies an implementation; the platform neither knows nor
// cares where the values come from.
//
// It asks for facts, not for presentation. The glyph drawn when there is no
// avatar is IconUser, which this package owns — an authentication package has
// no business choosing a sprite, and asking it to would put a rendering
// decision behind a login.
//
// Roles, plural, because they are plural: a user holds N of them and no
// ordering exists to say which one matters. Asking for a single "area" forced
// every implementation to pick arbitrarily, and that decision was invisible.
type Identity interface {
	// UserName is who is logged in.
	UserName() string
	// UserAvatar is the URL of their picture. Empty is normal and expected.
	UserAvatar() string
	// UserRoles are display names, not authorization codes. May be empty.
	UserRoles() []string
}

// Brand is what the platform calls itself in its own chrome. The shell asks for
// a mark and a name; how they are drawn, sized and spaced is platformd's
// business.
//
// It is a READ contract, not a store, mirroring Identity: platformd renders it
// and never mutates it. The consumer supplies facts, not presentation — a
// Brand never hands the shell an svg.Icon, because picking the sprite is a
// rendering decision this package owns (the same reasoning that keeps IconUser
// out of Identity).
type Brand interface {
	// BrandName is shown beside the mark, and is the mark's alt text.
	BrandName() string
	// BrandMark is a URL or inline SVG data URI. Empty is normal and expected:
	// the shell falls back to its own glyph, exactly as UserAvatar does.
	BrandMark() string
}
