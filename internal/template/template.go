// Package template renders Go text/template data into Typst source files.
package template

import (
	"bytes"
	"fmt"
	"os"
	texttemplate "text/template"
)

// TypstText escapes markup characters for insertion into a Typst text context.
// It does not make values safe for Typst code, paths, URLs, or expressions.
func TypstText(value any) string {
	s := fmt.Sprint(value)
	var out bytes.Buffer
	for _, r := range s {
		switch r {
		case '\\', '#', '*', '_', '[', ']', '$', '<', '>', '`':
			out.WriteRune('\\')
		}
		out.WriteRune(r)
	}
	return out.String()
}

// RenderFile executes a template file into destination using Go's text/template.
func RenderFile(source, destination string, data any) error {
	content, err := os.ReadFile(source)
	if err != nil {
		return fmt.Errorf("read template %q: %w", source, err)
	}
	t, err := texttemplate.New(source).
		Option("missingkey=error").
		Funcs(texttemplate.FuncMap{"typstText": TypstText}).
		Parse(string(content))
	if err != nil {
		return fmt.Errorf("parse template %q: %w", source, err)
	}
	var rendered bytes.Buffer
	if err := t.Execute(&rendered, data); err != nil {
		return fmt.Errorf("execute template %q: %w", source, err)
	}
	if err := os.WriteFile(destination, rendered.Bytes(), 0o600); err != nil {
		return fmt.Errorf("write rendered template: %w", err)
	}
	return nil
}
