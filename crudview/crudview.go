package crudview

import (
	. "webtyp.com/dom"
	. "webtyp.com/html"

	"webtyp.com/components/modaldialog"
	"webtyp.com/components/targetlist"
	"webtyp.com/fmt"
	"webtyp.com/fmt/lang"
	"webtyp.com/form"
	"webtyp.com/icons/pencil"
	"webtyp.com/icons/plus"
	"webtyp.com/icons/trash"
	"webtyp.com/icons/undo"
	"webtyp.com/layout/rightpanel"
	"webtyp.com/view"
	"webtyp.com/widget"
)

const NameCrudView widget.Name = "crudview"

var (
	clsBoxContent          = NameCrudView.Class("fields")
	clsBtnCrud             = NameCrudView.Class("action")
	clsBtnCrudDelete       = NameCrudView.Class("action-delete")
	clsBtnCrudEdit         = NameCrudView.Class("action-edit")
	clsBtnCrudDeleteIcon   = NameCrudView.Class("action-delete-icon")
	clsBtnCrudEditIcon     = NameCrudView.Class("action-edit-icon")
	clsFooter              = NameCrudView.Class("footer")
	clsBtnCrudIconHidden   = NameCrudView.Class("action-hidden")
	clsListaBox            = NameCrudView.Class("list")
	clsDelConfirmBody      = NameCrudView.Class("delconfirm-body")
	clsDelConfirmActions   = NameCrudView.Class("delconfirm-actions")
	clsDelConfirmBtn       = NameCrudView.Class("delconfirm-btn")
	clsDelConfirmBtnDanger = NameCrudView.Class("delconfirm-btn-danger")
	clsDelConfirmMount     = NameCrudView.Class("delconfirm-mount")
)

const (
	// The single toggle button swaps between these two glyphs reactively — see
	// the "toggle" block in Render(). All four crud glyphs come from
	// webtyp/icons; trash/pencil are also drawn by the list marks
	// (targetlist/targetdate) from the same package, so the buttons and the
	// rows they act on can never draw a different shape.
	iconCrudNew    = plus.Ref // "+"  — nothing selected
	iconCrudCancel = undo.Ref // "↺" — a row is selected / draft in progress (undo)
)

// crudMode is the single source of truth for which chrome the footer shows and
// what a tap on a row means. A closed set, in one signal: three booleans would
// make "menu open AND selecting" representable, and it is not a state — it is
// a bug waiting for a race between two clicks.
type crudMode string

const (
	modeNormal   crudMode = ""       // "+" plus the two bulk entries, 🗑 and ✏
	modeDeleting crudMode = "delete" // ↺ + 🗑 N, rows show checks
	modeEditing  crudMode = "edit"   // ↺ + ✏ N, rows show checks
)

// ListView is the row-rendering half of a CrudView. targetlist.TargetList
// (plain label rows) and targetdate.TargetDate (a leading date/time badge
// instead of a plain label — see view.Item's LeadTop/Main/Bottom) both
// satisfy it, so Config.List can swap one for the other without CrudView
// ever knowing which it got.
type ListView interface {
	Component
	SetItems(items []view.Item)
	Items() []view.Item
	Count() int
	// Selection mode — targetlist and targetdate both implement it by
	// assembling the components/listselect lego piece.
	SetSelectMode(on bool)
	// Danger tone — the list paints checked rows red instead of blue while
	// armed. CrudView arms it exactly in delete mode; a plain list that
	// never hears it keeps its single Accent language.
	SetDanger(on bool)
	CheckedIDs() []string
	OnCheckedChange(fn func(n int))
}

func (v *CrudView) WidgetName() widget.Name { return NameCrudView }
func (v *CrudView) WidgetKind() widget.Kind { return widget.Disclosure }

