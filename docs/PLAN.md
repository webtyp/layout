---
PLAN: "feat(login): ranura de mensaje de error tipada — la app deja de recurrir a alert()"
EXECUTOR: jules
REVIEWER: none
---

> This plan is dispatched via the CodeJob workflow. See skill: agents-workflow.

# PLAN — `login.Login` no tiene dónde mostrar un error

## 1. El defecto

`login/login.go` define la pantalla previa a la autenticación:

```go
type Login struct {
	Element
	Title    string
	Subtitle string
	Form     Component
	LogoMark string
}
```

No hay ranura para el resultado del envío. Cuando el servidor rechaza unas
credenciales, el componente no tiene dónde decirlo: `Form` es un `Component`
opaco que este paquete solo renderiza, y `form.Form` tampoco muestra el error
de submit por su cuenta (`done(err)` solo reactiva el botón).

**Lo que eso provoca aguas abajo.** Una app en producción terminó escribiendo
esto en su composition root:

```go
js.Global().Call("alert", "Acceso no autorizado en este dispositivo")
```

Un `alert()` del navegador en la pantalla de login: bloquea el hilo, no se puede
estilar, no respeta el tema, y en móvil aparece con el nombre del host. Se
escribió con un comentario que decía «puente explícito y temporal hasta que
exista esa ranura». Esta es esa ranura.

