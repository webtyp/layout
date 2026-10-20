package platformd

import (
	"sync"

	"webtyp.com/msgtype"

	"webtyp.com/components/themetoggle"
	"webtyp.com/components/usermenu"
	"webtyp.com/layout"
	"webtyp.com/svg"
	"webtyp.com/widget"

	. "webtyp.com/dom"
	. "webtyp.com/fmt"
	. "webtyp.com/html"
	"webtyp.com/time"
)

const NamePlatform widget.Name = "pd"

var (
	clsRoot            = NamePlatform.Root()
	clsHeader          = NamePlatform.Class("header")
	clsDrawerPanel     = NamePlatform.Class("drawer-panel")
	clsAppName         = NamePlatform.Class("app-name")
	clsDrawerIdentity  = NamePlatform.Class("drawer-identity")
	clsDrawerBrand     = NamePlatform.Class("drawer-brand")
	clsBrand           = NamePlatform.Class("brand")
	clsBrandMark       = NamePlatform.Class("brand-mark")
	clsBrandName       = NamePlatform.Class("brand-name")
	clsBrandMenu       = NamePlatform.Class("brand-menu")
	clsBrandTrigger    = NamePlatform.Class("brand-trigger")
	clsBrandPanel      = NamePlatform.Class("brand-panel")
	clsBrandPanelName  = NamePlatform.Class("brand-panel-name")
	clsBrandPanelVer   = NamePlatform.Class("brand-panel-version")
	clsBrandBackdrop   = NamePlatform.Class("brand-backdrop")
	clsHeaderTitle     = NamePlatform.Class("header-title")
	clsHeaderTitleText = NamePlatform.Class("header-title-text")
	clsMsgSlot         = NamePlatform.Class("msg-slot")
	clsMsg             = NamePlatform.Class("msg")
	clsHeaderRight     = NamePlatform.Class("header-right")
	clsBody            = NamePlatform.Class("body")
	clsStage           = NamePlatform.Class("stage")
	clsPanel           = NamePlatform.Class("panel")
	clsMenu            = NamePlatform.Class("menu")
	clsNavbar          = NamePlatform.Class("navbar")
	clsNavItem         = NamePlatform.Class("nav-item")
	clsNavLink         = NamePlatform.Class("nav-link")
	clsLinkText        = NamePlatform.Class("link-text")
	ClsNavIcon         = NamePlatform.Class("nav-icon")
	clsHamburger       = NamePlatform.Class("hamburger")
	clsNavOverlay      = NamePlatform.Class("nav-overlay")
	clsMsgStack        = NamePlatform.Class("msg-stack")
	clsMsgSlotMobile   = NamePlatform.Class("msg-slot-mobile")
	clsSubmenu         = NamePlatform.Class("submenu")
	clsSubnavItem      = NamePlatform.Class("subnav-item")
	clsSubnavLink      = NamePlatform.Class("subnav-link")
	clsSubnavIcon      = NamePlatform.Class("subnav-icon")
	clsChevron         = NamePlatform.Class("chevron")
	clsNavParent       = NamePlatform.Class("nav-parent")
	clsNavGroupIcon    = NamePlatform.Class("nav-group-icon")
	clsNavActiveIcon   = NamePlatform.Class("nav-active-icon")
)

const (
	IconUser        = svg.Icon("pd-user")
	IconBrand       = svg.Icon("pd-brand")
	iconMenu        = svg.Icon("pd-menu")
	iconChevronDown = svg.Icon("pd-chevron-down")
)

// UIModule is a module that provides its UI to the platform chassis.
// The chassis takes id/hash/route from ModelName() and the rest of the
// presentation from these methods.
type UIModule interface {
	layout.Module    // identity: ModelName() → used as ID
	Label() string   // text in the nav rail
	Icon() svg.Icon  // chassis renders via the sprite
	View() Component // module content (often a *rightpanel.RightPanel)
	Activate()       // called by chassis when the module is activated (on-demand loading)
}

func (p *Platform) WidgetName() widget.Name { return NamePlatform }
func (p *Platform) WidgetKind() widget.Kind { return widget.Menu }

