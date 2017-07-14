package composition

import "golang.org/x/net/html"

type ScriptElement *struct {
	Attrs []html.Attribute
	Text  []byte
}

func newScriptElement(attr []html.Attribute, text []byte) ScriptElement {
	return &struct {
		Attrs []html.Attribute
		Text  []byte
	}{attr, text}
}