Es el caso literal del
[CONSTRUCTION_HARNESS](https://github.com/webtyp/app-releases/blob/main/docs/CONSTRUCTION_HARNESS.md):
*«A missing contract at a boundary is a defect in the library, not in the
consumer.»* El error cruza de la app al componente y no hay tipo que lo nombre.

## 2. El cambio

### 2.1 `login/login.go` — la parte y el campo

Añadir la parte al bloque de constantes que ya existe, **al final**, para no
alterar las que ya están:

```go
const (
	PartCard     = widget.Part("card")
	PartHeader   = widget.Part("header")
	PartTitle    = widget.Part("title")
	PartSubtitle = widget.Part("subtitle")
	PartMark     = widget.Part("mark")
	PartMessage  = widget.Part("message") // el resultado del último intento de envío
)
```

y su clase, junto a las otras:

```go
var (
	...
	clsMessage = NameLogin.Class(PartMessage)
)
```

Actualizar el comentario de `NameLogin`, que enumera las clases producidas,
para incluir `"login__message"`.

El campo, después de `Form`:

```go
	// Message es el resultado del último intento de envío: vacío mientras no
	// haya ninguno, el texto a mostrar cuando el servidor rechaza. Es una
	// señal y no un string porque llega DESPUÉS del render — la respuesta del
	// servidor es asíncrona y el árbol ya está montado.
	//
	// Opcional: nil no renderiza la ranura en absoluto. Una pantalla cuyo
	// formulario no puede fallar no paga por un nodo que nunca se llena.
	//
	// Este paquete no decide el texto. Quién falló y por qué lo sabe la app;
	// aquí solo se muestra.
	Message *SignalString
```

### 2.2 `Render()` — dónde va

Entre el header y el formulario. Un error se lee **antes** de volver a
intentar, no después de rellenar otra vez:

```go
func (l *Login) Render() *Element {
	header := Div().Set(clsHeader.AsAttr())
	if l.LogoMark != "" {
		header.Child(NewElement("img").NoCloseTag().Set(clsMark.AsAttr()).Attr("src", l.LogoMark).Attr("alt", ""))
	}
	header.Child(H1().Set(clsTitle.AsAttr()).Text(l.Title))
	if l.Subtitle != "" {
		header.Child(P().Set(clsSubtitle.AsAttr()).Text(l.Subtitle))
	}

	card := Div().Set(clsCard.AsAttr()).Child(header)

	// Sin señal no hay ranura: nada que mostrar, ningún nodo que mantener.
	if l.Message != nil {
		// role="alert" hace que el lector de pantalla lo anuncie cuando
		// aparece, que es justo el momento en que importa. Show mantiene el
		// nodo montado y solo alterna display, así que el anuncio se dispara
		// por el cambio de contenido: el texto entra en un nodo que el
		// usuario no estaba leyendo.
		msg := Div().Set(clsMessage.AsAttr()).
			Attr("role", "alert").
			BindText(l.Message)
		card.Child(Show(DeriveBool(func() bool { return l.Message.Get() != "" }), msg))
	}

	card.Child(l.Form)

	root := Div().Set(clsLogin.AsAttr())
	root.Child(card)
	return root
}
```

Firmas que ya existen y hay que usar tal cual — no inventar variantes:

| Símbolo | Paquete | Firma |
|---|---|---|
| `Show` | `webtyp.com/dom` | `func Show(cond *SignalBool, content Component) *Element` |
| `DeriveBool` | `webtyp.com/dom` | `func DeriveBool(compute func() bool) *SignalBool` |
| `BindText` | `webtyp.com/dom` | `func (b *Element) BindText(s *SignalString) *Element` |
| `NewString` | `webtyp.com/dom` | `func NewString(v string) *SignalString` |

`dom` ya está importado con `.` en este archivo, así que se escriben sin
prefijo, como el resto del paquete.

### 2.3 `login/css.go` — la receta

Añadir al final de la cadena de `RenderCSS`, antes de `.Stylesheet()`, usando
**solo tokens que ya existen** en `webtyp.com/widget/style`:

```go
		// DangerWash, no Danger: un error de login es una nota dentro de la
		// tarjeta, no un bloque de alarma que compita con el formulario que
		// la persona tiene que volver a usar. El wash tiñe el fondo y deja
		// el texto legible sobre la superficie Inset de la tarjeta.
		Part(PartMessage,
			style.As(style.DangerWash),
			style.Pad(style.Space3),
			style.Round(style.RadiusMd),
			style.FontSize(style.TextSm),
		).
```

Los cuatro tokens (`DangerWash`, `Space3`, `RadiusMd`, `TextSm`) están
verificados como existentes. **Si alguno no compila, no lo sustituyas por CSS a
mano ni por un valor literal:** para y repórtalo en el PR — una receta que falta
es un defecto de `widget/style`, no algo que este paquete resuelva por su
cuenta.

## 3. Los tests

`login/login_test.go` y `login/login_css_test.go` ya existen; sigue su estilo.

**En `login_test.go`:**

1. `Login` con `Message == nil` → el árbol renderizado **no** contiene la clase
   `login__message`. (La ranura opcional no cuesta un nodo.)
2. `Login` con `Message = NewString("")` → el nodo existe pero está oculto.
3. `Login` con `Message = NewString("Acceso denegado")` → el nodo está visible
   y su texto es ese.
4. El nodo del mensaje aparece **antes** que `Form` en el orden de hijos de la
   tarjeta.
5. El nodo lleva `role="alert"`.

**En `login_css_test.go`:** la hoja generada declara una regla para
`login__message`.

**Test de forma-consumidor** (la regla que hace publicable la API): un test que
construya un `Login` con un `Form` real, simule un envío fallido llamando
`Message.Set("...")` **después** del render, y compruebe que el texto aparece.
Ese es el camino exacto que recorre una app; si resulta incómodo de escribir,
la API está mal y hay que decirlo en el PR en vez de forzar el test.

## 4. Criterios de aceptación

- [ ] `gotest ./...` verde en el repo.
- [ ] Los cinco tests de `login_test.go` y el de `css` existen y pasan.
- [ ] Existe el test de forma-consumidor de §3.
- [ ] `Message == nil` no renderiza ningún nodo de mensaje.
- [ ] `PartMessage` es la **última** constante del bloque; las demás no cambian de valor.
- [ ] El comentario de `NameLogin` enumera `"login__message"`.
- [ ] Ningún token nuevo en `widget/style`; ningún CSS escrito a mano.

## 5. Fuera de alcance

- No tocar `form.Form` ni añadirle presentación de errores. Un formulario dentro
  de un CRUD tiene otro sitio donde mostrarlos; unificarlo es otro plan.
- No añadir variantes de severidad (`info`, `warning`). Una ranura, un uso.
  Cuando aparezca el segundo caso real se tipa entonces, no antes.
- No tocar `platformd` ni ningún otro paquete de este repo.