// Platform is the typed skeleton root.
type Platform struct {
	Element

	// AppName titles the drawer on a phone when there is no Brand to lead it —
	// the panel opens wholesale there and there is room for a line of text. The
	// collapsed rail never shows it: text with no icon beside it would have to
	// appear from nowhere when the rail expands.
	//
	// It is a fallback for a missing Brand, not a second header: a Brand
	// supersedes it in the drawer, since showing both would say the app's own
	// name twice in the same panel. A platform with neither renders no head at
	// all above the drawer's nav.
	AppName string

	// AppVersion is shown with the app's name in the panel the header's mark
	// opens (what a support request asks for). Optional: empty shows the name
	// alone.
	AppVersion string

	// Brand is what the platform calls itself: the mark and the name at the
	// header's leading slot, mirrored at the head of the phone drawer where
	// there is no header to hold it. Tapping it — either surface — is the "go
	// home" control, landing on DefaultID (or the first viewable module).
	// Optional — a platform without a logo renders no brand slot, AppName
	// takes the drawer's head instead, and the header's message block stays
	// centred between nothing and the menu.
	Brand Brand

	// User is the logged-in identity. Required: the header's outer thirds and
	// the drawer's first entry are built from it.
	User Identity

	// UserActions slot — the app's own actions inside the user menu panel
	// (sign-out, site picker). The theme control is NOT one of them: the shell
	// always renders it in the same menu. Optional.
	// A factory, not a Component: the shell renders one menu per surface and a
	// single element cannot have two parents. Passing one instance put the same
	// element in both menus, which rendered twice with one id and left the copy
	// in the drawer inert.
	UserActions func() Component

	// Status slot — shown at the header RIGHT, before the user menu: what the app
	// says about itself right now (connection state, a new version waiting).
	// Optional. A factory, like UserActions.
	Status func() Component

	// Modules registered in order — appearance order in the nav rail.
	Modules []UIModule

	// CanView filters which modules the shell presents. nil = show all.
	CanView func(resource string) bool

	// DefaultID is the ModelName() of the module to show initially.
	// If empty, the first module is used.
	DefaultID string

	// IdleTimeout is the number of seconds without user activity anywhere on
	// the page before OnIdle fires once. Activity is whatever
	// dom.OnUserActivity reports (pointer movement/press, keypress, wheel,
	// scroll) — see Init. 0 — the default — disables the idle lock entirely:
	// nothing is armed, no listener changes behavior.
	IdleTimeout int

	// OnIdle is called when IdleTimeout seconds pass without activity. Required
	// when IdleTimeout > 0 — Init panics otherwise (a configured idle lock with
	// no consequence would be a silent failure). The platform does NOT re-arm
	// after firing: the next activity starts a fresh period. Typical use: post
	// the logout route and reload.
	OnIdle func()

	// internal state
	active              *SignalString
	menuOpen            *SignalBool
	brandOpen           *SignalBool // the header's app panel (mark → name and version)
	expandedGroup       *SignalString
	notifications       *SignalNodes // desktop toasts (header msg-slot)
	notificationsMobile *SignalNodes // mobile toasts (msg-stack under the hamburger)
	navIcon             *SignalNodes
	navStowed           *SignalBool
	idleTimer           time.Timer
	groupActiveIcons    []groupIconSlot

	rawNotifications []notification
	mu               sync.Mutex

	// lastScrollTop es el testigo de la última posición leída por onScroll. Los
	// eventos de scroll llegan por el hilo de JS, igual que los click, así que no
	// va bajo p.mu — ese mutex existe por time.AfterFunc, que descarta
	// notificaciones desde otra goroutine.
	lastScrollTop float64
	scrollAnchor  float64
}

type groupIconSlot struct {
	id    string
	nodes *SignalNodes
}

func (p *Platform) getGroupIconSlot(id string) *SignalNodes {
	for _, s := range p.groupActiveIcons {
		if s.id == id {
			return s.nodes
		}
	}
	return nil
}

// scrollStowFloor is the minimum scroll offset from the top before the bar can stow.
// Keeps the bar stable near the very top of the page.
const scrollStowFloor = 24

// scrollDeltaThreshold is the intentional travel distance in either direction
// to trigger a stow or reveal. Filters out touch jitter while responding promptly.
const scrollDeltaThreshold = 12

type notification struct {
	Type msgtype.Type
	Msg  string
	ID   string
	// expiryNs is the auto-dismiss deadline; 0 = persistent. The
	// timer is stored on the notification so pause/resume can cancel and
	// re-arm it — see pauseToast/resumeToast.
	expiryNs time.Nano
	timer    time.Timer
}

// Duration is how long a notification stays before dismissing itself. A
// Duration is a DECISION, not a number: 0-as-persistent and -1-as-automatic
// were magic numbers nobody could read at the call site, and auto-sizing
// needs the message text that only Notify has.
type Duration struct {
	millis func(msg string) int // nil = persistent
	auto   bool                 // built by Auto(): the caller left the length to the platform
}

// Auto sizes the duration to the message: ~350ms per word plus ~2.5s to
// notice the toast and find it, floored at 4s and capped at 10s so a long
// message cannot hold the screen hostage. Two words → 4s; fifteen → 7.75s.
// (It was 1.2s and a 2s floor: a toast on a phone, in the corner of the eye,
// was gone before it was read.)
func Auto() Duration {
	return Duration{auto: true, millis: func(msg string) int {
		words := Count(msg, " ") + 1
		ms := 2500 + words*350
		if ms < 4000 {
			return 4000
		}
		if ms > 10000 {
			return 10000
		}
		return ms
	}}
}

// An error left to Auto() is persistent: the platform, not each caller, owns
// the rule that an error must not vanish before it is read (see Notify).

// Persistent keeps the notification until the user dismisses it. The case
// for it is error reporting: a message that vanishes before it is read
// defeats the report (WCAG 2.2.1, Timing Adjustable — the user must be able
// to extend or disable a time limit).
func Persistent() Duration { return Duration{} }

// For pins an exact duration in milliseconds.
func For(ms int) Duration {
	return Duration{millis: func(string) int { return ms }}
}

