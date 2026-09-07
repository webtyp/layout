---
PLAN: "feat(crudview): Context slot in the title band"
TAG: v0.3.0
EXECUTOR: local
REVIEWER: none
---

# PLAN — `layout/crudview` slot `Context` (Etapa E del `DEMO_AGENDA_MASTER_PLAN`)

Orquestador: `webtyp/docs/DEMO_AGENDA_MASTER_PLAN.md` §4.2, §7 fila E.
(Copia local: `/home/cesar/Dev/Project/webtyp/docs/DEMO_AGENDA_MASTER_PLAN.md`.)

## Objetivo

`crudview` necesita un lugar donde el consumidor ponga un control de **contexto**
(en la demo: selects de área + médico; en el legado Pa100T era el dropdown de
profesional junto al título). El slot de filtro (`Config.Filter`) ya está tomado
por el calendario y va al `AsideControls`. El slot de contexto va en la **banda
de título**, que `rightpanel.RightPanel` ya expone como `HeadControls`
("e.g. select with search") — así que esto NO toca `rightpanel`.

## Decisión de alcance

El `Context` es un **slot de render, no un filtro auto-cableado**. Razón: si
tanto `Filter` (calendario, `widget.Filterable`) como `Context` (médico) fueran
`Filterable`, ambos escribirían el mismo término y `Presenter.Filter(term)` —
que es de un solo término por contrato — no podría representar "médico X + día
Y". Elegir médico **re-acota los datos** (es scope, no un término de búsqueda):
eso lo maneja el módulo que compone el `Context` llamando a su presenter +
`crudview.Reload()` (Etapa F). `crudview` solo lo pinta.

Si en el futuro se quiere auto-cableo, es un añadido separado con su propia
decisión sobre términos compuestos.

## Cambios

### `crud.go` — `Config`

```go
type Config struct {
    // … campos actuales …

    // Context es un control que se pinta en la banda de título (bajo el H1,
    // sobre el artículo), para un selector de contexto/scope: p. ej. "qué
    // profesional / qué área". Opcional: nil no pinta la banda.
    //
    // A diferencia de Filter, Context NO se auto-cablea al filtro de la lista
    // aunque satisfaga widget.Filterable: cambiar el contexto re-acota los
    // datos (el módulo que lo compone llama a su presenter + CrudView.Reload),
    // no filtra un término. crudview solo lo renderiza.
    Context dom.Component
}
```

### `crudview.go`

- Campo `Context Component` en `CrudView` (doc que espeja el de `Config`).
- `New(cfg)`: `v.Context = cfg.Context` (sin default — nil = sin banda).
- `Render()`: `v.panel.HeadControls = v.Context` (hoy `HeadControls` no lo setea
  nadie desde crudview; `rightpanel.Render` ya lo pinta como
  `Div(clsHeadControls).Child(r.HeadControls)` bajo el `titleRow`). Verificar que
  no colisione con ningún uso actual de `HeadControls` (grep: hoy nadie lo usa
  desde crudview).
- **Sin** wiring en `Init`.

### `Reload()` ya es público

Confirmar que `CrudView.Reload() error` es exportado (lo es: `crudview.go:315`) —
Etapa F lo llama desde el handler del select de médico. Si hiciera falta un
`Reload` que además limpie la selección, evaluar en Etapa F; no anticiparlo aquí.

## CSS

`rightpanel/css.go` ya tiene `partHeadControls`/`clsHeadControls` con estilo
(banda bajo el título). Verificar que se ve bien con 1–2 `<select>` dentro; si
necesita un pequeño ajuste (gap, wrap en mobile) hacerlo en `rightpanel/css.go`,
mínimo. No crear CSS nuevo en `crudview` para esto salvo que sea imprescindible.

## Tests (`crudview/`, `gotest`)

- `TestContext_RendersInHeadBand` — `crudview.New(Config{…, Context: stub})` →
  el HTML tiene el `stub` dentro de `.rp__head-controls` (o la clase real), bajo
  el `<h1>`.
- `TestContext_NilNoBand` — `Context: nil` → no se emite la banda; snapshot/HTML
  idéntico a hoy (sin regresión para `item_catalog`, `devices`, etc.).
- `TestContext_NotWiredToFilter` — un `Context` que también es `widget.Filterable`
  NO recibe `OnFilterChange` de crudview (contar llamadas: 0).
- Conformance test existente sigue verde.

## Criterios de aceptación

- `gotest ./...` verde en `layout`.
- `GOOS=js GOARCH=wasm go build ./...` OK.
- Consumidores actuales sin cambios de comportamiento (`Context` nil por defecto).
- `crudview` README / `docs/ARCHITECTURE.md` documentan el slot y la decisión de
  "no auto-cableo".

## Fuera de alcance

- Los selects de área/médico en sí (Etapa F, `app-demo`).
- Términos de filtro compuestos.
