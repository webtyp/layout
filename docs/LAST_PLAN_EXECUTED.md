---
PLAN: "fix: replace inline style binding in login message with hidden attribute"
EXECUTOR: jules
REVIEWER: none
---

> Read AGENTS.md at the repo root before any change.
> **Finish by yourself:** when the delivery check passes, commit, push and open the pull request. Never ask whether to proceed or finalize — nobody answers, and a session that waits ends without your work.

# Plan: `layout` sin atributo `style=""` en `login` (CSP `style-src 'self'`)

Master: `SECURITY_OBSERVABILITY_MASTER_PLAN.md` (D3, O3b-6).

## Tools

```bash
export GOPRIVATE=webtyp.com/*,github.com/veltylabs/*
export GOFLAGS=-mod=mod
go install webtyp.com/devflow/cmd/gotest@latest
export PATH="$PATH:$(go env GOPATH)/bin"
```

## Ecosystem search

- `grep -rn '"style"' layout/ --include='*.go'` muestra exactamente 1 sitio: `layout/login/login.go:117` (`BindAttrFunc("style", ...)` que retorna `"display:none"` cuando `l.Message.Get() == ""`).
- `dom` ya provee `BindAttrBoolFunc(name string, fn func() bool) *Element` (`dom/element.go:349`).
- **Decisión:** usar el atributo estándar HTML `hidden` con `BindAttrBoolFunc("hidden", ...)` (100% permitido bajo `style-src 'self'` sin estilos en línea).

## Design gate

No cambia ninguna firma pública de `webtyp.com/layout`.

## APIs used

| Symbol | Exact signature | Tag | File |
|---|---|---|---|
| `(*dom.Element).BindAttrBoolFunc` | `func (b *Element) BindAttrBoolFunc(name string, fn func() bool) *Element` | `v0.13.29` | `webtyp.com/dom` |

If a signature in the published tag differs from this table, stop and name the symbol in the PR; never adapt it by guessing.

## Reglas de entrega y anti-footguns

- `layout/login/login.go` compila a WASM con TinyGo.
- Hacer commit después de cada etapa. No usar `git reset`, `git restore`, `git checkout -- <file>` ni `git stash`.

---

## Etapa 1: Prueba (TDD)

En `login/login_test.go` (o `tests/login_test.go`, registrado en `docs/TESTS.md` bajo `| Test | Rule |`):
- `TestLogin_MessageNoInlineStyle`: renderizar `Login` con `Message` vacío y con texto; verificar que el HTML no contiene `style=` y que con mensaje vacío incluye el atributo `hidden`.

## Etapa 2: Reemplazar `BindAttrFunc("style", ...)` en `login/login.go`

En `login/login.go:115-123`, reemplazar:
```go
		msg := Div().Set(clsMessage.AsAttr()).
			Attr("role", "alert").
			BindAttrFunc("style", func() string {
				if l.Message.Get() == "" {
					return "display:none"
				}
				return ""
			}).
			BindText(l.Message)
```
por:
```go
		msg := Div().Set(clsMessage.AsAttr()).
			Attr("role", "alert").
			BindAttrBoolFunc("hidden", func() bool {
				return l.Message.Get() == ""
			}).
			BindText(l.Message)
```

## Etapa 3: Verificación y entrega

1. Verificar con grep: `grep -rn '"style"' . --include='*.go' | grep -v '_test.go'` → vacío.
2. Ejecutar `gotest`.

## Tabla de etapas

| Etapa | Archivos | Qué |
|---|---|---|
| 1 | `login/login_test.go`, `docs/TESTS.md` | Prueba TDD sin `style=` |
| 2 | `login/login.go` | Reemplazo por `BindAttrBoolFunc("hidden", ...)` |
| 3 | `docs/TESTS.md` | Verificación grep y `gotest` verde |