// Init initializes the platform state and routing.
func (p *Platform) Init(ctx Ctx) {
	p.active = NewString("")
	p.menuOpen = NewBool(false)
	p.brandOpen = NewBool(false)
	p.expandedGroup = NewString("")
	p.notifications = NewNodes()
	p.notificationsMobile = NewNodes()
	p.navIcon = NewNodes()
	p.navStowed = NewBool(false)
	p.groupActiveIcons = nil

	for _, m := range p.Modules {
		if gm, ok := m.(GroupModule); ok {
			p.groupActiveIcons = append(p.groupActiveIcons, groupIconSlot{
				id:    gm.ModelName(),
				nodes: NewNodes(),
			})
			for _, ch := range gm.Children() {
				if u, ok := ch.(UsesNotifier); ok {
					u.UseNotifier(p)
				}
			}
		}
		if u, ok := m.(UsesNotifier); ok {
			u.UseNotifier(p)
		}
	}

	if p.IdleTimeout > 0 && p.OnIdle == nil {
		panic("platformd: IdleTimeout requires OnIdle")
	}

	OnHashChange(func(hash string) {
		if len(hash) > 0 && hash[0] == '#' {
			p.Activate(hash[1:])
		}
	})

	// El scroll no burbujea: en captura sobre el documento es la única forma de que
	// el chasis vea el desplazamiento de un contenedor que pertenece a otro paquete.
	// (No llama a p.activity() aquí: dom.OnUserActivity, armado más abajo, ya
	// escucha "scroll" por su cuenta — llamarlo dos veces sería la misma señal
	// de presencia contada por partida doble.)
	OnScrollCapture(func(top float64) {
		p.onScroll(top)
	})

	// La presencia es un hecho del documento, no del subárbol de esta shell:
	// dom.OnUserActivity ve el movimiento del puntero, teclas, wheel y scroll
	// en cualquier parte de la página. Se registra una sola vez — Init corre
	// una sola vez por componente — así que no hace falta un flag "ya armado".
	if p.IdleTimeout > 0 {
		OnUserActivity(p.activity)
		p.activity() // abrir la app cuenta como presencia
	}

	hash := GetHash()
	if hash != "" && len(hash) > 0 && hash[0] == '#' {
		id := hash[1:]
		if p.isViewable(id) {
			p.Activate(id)
		} else {
			p.fallback()
		}
	} else {
		p.goHome()
	}
}

// goHome activa el módulo raíz: el mismo cálculo que Init usa cuando la app
// abre sin hash, reutilizado por el clic en la marca — pulsar el logo lleva
// al mismo sitio que abrir la app desde cero.
func (p *Platform) goHome() {
	if p.DefaultID != "" && p.isViewable(p.DefaultID) {
		p.Activate(p.DefaultID)
		return
	}
	p.fallback()
}

func (p *Platform) isViewable(id string) bool {
	if p.CanView == nil {
		return true
	}
	if p.CanView(id) {
		return true
	}
	for _, m := range p.Modules {
		if m.ModelName() == id {
			if gm, ok := m.(GroupModule); ok {
				for _, ch := range gm.Children() {
					if p.isViewable(ch.ModelName()) {
						return true
					}
				}
			}
		}
	}
	return false
}

// activeIcon es el glifo del módulo en el que estamos. iconMenu es el respaldo:
// UIModule permite que Icon() devuelva la cadena vacía, y un botón sin glifo no
// se puede pulsar porque no se ve.
func (p *Platform) activeIcon() svg.Icon {
	id := p.active.Get()
	for _, m := range p.Modules {
		if gm, ok := m.(GroupModule); ok {
			for _, child := range gm.Children() {
				if child.ModelName() == id {
					if ic := child.Icon(); ic != "" {
						return ic
					}
					if ic := gm.Icon(); ic != "" {
						return ic
					}
					break
				}
			}
		}
		if m.ModelName() == id {
			if ic := m.Icon(); ic != "" {
				return ic
			}
			break
		}
	}
	return iconMenu
}

// activity rearms the idle timer; called from every presence signal.
//
// idleTimer is guarded by p.mu: the AfterFunc callback below runs on its own
// goroutine (the stdlib backend lane; WASM is single-threaded, but the type
// is shared code) and writes idleTimer to nil concurrently with whatever
// goroutine calls activity() next — exactly the race p.mu was already
// documented, on lastScrollTop, as existing for.
func (p *Platform) activity() {
	if p.IdleTimeout <= 0 {
		return
	}
	p.mu.Lock()
	if p.idleTimer != nil {
		p.idleTimer.Stop()
	}
	p.idleTimer = time.AfterFunc(p.IdleTimeout*1000, func() {
		p.mu.Lock()
		p.idleTimer = nil
		p.mu.Unlock()
		p.OnIdle()
	})
	p.mu.Unlock()
}

// activeLabel returns the display label of the currently active module.
func (p *Platform) activeLabel() string {
	id := p.active.Get()
	for _, m := range p.Modules {
		if gm, ok := m.(GroupModule); ok {
			for _, child := range gm.Children() {
				if child.ModelName() == id {
					if l := child.Label(); l != "" {
						return l
					}
					return gm.Label()
				}
			}
		}
		if m.ModelName() == id {
			return m.Label()
		}
	}
	return ""
}

