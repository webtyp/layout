package apiexplorer

import (
	"webtyp.com/fetch"
	"webtyp.com/lang"
	"webtyp.com/router"
	"webtyp.com/time"

	. "webtyp.com/dom"
	. "webtyp.com/fmt"
	. "webtyp.com/html"
)

func buildTryPanel(r router.RouteRecord) *Element {
	panel := Div().Set(clsTryForm.AsAttr())

	panel.OnClick(func(evt Event) {
		evt.StopPropagation()
	})

	pathParams := extractPathParams(r.Path)

	var pathInputs []*SignalString
	var argSignals []*SignalString
	var argRecords []router.ArgRecord

	pathForm := Div().Set(NameAPIExplorer.Class("path-form").AsAttr())
	for _, p := range pathParams {
		lbl := Label().Text(p)
		sig := NewString("")
		inp := Input("text").Attr("placeholder", p).Bind(sig).OnChange(func(e Event) {
			sig.Set(e.TargetValue())
		})
		pathInputs = append(pathInputs, sig)
		pathForm.Child(lbl, inp)
	}

	if len(pathParams) > 0 {
		panel.Child(pathForm)
	}

	rawBodySig := NewString("")
	if r.Method != "GET" && r.Method != "HEAD" {
		bodyForm := Div().Set(NameAPIExplorer.Class("body-form").AsAttr())
		if !r.HasArgs {
			rawBody := NewElement("textarea").Attr("placeholder", lang.Translate("Raw JSON").String()).Bind(rawBodySig).OnChange(func(e Event) {
				rawBodySig.Set(e.TargetValue())
			})
			bodyForm.Child(Label().Text(lang.Translate("Body").String()), rawBody)
		} else {
			for _, arg := range r.Args {
				argRecords = append(argRecords, arg)
				lbl := Label().Text(arg.Name)
				if arg.Required {
					lbl.Text(arg.Name + " *")
				}

				sig := NewString("")
				var inp *Element
				switch arg.Kind {
				case "int", "int64", "float64":
					inp = Input("number").Bind(sig).OnChange(func(e Event) { sig.Set(e.TargetValue()) })
				case "bool":
					inp = Input("checkbox").OnChange(func(e Event) {
						if e.TargetChecked() {
							sig.Set("true")
						} else {
							sig.Set("false")
						}
					})
				default:
					inp = Input("text").Bind(sig).OnChange(func(e Event) { sig.Set(e.TargetValue()) })
				}
				argSignals = append(argSignals, sig)
				bodyForm.Child(lbl, inp)
			}
		}
		panel.Child(bodyForm)
	}

	resultPanel := Div().Set(NameAPIExplorer.Class("result-panel").AsAttr())
	statusMsg := NewString("")
	bodyMsg := NewString("")
	resultPanel.Child(Div().BindText(statusMsg), Pre().BindText(bodyMsg))

	btnText := lang.Translate("Send").String()
	if r.Method != "GET" {
		btnText = lang.Translate("Execute").String() + " " + r.Method
	}

	sendBtn := Button().Text(btnText)
	confirmState := false

	sendBtn.OnClick(func(evt Event) {
		evt.StopPropagation()

		if r.Method != "GET" && !confirmState {
			sendBtn.Text(lang.Translate("Confirm").String() + " " + btnText + "?")
			confirmState = true

			time.AfterFunc(3000, func() {
				if confirmState {
					confirmState = false
					sendBtn.Text(btnText)
				}
			})
			return
		}

		confirmState = false
		sendBtn.Text(btnText)

		finalPath := r.Path
		allValid := true
		for i, p := range pathParams {
			val := pathInputs[i].Get()
			if val == "" {
				allValid = false
				break
			}
			finalPath = ReplaceAll(finalPath, "{"+p+"}", val)
		}

		if !allValid {
			statusMsg.Set(lang.Translate("Path parameters cannot be empty").String())
			return
		}

		statusMsg.Set(lang.Translate("Sending...").String())
		bodyMsg.Set("")

		var req *fetch.Request
		switch r.Method {
		case "POST":
			req = fetch.Post(finalPath)
		case "PUT":
			req = fetch.Put(finalPath)
		case "DELETE":
			req = fetch.Delete(finalPath)
		default:
			req = fetch.Get(finalPath)
		}

		if r.Method != "GET" && r.Method != "HEAD" {
			req.ContentTypeJSON()
			if !r.HasArgs {
				req.Body([]byte(rawBodySig.Get()))
			} else if r.HasArgs {
				req.Body([]byte(buildJSON(argRecords, argSignals)))
			}
		}

		req.Send(func(resp *fetch.Response, err error) {
			if err != nil {
				statusMsg.Set(lang.Translate("Error").String() + ": " + err.Error())
				return
			}

			statusText := Sprintf("%d", resp.Status)
			if resp.Status == 403 {
				rolesStr := lang.Translate("nobody").String()
				if len(r.Roles) > 0 {
					rolesStr = joinStrings(r.Roles, ", ")
				}
				statusText += " (" + lang.Translate("requires").String() + " " + r.Resource + ":" + r.Action + ", " + lang.Translate("held by").String() + ": " + rolesStr + ")"
			}
			statusMsg.Set(statusText)

			bodyMsg.Set(string(resp.Body()))
		})
	})

	panel.Child(sendBtn)
	panel.Child(resultPanel)

	return panel
}

func extractPathParams(path string) []string {
	var params []string
	inParam := false
	var current string
	for _, r := range path {
		if r == '{' {
			inParam = true
			current = ""
		} else if r == '}' {
			if inParam {
				params = append(params, current)
			}
			inParam = false
		} else if inParam {
			current += string(r)
		}
	}
	return params
}

func buildJSON(args []router.ArgRecord, signals []*SignalString) string {
	jsonStr := "{"
	for i, arg := range args {
		if i > 0 {
			jsonStr += ","
		}
		jsonStr += Sprintf(`"%s":`, arg.Name)

		val := signals[i].Get()
		if arg.Kind == "bool" {
			if val == "true" {
				jsonStr += "true"
			} else {
				jsonStr += "false"
			}
		} else {
			if arg.Kind == "int" || arg.Kind == "int64" || arg.Kind == "float64" {
				if val == "" {
					jsonStr += "0"
				} else {
					jsonStr += val
				}
			} else {
				escaped := ReplaceAll(val, `"`, `\"`)
				escaped = ReplaceAll(escaped, "\n", `\n`)
				jsonStr += Sprintf(`"%s"`, escaped)
			}
		}
	}
	jsonStr += "}"
	return jsonStr
}