type CrudView struct {
	Element // value embed — NEVER *dom.Element

	Title     string
	Form      Component // what Render paints (may stay nil in standalone mode)
	Presenter view.Presenter

	// Filter is the control that narrows the list — a searchbar.SearchBar, a
	// calendar, a select. nil paints no controls band at all. If it implements
	// widget.Filterable, crudview wires it to the list filter in Init.
	//
	// The type is Component, not widget.Filterable, on purpose: a control with
	// no live output (a static legend, a chip strip that navigates elsewhere) is
	// a legitimate occupant and must not be forced to implement a callback it
	// would never fire.
	Filter Component

	// Context is a control rendered in the title band (below the H1), for a
	// context/scope selector: "which professional / which área". nil paints no
	// band. Set by New from Config.Context.
	//
	// Deliberately NOT auto-wired to the list filter: changing the context
	// re-scopes the data (the composing module calls its presenter +
	// Reload), it does not write a search term. crudview only renders it.
	Context Component

	// List builds the row-rendering widget, given the Selected signal and the
	// OnSelect/OnDelete callbacks CrudView owns — Init calls it once. Set by
	// New from Config.List; nil there resolves to a targetlist.TargetList
	// factory, same "ergonomic default, not a decision imposed" as Filter.
	List func(selected *SignalString, onSelect func(view.Item)) ListView

	// Additive user hooks — called AFTER the built-in behavior. Assigning them
	// can never disable list→form fill, save or delete wiring.
	OnSelect func(it view.Item)
	OnNew    func()
	OnSaved  func(err error)
	// OnDeleted fires after both the single-record delete (row → ⋮ → Eliminar)
	// and the bulk delete (selection mode → 🗑 N). ids mirrors view.Deleter's
	// own variadic shape rather than staying a lone string: one string could
	// never have named 3 deleted records, and a second callback just for the
	// bulk case would have been two events for what is, to a host, the same
	// thing happening at a different N.
	OnDeleted func(ids []string, err error)
	// OnUpdated fires after a bulk field patch (selection mode → ✏ N). Same
	// [ids]+err shape as OnDeleted, for the same reason. There is no
	// single-record counterpart: editing one loaded record persists through
	// OnSaved (autoSaveAction → Saver.Save), by design — see the master plan's
	// "edición de un registro vs. edición masiva".
	OnUpdated func(ids []string, err error)
	OnCancel  func()

	// internal
	form          *form.Form             // typed handle set by New; nil when standalone
	list          ListView               // owns the row rendering + ⋮ menu
	panel         *rightpanel.RightPanel // the skeleton this controller fills
	confirmDelete *modaldialog.ModalDialog
	selected      *SignalString
	search        *SignalString
	canDelete     *SignalBool
	deleteID      *SignalString // record pending confirmation (⋮ → Eliminar)
	deleteLabel   *SignalString // its label, for the confirmation message
	composing     *SignalBool   // "+" was pressed and nothing saved/cancelled yet (see active())
	mode          *SignalString // single source of truth for mode (crudMode)
	// hasChecked and hasEdits are what the commit buttons READ. The count
	// itself has no display here — listselect.Header's "k / N" strip is the
	// one place it's shown; a second count on the footer button (a
	// countbadge bubble) duplicated it for no reason.
	hasChecked *SignalBool
	hasEdits   *SignalBool
	// hasRows mirrors "the list currently shows at least one record", set in
	// filter() on every reload/search. The footer's 🗑/✏ read it: on an empty
	// list there is nothing to delete or bulk-edit, so only "+" shows.
	hasRows *SignalBool
	// hasMultiRows is hasRows' sibling for the ONE affordance that is
	// meaningless below N=2: the bulk-edit ✏. Delete (🗑) reads hasRows —
	// it is variadic and a single-row delete is a real case; ✏ patches a
	// SET by one delta, so a lone row would just be a worse single edit
	// (which the form already does — master plan §4).
	hasMultiRows *SignalBool
}

// active reports whether the toggle button should show "↺" (cancel/undo):
// either an existing row is selected, or the user is mid-composing a new one
// (pressed "+", hasn't saved or cancelled). Nothing else should read
// v.selected/v.composing directly to decide the icon/branch — this is the
// single place that combines them.
func (v *CrudView) active() bool {
	return v.selected.Get() != "" || v.composing.Get()
}