// onScroll guarda la barra de cabecera mientras el usuario baja en el contenido y la
// devuelve de inmediato en cuanto sube, siguiendo el estándar de la industria (Material 3 enterAlways / iOS Safari).
func (p *Platform) onScroll(top float64) {
	if top <= 0 {
		p.lastScrollTop = 0
		p.scrollAnchor = 0
		p.navStowed.Set(false)
		return
	}

	if top > p.lastScrollTop {
		// Moving down into content
		p.scrollAnchor = top
		if top > scrollStowFloor && top-p.lastScrollTop >= scrollDeltaThreshold {
			p.lastScrollTop = top
			p.navStowed.Set(true)
		} else {
			p.lastScrollTop = top
		}
	} else if top < p.lastScrollTop {
		// Moving up (user intends to navigate or search)
		if p.lastScrollTop-top >= scrollDeltaThreshold || top <= scrollStowFloor {
			p.lastScrollTop = top
			p.scrollAnchor = top
			p.navStowed.Set(false)
		} else {
			p.lastScrollTop = top
		}
	}
}

func (p *Platform) fallback() {
	for _, m := range p.Modules {
		if gm, ok := m.(GroupModule); ok {
			for _, ch := range gm.Children() {
				chID := ch.ModelName()
				if p.isViewable(chID) {
					p.Activate(chID)
					return
				}
			}
			continue
		}
		id := m.ModelName()
		if p.isViewable(id) {
			p.Activate(id)
			return
		}
	}
}

// Render builds the DOM tree (implements ViewRenderer).
// userMenu adapts the Identity contract into the component's plain props. The
// component must not know the contract: it lives below layout in the graph, and
// typing Identity there would invert the dependency.
//
// A fresh instance per call on purpose — the shell renders one for the header
// and one for the drawer, and a single *UserMenu cannot have two parents.
// brand builds the header's leading slot from the Brand contract. The mark is
// an <img> when the contract returns a URL and the default glyph otherwise —
// the avatar's exact treatment, mirrored.
// Clickable: the brand doubles as the "go home" control, the standard
// meaning of tapping a logo. It fires the same navigation Init runs when the
// app opens with no hash, on the header and the drawer alike — one behaviour
// for one glyph, wherever it is drawn.
func (p *Platform) brand() *Element {
	slot := Div().Set(clsBrand.AsAttr())
	if url := p.Brand.BrandMark(); url != "" {
		slot.Child(NewElement("img").
			Set(clsBrandMark.AsAttr()).
			Attr("src", url).
			Attr("alt", p.Brand.BrandName()).
			Attr("loading", "lazy"))
	} else {
		slot.Child(IconBrand.Render(string(clsBrandMark)))
	}
	slot.Child(Span().Set(clsBrandName.AsAttr()).Text(p.Brand.BrandName()))
	slot.OnClick(func(Event) { p.goHome() })
	return slot
}

// appMenu is the header's leading control: the app's mark, which opens a small
// panel with the app's name and version. It follows the user menu's native
// <details> pattern (no script; a viewport backdrop closes it with a tap
// anywhere else). The drawer's own brand, on a phone, stays the "go home"
// control: there the name is the head of the panel, so nothing competes.
func (p *Platform) appMenu() *Element {
	name := p.AppName
	if p.Brand != nil {
		name = p.Brand.BrandName()
	}

	trigger := Summary().Set(clsBrandTrigger.AsAttr()).Attr("aria-label", name)
	if p.Brand != nil {
		if url := p.Brand.BrandMark(); url != "" {
			trigger.Child(NewElement("img").
				Set(clsBrandMark.AsAttr()).
				Attr("src", url).
				Attr("alt", "").
				Attr("loading", "lazy"))
		} else {
			trigger.Child(IconBrand.Render(string(clsBrandMark)))
		}
	} else {
		trigger.Child(IconBrand.Render(string(clsBrandMark)))
	}

	panel := Div().Set(clsBrandPanel.AsAttr()).
		Child(Span().Set(clsBrandPanelName.AsAttr()).Text(name))
	if p.AppVersion != "" {
		panel.Child(Span().Set(clsBrandPanelVer.AsAttr()).Text(p.AppVersion))
	}

	// Mirrored into a signal because <details> writes its native `open` onto
	// itself, which the backdrop (a sibling) cannot select on — the user menu's
	// mechanism, same reason.
	backdrop := Div().Set(clsBrandBackdrop.AsAttr()).
		BindStateFunc(widget.Open, func() bool { return p.brandOpen.Get() })

	menu := Details().Set(clsBrandMenu.AsAttr()).
		Child(trigger).
		Child(backdrop).
		Child(panel)

	// The header toggles the phone drawer on any tap: a tap that opens this
	// panel must not also open the drawer.
	menu.OnClick(func(e Event) { e.StopPropagation() })
	menu.OnToggle(func(Event) {
		if ref, ok := menu.Ref(); ok {
			p.brandOpen.Set(ref.GetAttr("open") != "<null>")
		}
	})
	backdrop.OnClick(func(Event) {
		if ref, ok := menu.Ref(); ok {
			ref.RemoveAttr("open")
		}
		p.brandOpen.Set(false)
	})

	return menu
}