func (v *CrudView) Init(ctx Ctx) {
	v.selected = NewString("")
	v.search = NewString("")
	v.canDelete = NewBool(false)
	v.deleteID = NewString("")
	v.deleteLabel = NewString("")
	v.composing = NewBool(false)
	v.mode = NewString(string(modeNormal))
	v.hasChecked = NewBool(false)
	v.hasEdits = NewBool(false)
	v.hasRows = NewBool(false)
	v.hasMultiRows = NewBool(false)

	// The list owns row rendering + the ⋮ menu and shares the selected signal
	// so its highlight follows the form. New resolves Config.List's default
	// before this runs, but Init must not assume it always went through
	// New — direct struct construction (&CrudView{...}; v.Init(ctx)) is an
	// established pattern in this package's own tests, same as Filter being
	// left nil is fine. Default here mirrors New's own default exactly.
	buildList := v.List
	if buildList == nil {
		buildList = func(selected *SignalString, onSelect func(view.Item)) ListView {
			return &targetlist.TargetList{Selected: selected, OnSelect: onSelect}
		}
	}
	v.list = buildList(v.selected, v.selectAction)
	v.list.OnCheckedChange(func(n int) {
		v.hasChecked.Set(n > 0)
	})

	// The filter control reports terms; crudview owns what a term means.
	// Assigned here, not in Render, so it survives a re-render and so a host
	// that never renders — the conformance driver — still filters.
	if src, ok := v.Filter.(widget.Filterable); ok {
		src.OnFilterChange(func(term string) {
			v.search.Set(term)
			v.filter()
		})
	}

	// Delete confirmation modal — Content is built once; its message reacts to
	// v.deleteLabel so the same instance is reused across every ⋮ → Eliminar.
	// No "×": the body carries Cancelar and Eliminar, and a third way out says
	// nothing Cancelar does not. A destructive confirmation wants exactly two
	// exits, both of them explicit.
	v.confirmDelete = &modaldialog.ModalDialog{
		Title:     lang.Translate("Confirm").String(),
		HideClose: true,
		Content:   v.renderDeleteConfirm(),
	}

	if v.Presenter != nil {
		if err := v.Reload(); err != nil {
			Log(err.Error())
		}
	}
}

// renderDeleteConfirm builds the confirmation modal's body: a message naming
// the record, plus Cancelar/Eliminar actions. Built once in Init and reused —
// see the confirmDelete field.
func (v *CrudView) renderDeleteConfirm() *Element {
	// The question names the action and the record. "¿Desea continuar?" would
	// read as safe and be signed without thought — what makes someone stop is
	// seeing the verb and the name of the thing it is about to happen to. Each
	// word is an independent dictionary key — joined with spaces at read time —
	// so consumers reuse the same entries ("Delete" is already the confirm
	// button's key).
	msg := P().BindTextFunc(func() string {
		tmpl := lang.Translate("Delete", "%s?", "This", "action", "cannot", "be", "undone.").String()
		return fmt.Sprintf(tmpl, "«"+v.deleteLabel.Get()+"»")
	})

	cancel := Button().Set(clsDelConfirmBtn.AsAttr()).Text(lang.Translate("Cancel").String()).
		On("click", func(Event) { v.confirmDelete.Close() })

	confirm := Button().Set(clsDelConfirmBtn.AsAttr(), clsDelConfirmBtnDanger.AsAttr()).Text(lang.Translate("Delete").String()).
		On("click", func(Event) { v.confirmDeleteAction() })

	actions := Div().Set(clsDelConfirmActions.AsAttr()).Child(cancel, confirm)

	// Classed so the gap between the message and the actions is declared, and
	// declared as the same step the dialog puts between its header and its
	// body — otherwise the question sits far from the title and right on top of
	// the buttons.
	return Div().Set(clsDelConfirmBody.AsAttr()).Child(msg, actions)
}

// editAction: load the record and focus the first field so the user can start
// typing right away. Kept as a method although no UI path calls it anymore —
// the ⋮ menu lost Editar with the lock it existed to undo — because the
// view/conformance suite's "edit_focuses_first_field" clause drives it
// directly (see conformance_test.go). Focus is synchronous, on purpose — see
// newAction below for why it must never be deferred.
func (v *CrudView) editAction(id string) {
	v.selectAction(view.Item{ID: id})
	if v.form != nil {
		v.form.Focus()
	}
}

// deleteRequest (⋮ → Eliminar): opens the confirmation modal instead of
// deleting immediately. confirmDeleteAction performs the actual delete.
func (v *CrudView) deleteRequest(id string) {
	if _, ok := v.Presenter.(view.Deleter); !ok || id == "" {
		return
	}
	label := id
	for _, it := range v.list.Items() {
		if it.ID == id {
			label = it.Label
			break
		}
	}
	v.deleteID.Set(id)
	v.deleteLabel.Set(label)
	v.confirmDelete.Open()
}

// confirmDeleteAction: the modal's "Eliminar" button. Deletes the record(s)
// pending confirmation and closes the modal.
func (v *CrudView) confirmDeleteAction() {
	if v.mode.Get() == string(modeDeleting) {
		if deleter, ok := v.Presenter.(view.Deleter); ok && v.list != nil {
			ids := v.list.CheckedIDs()
			if len(ids) > 0 {
				err := deleter.Delete(ids...)
				if err == nil {
					v.setMode(modeNormal)
					_ = v.Reload()
				} else {
					Log(err.Error())
				}
				if v.OnDeleted != nil {
					v.OnDeleted(ids, err)
				}
			}
		}
	} else {
		id := v.deleteID.Get()
		if deleter, ok := v.Presenter.(view.Deleter); ok && id != "" {
			v.deleteAction(deleter, id)
		}
	}
	v.confirmDelete.Close()
}

func (v *CrudView) Reload() error {
	if v.Presenter == nil {
		return nil
	}
	if err := v.Presenter.Reload(); err != nil {
		return err
	}
	v.filter()
	return nil
}

// selectAction: card click / driver Select. Selecting an existing row loads it
// into the form — immediately editable, because the form lock is gone: there
// is no Save button, every field commit persists on its own (autoSaveAction),
// so nothing needs protecting from edits. On mobile (horizontal scroll-snap
// strip) it also snaps the viewport to the form panel; a no-op on desktop,
// where both columns are already visible.
func (v *CrudView) selectAction(it view.Item) {
	v.selected.Set(it.ID)
	v.canDelete.Set(it.ID != "")
	if v.list != nil {
	}
	if it.ID != "" {
		v.composing.Set(false) // picking an existing row abandons any new-record draft
	}
	rec := v.Presenter.Select(it.ID)
	if v.form != nil {
		_ = v.form.LoadValues(rec) // nil record → LoadValues resets; not an error
	}
	if it.ID != "" {
		if v.panel != nil {
			v.panel.ShowMain()
		}
	}
	if v.OnSelect != nil {
		v.OnSelect(it)
	}
}

// newAction: the toggle button in its "+" state (nothing selected). Idempotent —
// safe to call even when already in the "new" state. Focuses the first field —
// standard behavior, see the view/conformance "new_focuses_first_field" clause.
// Marks composing=true: the button switches to "↺" for the rest of the draft
// (until saved-and-cancelled or explicitly undone), the same as selecting an
// existing row does — see active().
func (v *CrudView) newAction() {
	v.selected.Set("")
	v.canDelete.Set(false)
	if v.list != nil {
	}
	v.Presenter.Deselect()
	if v.form != nil {
		v.form.Reset()
		v.form.Focus()
	}
	v.composing.Set(true)
	// Focus before ShowMain, both synchronous, is deliberate: iOS Safari only
	// opens the keyboard for a focus() that happens inside the call stack of
	// the user gesture, so deferring it by any amount (tried at several
	// delays) moves DOM focus but leaves the keyboard shut. Order matters too
	// — Focus must not be the last scroll-affecting thing to run, which is why
	// ShowMain follows it and dom's Focus suppresses the browser's own
	// focus-scroll (see elementWasm.Focus): otherwise the two scrolls race and
	// the strip lands between snap points.
	if v.panel != nil {
		v.panel.ShowMain()
	}
	if v.OnNew != nil {
		v.OnNew()
	}
}

// undoAction: the toggle button in its "↺" state (active() — a row is
// selected, or a new-record draft is in progress). Undoes everything —
// deselects, clears the form, drops the draft — and returns the button to
// "+". Deliberately does NOT call Form.Focus(): cancelling must leave nothing
// selected/focused (standard behavior, see the view/conformance
// "cancel_clears_focus" clause) — unlike newAction/editAction, which enter an
// editable state and focus the first field on purpose.
func (v *CrudView) undoAction() {
	v.clearSelection()
	if v.OnCancel != nil {
		v.OnCancel()
	}
}