func (p *Platform) userMenu() Component {
	var actions Component
	if p.UserActions != nil {
		actions = p.UserActions()
	}
	return &usermenu.UserMenu{
		Name:     p.User.UserName(),
		Avatar:   p.User.UserAvatar(),
		Roles:    p.User.UserRoles(),
		Fallback: IconUser,
		// The theme is the shell's, not the app's: every platform gets the
		// same control in the same place, and no app has to wire one.
		Preferences: &themetoggle.ThemeToggle{},
		Actions:     actions,
	}
}

func (p *Platform) Render() *Element {
	root := Div().Set(clsRoot.AsAttr())

	// Touch swipe direction detection: enables header hide/reveal even on short
	// content or when swiping on non-scrolling areas of the mobile screen.
	var touchStartY float64
	root.OnTouchStart(func(e Event) {
		touchStartY = e.ClientY()
	})
	root.OnTouchMove(func(e Event) {
		y := e.ClientY()
		if touchStartY <= 0 || y <= 0 {
			return
		}
		delta := touchStartY - y
		if delta > 16 {
			// Swiped up -> advance into content -> stow header
			p.navStowed.Set(true)
		} else if delta < -16 {
			// Swiped down -> user intends to navigate or search -> reveal header immediately
			p.navStowed.Set(false)
		}
	})
	root.OnTouchEnd(func(Event) {
		touchStartY = 0
	})

	// ── header ───────────────────────────────────────────────────────────────
	header := Header().Set(clsHeader.AsAttr()).
		BindStateFunc(widget.Open, func() bool { return !p.navStowed.Get() || p.menuOpen.Get() }).
		OnClick(func(Event) { p.menuOpen.Toggle() })

	// The leading slot: the app's mark. The app's NAME and version are not
	// drawn here — they compete with the module's name, which is what the
	// person is looking at — but are one tap away, in the panel the mark opens.
	if p.Brand != nil || p.AppName != "" {
		header.Child(p.appMenu())
	}

	// The module's name, large beside the app's mark on a desktop, centred on a
	// phone. It is the page's single <h1> (an accessibility contract: one h1 per
	// page); inside the content only <h2> sections exist.
	headerTitle := Div().Set(clsHeaderTitle.AsAttr()).
		Child(H1().Set(clsHeaderTitleText.AsAttr()).BindTextFunc(p.activeLabel))
	header.Child(headerTitle)

	// Chassis singleton: one per page by construction, so a stable global id is
	// the intended use of ID() (see webtyp/dom, "Element ids are owned by dom").
	msgSlot := Div().Set(clsMsgSlot.AsAttr()).ID("pd-msg-slot").
		BindChildren(p.notifications)
	// Because elementToHTML/SSR doesn't process "children" bindings, initial nodes must be manually added
	for _, n := range p.notifications.Get() {
		msgSlot.Child(n)
	}
	header.Child(msgSlot)

	right := Div().Set(clsHeaderRight.AsAttr()).
		OnClick(func(e Event) { e.StopPropagation() })
	if p.Status != nil {
		right.Child(p.Status())
	}
	if p.User != nil {
		right.Child(p.userMenu())
	}

	header.Child(right)

	hamburger := Button().Set(clsHamburger.AsAttr()).
		Attr("aria-label", "Menu").
		BindChildren(p.navIcon).
		ID("pd-hamburger-btn")
	header.Child(hamburger)

	root.Child(header)

	// ── mobile message stack (mobile only) ─────────────────────────────────
	// The toasts' phone home, docked to the top-end.
	msgStack := Div().Set(clsMsgStack.AsAttr())
	msgSlotMobile := Div().Set(clsMsgSlotMobile.AsAttr()).
		BindChildren(p.notificationsMobile)
	for _, n := range p.notificationsMobile.Get() {
		msgSlotMobile.Child(n)
	}
	msgStack.Child(msgSlotMobile)
	// Added as the root's LAST child below — see that site for why DOM order
	// matters there (the stack ties with the drawer/overlay at --z-dropdown).

	// ── nav overlay backdrop (mobile) ────────────────────────────────────────
	overlay := Div().Set(clsNavOverlay.AsAttr()).
		BindStateFunc(widget.Open, func() bool { return p.menuOpen.Get() })
	overlay.OnClick(func(Event) {
		p.menuOpen.Set(false)
	})
	root.Child(overlay)

	// ── body (Sidebar wrapping stage and nav) ────────────────────────────────
	body := Div().Set(clsBody.AsAttr())

	// ── main stage ───────────────────────────────────────────────────────────
	stage := Main().Set(clsStage.AsAttr())

	for _, m := range p.Modules {
		m := m
		if gm, ok := m.(GroupModule); ok {
			for _, child := range gm.Children() {
				child := child
				cID := child.ModelName()
				if !p.isViewable(cID) {
					continue
				}
				panel := Section().Set(clsPanel.AsAttr()).
					Key(cID).
					Attr("data-id", cID).
					BindStateFunc(widget.Current, func() bool { return p.active.Get() == cID })

				if v := child.View(); v != nil {
					panel.Child(v)
				}
				stage.Child(panel)
			}
			continue
		}

		id := m.ModelName()
		if !p.isViewable(id) {
			continue
		}
		panel := Section().Set(clsPanel.AsAttr()).
			Key(id).
			Attr("data-id", id).
			BindStateFunc(widget.Current, func() bool { return p.active.Get() == id })

		if v := m.View(); v != nil {
			panel.Child(v)
		}
		stage.Child(panel)
	}
	body.Child(stage)

	// ── navigation menu (rail) ───────────────────────────────────────────────
	nav := Nav().Set(clsMenu.AsAttr()).
		BindStateFunc(widget.Open, func() bool { return p.menuOpen.Get() })
	navbar := Ul().Set(clsNavbar.AsAttr())

	for _, m := range p.Modules {
		m := m
		id := m.ModelName()
		if !p.isViewable(id) {
			continue
		}

		if gm, ok := m.(GroupModule); ok {
			children := gm.Children()
			navItem := Li().Set(clsNavItem.AsAttr())

			parentLink := A("#"+id).Set(clsNavParent.AsAttr()).
				Attr("data-id", id).
				BindStateFunc(widget.Open, func() bool {
					return p.expandedGroup.Get() == id
				}).
				BindStateFunc(widget.Current, func() bool {
					activeID := p.active.Get()
					for _, ch := range children {
						if ch.ModelName() == activeID {
							return true
						}
					}
					return false
				})

			slotNodes := p.getGroupIconSlot(id)
			if slotNodes == nil {
				slotNodes = NewNodes()
			}
			activeIconSlot := Span().Set(clsNavActiveIcon.AsAttr()).BindChildren(slotNodes)
			parentLink.Child(activeIconSlot)

			if icon := gm.Icon(); icon != "" {
				parentLink.Child(icon.Render(string(clsNavGroupIcon)))
			}
			parentLink.Child(Span().Set(clsLinkText.AsAttr()).Text(gm.Label()))
			parentLink.Child(iconChevronDown.Render(string(clsChevron)))

			parentLink.OnClick(func(Event) {
				if p.expandedGroup.Get() == id {
					p.expandedGroup.Set("")
				} else {
					p.expandedGroup.Set(id)
					for _, ch := range children {
						if p.isViewable(ch.ModelName()) {
							p.Activate(ch.ModelName())
							break
						}
					}
				}
			})

			navItem.Child(parentLink)

			submenu := Ul().Set(clsSubmenu.AsAttr()).
				BindStateFunc(widget.Open, func() bool {
					return p.expandedGroup.Get() == id
				})

			for _, ch := range children {
				ch := ch
				chID := ch.ModelName()
				if !p.isViewable(chID) {
					continue
				}

				chLink := A("#"+chID).Set(clsSubnavLink.AsAttr()).
					Attr("data-id", chID).
					BindStateFunc(widget.Current, func() bool { return p.active.Get() == chID })

				if icon := ch.Icon(); icon != "" {
					chLink.Child(icon.Render(string(clsSubnavIcon)))
				}
				chLink.Child(Span().Set(clsLinkText.AsAttr()).Text(ch.Label()))
				if b, ok := ch.(Badged); ok {
					if badge := b.Badge(); badge != nil {
						chLink.Child(badge.Render())
					}
				}

				submenu.Child(Li().Set(clsSubnavItem.AsAttr()).Child(chLink))
			}

			navItem.Child(submenu)
			navbar.Child(navItem)
			continue
		}

		link := A("#"+id).Set(clsNavLink.AsAttr()).
			Attr("data-id", id).
			BindStateFunc(widget.Current, func() bool { return p.active.Get() == id })

		if icon := m.Icon(); icon != "" {
			link.Child(icon.Render(string(ClsNavIcon)))
		}
		link.Child(Span().Set(clsLinkText.AsAttr()).Text(m.Label()))
		if b, ok := m.(Badged); ok {
			if badge := b.Badge(); badge != nil {
				link.Child(badge.Render())
			}
		}

		navbar.Child(Li().Set(clsNavItem.AsAttr()).Child(link))
	}

	// The drawer's head. On a phone this is where the brand rides — there is
	// no header to hold it — and on a wide screen the header carries it while
	// this panel floats out above the rail on hover.
	// One panel holding the head and the navbar. On a wide screen the whole
	// panel is what floats out over the content on hover — if the head expanded
	// inside the rail's flow instead, the rail would widen and push the stage.
	drawerPanel := Div().Set(clsDrawerPanel.AsAttr())

	// The brand leads the drawer and doubles as its "go home" control — tap
	// the mark, land back on the root module, the same gesture the header
	// offers. Wrapped in its own OnlyOn(Mobile) part: on a wide screen the
	// header already carries the brand, top-left, at all times — showing it
	// again in the hover-revealed rail panel would say the app's own name
	// twice on screen at once. AppName is the fallback for a platform with no
	// Brand: plain text, since there is no mark to make tappable. Brand and
	// AppName never render together — a phone showing both would say the
	// app's own name twice in the same panel — and AppName deliberately has
	// no CueWithin to reveal it on a wide screen either: the drawer opens
	// wholesale on a phone so a line of text costs nothing, while in the
	// collapsed rail it would have to materialise when the rail expands and
	// push everything under it down.
	if p.Brand != nil {
		drawerPanel.Child(Div().Set(clsDrawerBrand.AsAttr()).Child(p.brand()))
	} else if p.AppName != "" {
		drawerPanel.Child(Div().Set(clsAppName.AsAttr()).Text(p.AppName))
	}

	drawerPanel.Child(navbar)

	// The identity block trails the nav, not leads it: on open, what to DO
	// matters more than who is signed in. It is a second menu instance, for
	// the phone — there is no header there to hold the first one. This costs
	// nothing now that the contract returns strings — two elements built from
	// the same facts, of which exactly one is ever visible. It was impossible
	// while identity was a Component slot: one element has one parent.
	// Wrapped in a part this package can switch off: the component's own class
	// belongs to the component, and the shell needs somewhere to say "not on a
	// wide screen, where the header already has one".
	if p.User != nil {
		drawerPanel.Child(Div().Set(clsDrawerIdentity.AsAttr()).Child(p.userMenu()))
	}

	nav.Child(drawerPanel)
	body.Child(nav)

	root.Child(body)

	// The mobile toast stack is the root's LAST child: its Docked(Viewport)
	// ties with the drawer and the nav-overlay at --z-dropdown (platformd is
	// a Menu-kind widget, so --z-toast is not reachable from here), and the
	// cascade breaks the tie by DOM order — a toast must paint above the
	// drawer's veil, not under it.
	root.Child(msgStack)

	return root
}