// clearSelection puts the controller back in its resting state: nothing
// selected, no draft, no delete armed, an empty form, and — on a phone, where
// the two columns are a scroll-snap strip — the list back on screen instead of
// a form with nothing left in it.
//
// Two callers, deliberately: undoAction (the user pressed "↺") and
// dropSelectionOutOfScope (the filter moved and took the record with it). They
// are the same state change; only undoAction is a user-initiated cancel, so
// only undoAction fires OnCancel.
func (v *CrudView) setMode(m crudMode) {
	if crudMode(v.mode.Get()) == m {
		return
	}
	v.mode.Set(string(m))
	if m == modeDeleting || m == modeEditing {
		if v.list != nil {
			v.list.SetSelectMode(true)
			// Red is the delete mode's own tone: checked rows (and the
			// delete button, via its Invalid state) lean toward danger.
			// Edit mode keeps the plain Accent marks — red there would lie.
			v.list.SetDanger(m == modeDeleting)
		}
		if m == modeEditing && v.form != nil {
			v.form.Reset()
			v.form.MarkPristine()
			v.hasEdits.Set(false) // a blank form has nothing to apply yet
		}
	} else {
		if v.list != nil {
			v.list.SetSelectMode(false)
			v.list.SetDanger(false)
		}
		v.hasEdits.Set(false)
	}
}

func (v *CrudView) clearSelection() {
	v.setMode(modeNormal)
	v.selected.Set("")
	v.canDelete.Set(false)
	v.composing.Set(false)
	if v.list != nil {
	}
	if v.Presenter != nil {
		v.Presenter.Deselect()
	}
	if v.form != nil {
		v.form.Reset() // also clears the tracked FocusedFieldID()
	}
	if v.panel != nil {
		v.panel.ShowAside()
	}
}

// toggleAction: the single crud button's click handler. "↺"→undo whenever
// active() (a row is selected, or a new-record draft is in progress);
// "+"→create otherwise — but only when the presenter can actually save. A
// view without view.Saver has nothing to create, so the "+" is a no-op there
// (and the button renders disabled in that state — see Render); this guard is
// the belt for a driver that calls the action directly.
func (v *CrudView) toggleAction() {
	if v.active() {
		v.undoAction()
		return
	}
	if _, ok := v.Presenter.(view.Saver); !ok {
		return
	}
	v.newAction()
}

// saveAction persists the current form values. Only reachable when saver != nil.
// A silent no-op when the form isn't dirty (see Form.IsDirty) — moving focus
// through a field without changing it must never persist or fire OnSaved
// (a host's "Guardado" toast on an untouched field would be pure noise, and
// on mobile — where the module fills the screen — actively in the way).
func (v *CrudView) saveAction(saver view.Saver) {
	if !v.form.IsDirty() {
		return
	}
	err := v.form.Validate()
	if err == nil {
		record := v.Presenter.Record()
		if err = v.form.SyncValues(record); err == nil {
			if err = saver.Save(record); err == nil {
				v.form.MarkPristine() // a later untouched commit isn't dirty again
				if v.composing.Get() {
					// A new-record draft just saved successfully: the draft is
					// done, return to the "+" ready state. Not undoAction — that
					// fires OnCancel ("Cancelado"), wrong after a real save; this
					// is a silent reset, same shape as undoAction minus the
					// callback. Only for composing: editing an EXISTING record
					// (selected≠"", composing=false) must NOT reset here, or
					// every auto-save on blur would kick the user out of the
					// record they're still editing.
					v.composing.Set(false)
					v.selected.Set("")
					v.canDelete.Set(false)
					v.Presenter.Deselect()
					v.form.Reset()
				}
				_ = v.Reload()
			}
		}
	}
	if err != nil {
		Log(err.Error())
	}
	if v.OnSaved != nil {
		v.OnSaved(err)
	}
}