// toastNodes builds one toast element per raw notification. Called twice per
// refresh with different suffixes: the same notification renders into the
// desktop header slot AND the mobile stack, and the two must be distinct
// *Elements — one element cannot have two parents — so their IDs/keys differ
// by suffix and each BindChildren reconciles its own set.
//
// Accessibility: role=status (aria-live polite) announces info/success
// without stealing focus; role=alert (assertive) is for warnings and errors,
// which are announcements, not focus grabs. This is the only thing severity
// still drives — every message shares one visual class (clsMsg, styled once
// in css.go with no per-type variant) after a reported bug where the mobile
// and desktop toasts, and the delete-confirmation dialog, each showed a
// different color because each was its own declaration. role isn't a CSS
// class, so it never risked the same drift.
//
// The toast itself is the manual dismiss affordance — tapping it closes it.
// Pause-on-hover lets a long message hold the screen; focusin/focusout mirror
// it for the keyboard even though toasts are deliberately not focusable (a
// focusable toast would add tab stops that appear and vanish with each
// notification).
func (p *Platform) toastNodes(suffix string) []*Element {
	p.mu.Lock()
	defer p.mu.Unlock()

	nodes := make([]*Element, 0, len(p.rawNotifications))
	for _, n := range p.rawNotifications {
		role := "status"
		switch n.Type {
		case msgtype.Warning, msgtype.Error:
			role = "alert"
		}

		id := n.ID
		nodes = append(nodes, Div().Set(clsMsg.AsAttr()).
			Key(id+suffix).
			Attr("role", role).
			Text(n.Msg).
			OnClick(func(e Event) {
				e.StopPropagation()
				p.dismiss(id)
			}).
			OnMouseEnter(func(Event) { p.pauseToast(id) }).
			OnMouseLeave(func(Event) { p.resumeToast(id) }).
			OnFocusIn(func(Event) { p.pauseToast(id) }).
			OnFocusOut(func(Event) { p.resumeToast(id) }))
	}
	return nodes
}

// Notify queues a typed notification in both viewport slots (header on
// desktop, msg-stack on mobile). The duration is a decision, not a number:
// Auto() sizes it to the message, Persistent() leaves it until dismissed,
// For(ms) pins it. Errors are the one case that must not vanish on their own
// — hand them Persistent(), or a generous For().
func (p *Platform) Notify(t msgtype.Type, msg string, d Duration) {
	p.mu.Lock()
	// An error whose length was left to the platform stays until dismissed
	// (WCAG 2.2.1): a message that vanishes before it is read defeats the
	// report. An explicit For(ms) is the caller's decision and is respected.
	if t == msgtype.Error && d.auto {
		d = Persistent()
	}
	// The same message already on screen is not stacked again — a failing
	// auto-save would otherwise pile one identical toast per field commit.
	// A timed one gets its window restarted instead.
	for i := range p.rawNotifications {
		n := &p.rawNotifications[i]
		if n.Type != t || n.Msg != msg {
			continue
		}
		if n.timer != nil && d.millis != nil {
			ms := d.millis(msg)
			n.timer.Stop()
			id := n.ID
			n.expiryNs = time.NowNano() + time.Nano(ms)*nanosPerMilli
			n.timer = time.AfterFunc(ms, func() { p.dismiss(id) })
		}
		p.mu.Unlock()
		return
	}
	ms := 0
	if d.millis != nil {
		ms = d.millis(msg)
	}
	n := notification{
		Type: t,
		Msg:  msg,
		ID:   "pd-notification-" + p.GetID() + "-" + Sprint(int64(time.NowNano())),
	}
	if ms > 0 {
		n.expiryNs = time.NowNano() + time.Nano(ms)*nanosPerMilli
		n.timer = time.AfterFunc(ms, func() { p.dismiss(n.ID) })
	}
	p.rawNotifications = append(p.rawNotifications, n)
	p.mu.Unlock()

	p.notifications.Set(p.toastNodes(""))
	p.notificationsMobile.Set(p.toastNodes("-m"))
}