// autoSaveAction is the Form.OnFieldChange hook: every field commit (blur/change)
// persists immediately — there is no explicit Save button. A no-op when the
// presenter cannot save (standalone/read-only views) or when in bulk editing mode.
//
// This is where editing ONE loaded record persists — through view.Saver, with
// the whole record. It deliberately does NOT route through view.Updater even
// when the presenter has it: Updater / the ✏ button is the BULK path (patch N
// rows with a delta). The split is the master plan's, §4 — "editar un registro
// vs. editar en lote".
func (v *CrudView) autoSaveAction() {
	if v.mode != nil && v.mode.Get() == string(modeEditing) {
		// Bulk edit: nothing is persisted per field. What a commit DOES change
		// is whether there is anything to apply, so the apply button can be
		// dead until the user actually edits something. Tracked here because
		// this is the one hook the form already fires on every commit.
		if v.form != nil && v.hasEdits != nil {
			v.hasEdits.Set(len(v.form.DirtyFields()) > 0)
		}
		return
	}
	if saver, ok := v.Presenter.(view.Saver); ok && v.form != nil {
		v.saveAction(saver)
	}
}

// deleteAction: delete button / driver Delete. Only reachable when deleter != nil.
func (v *CrudView) deleteAction(deleter view.Deleter, id string) {
	err := deleter.Delete(id)
	if err == nil {
		v.selected.Set("")
		v.canDelete.Set(false)
		_ = v.Reload()
	} else {
		Log(err.Error())
	}
	if v.OnDeleted != nil {
		v.OnDeleted([]string{id}, err)
	}
}

// filter repopulates the list for the current term and then enforces the one
// invariant a list-detail controller cannot let slip: what the form holds must
// be something the list still shows.
func (v *CrudView) filter() {
	v.list.SetItems(v.Presenter.Filter(v.search.Get()))
	if v.hasRows != nil {
		v.hasRows.Set(v.list.Count() > 0)
	}
	if v.hasMultiRows != nil {
		v.hasMultiRows.Set(v.list.Count() > 1)
	}
	v.dropSelectionOutOfScope()
}

// deleteEntryAction: the 🗑 button's click handler. A loaded record deletes
// directly — deleteRequest opens that record's own confirmation, no
// selection mode. With nothing loaded it enters delete mode instead; an
// unsaved draft (composing, no id) does nothing, and the button is disabled
// for it anyway (see Render).
func (v *CrudView) deleteEntryAction() {
	if v.mode.Get() == string(modeNormal) {
		if id := v.selected.Get(); id != "" {
			v.deleteRequest(id)
		} else if !v.active() {
			v.setMode(modeDeleting)
		}
	} else if v.mode.Get() == string(modeDeleting) {
		v.bulkDeleteAction()
	}
}

// bulkDeleteAction commits the marked rows. One call, not a loop: view.Deleter
// is variadic precisely so the whole batch is one statement, and a loop would
// bring back the half-applied failure the plural contract exists to prevent.
func (v *CrudView) bulkDeleteAction() {
	if v.list == nil {
		return
	}
	ids := v.list.CheckedIDs()
	if len(ids) == 0 {
		return
	}
	var label string
	if len(ids) == 1 {
		label = ids[0]
		for _, it := range v.list.Items() {
			if it.ID == ids[0] {
				label = it.Label
				break
			}
		}
	} else {
		// "3 records" / "3 registros", never a bare "3": the modal reads
		// "Delete %s?", and a lone number there names nothing.
		label = fmt.Sprintf("%d %s", len(ids), lang.Translate("records").String())
	}
	v.deleteID.Set("")
	v.deleteLabel.Set(label)
	v.confirmDelete.Open()
}

// bulkEditAction patches the marked rows with ONLY the fields the user
// touched. Sending whole records instead would silently revert every column
// someone else changed since this client last reloaded — see the master plan's
// "por qué ids + delta".
func (v *CrudView) bulkEditAction() {
	if v.list == nil || v.form == nil {
		return
	}
	ids := v.list.CheckedIDs()
	if len(ids) == 0 {
		return
	}
	fields := v.form.DirtyFields()
	if len(fields) == 0 {
		// Unreachable through the UI — the apply button is disabled until
		// hasEdits is true — so this is a belt, not a mode. Loud rather than
		// silent: an empty change set here means a caller bypassed the button.
		Log("crudview: bulk edit with no changed fields")
		return
	}
	updater, ok := v.Presenter.(view.Updater)
	if !ok {
		return
	}
	record := v.Presenter.Record()
	if err := v.form.SyncValues(record); err != nil {
		Log(err.Error())
		return
	}
	err := updater.Update(ids, record, fields)
	if err == nil {
		v.setMode(modeNormal)
		v.form.Reset()
		_ = v.Reload()
	} else {
		Log(err.Error())
	}
	if v.OnUpdated != nil {
		v.OnUpdated(ids, err)
	}
}