// dismiss removes the notification with the given id — from a tap on the
// toast itself, or from the auto-dismiss timer firing. The timer is stopped
// on the manual path so it cannot fire later against a removed notification
// (the auto path's timer has already fired).
func (p *Platform) dismiss(id string) {
	p.mu.Lock()
	for i, n := range p.rawNotifications {
		if n.ID == id {
			if n.timer != nil {
				n.timer.Stop()
				n.timer = nil
			}
			p.rawNotifications = append(p.rawNotifications[:i], p.rawNotifications[i+1:]...)
			p.mu.Unlock()
			p.notifications.Set(p.toastNodes(""))
			p.notificationsMobile.Set(p.toastNodes("-m"))
			return
		}
	}
	p.mu.Unlock()
}

// pauseToast stops a notification's auto-dismiss timer, remembered so the
// pause is transparent: the deadline stays fixed, and resume re-arms with
// whatever remains of it. Hovering or focusing a toast is the user saying
// "I'm still reading this" — the timer must not run through that.
func (p *Platform) pauseToast(id string) {
	p.mu.Lock()
	for i := range p.rawNotifications {
		n := &p.rawNotifications[i]
		if n.ID != id {
			continue
		}
		if n.timer != nil {
			n.timer.Stop()
			n.timer = nil
		}
		break
	}
	p.mu.Unlock()
}

// resumeToast re-arms the auto-dismiss timer for the time remaining on the
// original deadline. A persistent notification (no deadline) or an already
// dismissed one is a no-op.
func (p *Platform) resumeToast(id string) {
	p.mu.Lock()
	for i := range p.rawNotifications {
		n := &p.rawNotifications[i]
		if n.ID != id {
			continue
		}
		if n.expiryNs > 0 && n.timer == nil {
			if remaining := n.expiryNs - time.NowNano(); remaining > 0 {
				n.timer = time.AfterFunc(int(remaining/nanosPerMilli), func() { p.dismiss(n.ID) })
			}
		}
		break
	}
	p.mu.Unlock()
}

func (p *Platform) notificationCount() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return len(p.rawNotifications)
}

// Activate programmatically switches to a module by ID
// (also updates window.location.hash on wasm builds).
func (p *Platform) Activate(moduleID string) {
	if !p.isViewable(moduleID) {
		return
	}

	// If moduleID is a GroupModule itself, activate its first viewable child instead!
	for _, m := range p.Modules {
		if gm, ok := m.(GroupModule); ok && m.ModelName() == moduleID {
			for _, ch := range gm.Children() {
				if p.isViewable(ch.ModelName()) {
					p.Activate(ch.ModelName())
					return
				}
			}
		}
	}

	var foundParentID string
	var targetModule UIModule

	for _, m := range p.Modules {
		if gm, ok := m.(GroupModule); ok {
			for _, ch := range gm.Children() {
				if ch.ModelName() == moduleID {
					foundParentID = gm.ModelName()
					targetModule = ch
					break
				}
			}
		}
		if m.ModelName() == moduleID {
			targetModule = m
		}
	}

	if foundParentID != "" {
		p.expandedGroup.Set(foundParentID)
	} else {
		p.expandedGroup.Set("")
	}

	if p.active.Get() == moduleID && !p.menuOpen.Get() {
		return
	}

	p.active.Set(moduleID)
	p.menuOpen.Set(false)

	if targetModule != nil {
		targetModule.Activate()
	}

	p.navIcon.Set([]*Element{p.activeIcon().Render(string(ClsNavIcon)).Key("hmod-" + moduleID)})

	for _, s := range p.groupActiveIcons {
		if s.id == foundParentID && targetModule != nil {
			if ic := targetModule.Icon(); ic != "" {
				s.nodes.Set([]*Element{ic.Render(string(ClsNavIcon)).Key("gact-" + s.id + "-" + moduleID)})
			} else {
				s.nodes.Set(nil)
			}
		} else {
			s.nodes.Set(nil)
		}
	}

	// Cambiar de sección reinicia el cromo: el módulo nuevo empieza desde arriba y
	// con el botón a mano.
	p.lastScrollTop = 0
	p.scrollAnchor = 0
	p.navStowed.Set(false)

	// Update window hash if needed
	if GetHash() != "#"+moduleID {
		SetHash("#" + moduleID)
	}
}

// nanosPerMilli converts the milliseconds of a toast's life (time.AfterFunc) to the nanoseconds of its deadline.
const nanosPerMilli time.Nano = 1000000