// dropSelectionOutOfScope clears the selection when the freshly filtered list
// no longer contains it. A picker that changes the SCOPE (a patient selector
// narrowing to that patient's records) leaves the previously loaded record
// orphaned: still in the form, still editable, still saveable — against a list
// that no longer shows it. That is a data-corruption path, not a cosmetic one.
//
// Unconditional, with no config flag to turn it off: a form holding a record
// outside the visible list is an illegal state, not a preference. A composing
// draft goes too — it belongs to the scope that just left.
func (v *CrudView) dropSelectionOutOfScope() {
	id := v.selected.Get()
	if id == "" && !v.composing.Get() {
		return // nothing loaded; nothing to orphan
	}
	if id != "" && v.list != nil {
		for _, it := range v.list.Items() {
			if it.ID == id {
				return // still in scope
			}
		}
	}
	v.clearSelection()
}

func (v *CrudView) Render() *Element {
	hasSource := v.Presenter != nil

	// The form's inset. The card around it is rightpanel's own `rp__article`
	// (the skeleton's Part("article")) — crudview paints no frame, so the form
	// area is exactly one layer, not a card nested in a card.
	boxContent := Div().Set(clsBoxContent.AsAttr())
	if v.Form != nil {
		boxContent.Child(v.Form)
	}

	v.panel = &rightpanel.RightPanel{
		Title:        v.Title,
		Article:      boxContent,
		HeadControls: v.Context,
	}

	if hasSource {
		// The list — its own inset card inside the aside's content band.
		v.panel.Aside = Div().Set(clsListaBox.AsAttr()).
			Child(v.list)

		// The filter is the consumer's control. crudview supplies no card
		// around it: rightpanel's controls band already keeps its size, and a
		// second frame around a control that has one reads as a box in a box.
		v.panel.AsideControls = v.Filter

		// The create action ("+") needs view.Saver; 🗑 needs view.Deleter; ✏
		// (bulk patch) needs view.Updater. 🗑/✏ already gate their own render
		// below. The toggle button always renders — it is also the "↺" that
		// leaves selection mode — but its "+" state goes disabled without a
		// Saver (see its disabled bind). Editing ONE loaded record is not
		// gated: it rides Saver via autoSaveAction.
		_, hasSaver := v.Presenter.(view.Saver)

		toggle := Button().Set(clsBtnCrud.AsAttr()).
			Attr("name", "cv-crudtoggle").
			BindStateFunc(widget.Open, func() bool { return v.mode.Get() != string(modeNormal) || v.active() }).
			// Dead in normal mode when there is nothing to create: a Deleter-
			// or Updater-only view still needs this button as the "↺" that
			// leaves selection mode, so it stays — just disabled while it
			// would only mean "+".
			BindAttrBool("disabled", DeriveBool(func() bool {
				return !hasSaver && v.mode.Get() == string(modeNormal) && !v.active()
			})).
			Child(
				iconCrudNew.Render(string(NameCrudView.Class("action-new"))).
					BindStateFunc(widget.Open, func() bool { return v.mode.Get() != string(modeNormal) || v.active() }),
				iconCrudCancel.Render(string(NameCrudView.Class("action-cancel"))).
					BindStateFunc(widget.Open, func() bool { return v.mode.Get() != string(modeNormal) || v.active() }),
			)
		toggle.On("click", func(Event) {
			if v.mode.Get() != string(modeNormal) {
				v.setMode(modeNormal)
			} else {
				v.toggleAction()
			}
		})

		// The footer's Open state is the delete-mode tone: WhenWithin
		// paints the delete button red only while it holds (see css.go).
		// Footer order is [delete][toggle][edit]: delete sits at the leading
		// edge of add, every button sharing the same box (see css.go).
		footer := Div().Set(clsFooter.AsAttr()).
			BindStateFunc(widget.Open, func() bool { return v.mode.Get() == string(modeDeleting) })

		if _, ok := v.Presenter.(view.Deleter); ok {
			btnDelete := Button().Set(clsBtnCrudDelete.AsAttr()).
				Attr("name", "cv-cruddelete").
				// Shown while deleting (the commit button) OR, in normal mode,
				// only when there is a record to act on: rows exist and we are
				// not mid-drafting a new one. Composing collapses the footer to
				// just "↺" (nothing to delete yet, the only exits are finish or
				// cancel), and an empty list shows only "+" — same spirit as
				// the two selection modes, which already hide the button they
				// are not.
				BindStateFunc(widget.Open, func() bool {
					if v.mode.Get() == string(modeDeleting) {
						return true
					}
					return v.mode.Get() == string(modeNormal) && !v.composing.Get() && v.hasRows.Get()
				}).
				// In delete mode the button leans red with the rows it will
				// act on (see css.go); everywhere else it stays Primary.
				// The tone rides the FOOTER's Open state, not the button's:
				// this sheet's Disclosure kind admits no other state, and
				// the button's own Open already means "visible", which is
				// true in normal mode too.
				BindAttrBool("disabled", DeriveBool(func() bool {
					if v.mode.Get() == string(modeNormal) {
						// A loaded record deletes directly (its own confirm
						// dialog); only an unsaved draft — nothing to name
						// in the dialog — keeps the button dead.
						return v.composing.Get()
					}
					return !v.hasChecked.Get()
				})).
				// No count badge here: listselect.Header's "k / N" strip
				// already shows how many are marked, right above the list —
				// a second count bubble on this button duplicated it.
				Child(trash.Ref.Render(string(clsBtnCrudDeleteIcon)))
			btnDelete.On("click", func(Event) { v.deleteEntryAction() })
			footer.Child(btnDelete)
		}

		footer.Child(toggle)

		if _, ok := v.Presenter.(view.Updater); ok {
			btnEdit := Button().Set(clsBtnCrudEdit.AsAttr()).
				Attr("name", "cv-crudedit").
				// Shown while editing (the apply button) OR, in normal mode,
				// only when a bulk edit could start: rows exist and nothing is
				// loaded or being drafted (active()), since entering selection
				// mode would discard the form. Empty list → hidden (nothing to
				// patch); composing/loaded → hidden (was a dead disabled
				// button).
				BindStateFunc(widget.Open, func() bool {
					if v.mode.Get() == string(modeEditing) {
						return true
					}
					// ✏ is always a bulk patch — hidden below 2 rows (a single
					// row is edited by tapping it + the form, master plan §4).
					return v.mode.Get() == string(modeNormal) && !v.active() && v.hasMultiRows.Get()
				}).
				BindAttrBool("disabled", DeriveBool(func() bool {
					if v.mode.Get() == string(modeNormal) {
						return v.active()
					}
					// Marked rows AND something to write. Without the second
					// half, pressing apply with an untouched form called
					// nothing and reported nothing — a dead button that looked
					// alive. An unpressable button needs no error message.
					return !v.hasChecked.Get() || !v.hasEdits.Get()
				})).
				// No count badge here either — see btnDelete above.
				Child(pencil.Ref.Render(string(clsBtnCrudEditIcon)))
			btnEdit.On("click", func(Event) {
				if v.mode.Get() == string(modeNormal) {
					if !v.active() {
						v.setMode(modeEditing)
					}
				} else if v.mode.Get() == string(modeEditing) {
					v.bulkEditAction()
				}
			})
			footer.Child(btnEdit)
		}

		v.panel.AsideFooter = footer
	}

	root := v.panel.Render()

	if hasSource {
		// v.confirmDelete's Show() wraps its content in a bare, class-less div
		// even while hidden. As an un-placed child of the frame's grid it got
		// auto-placed into an implicit second row and doubled the bottom gutter.
		// clsDelConfirmMount is position:fixed, which removes it from grid item
		// participation entirely, regardless of visibility.
		root.Child(Div().Set(clsDelConfirmMount.AsAttr()).Child(v.confirmDelete))
	}

	return root
}
