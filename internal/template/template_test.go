package template

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRenderFile(t *testing.T) {
	tests := []struct {
		name string
		text string
		data any
		want string
	}{
		{"map and range", `{{ .Customer.Name }}:{{ range .Items }}{{ . }};{{ end }}`, map[string]any{"Customer": map[string]any{"Name": "Rvin"}, "Items": []string{"a", "b"}}, "Rvin:a;b;"},
		{"nested struct", `{{ .Customer.Name }}`, struct {
			Customer struct{ Name string }
		}{Customer: struct{ Name string }{Name: "Nested"}}, "Nested"},
		{"struct and conditional", `{{ if .Active }}{{ .Name }}{{ else }}off{{ end }}`, struct {
			Name   string
			Active bool
		}{"A", true}, "A"},
		{"array", `{{ range . }}{{ . }}{{ end }}`, [3]int{1, 2, 3}, "123"},
		{"nil data", `static`, nil, "static"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			source, destination := filepath.Join(dir, "in.typ"), filepath.Join(dir, "out.typ")
			if err := os.WriteFile(source, []byte(tt.text), 0o600); err != nil {
				t.Fatal(err)
			}
			if err := RenderFile(source, destination, tt.data); err != nil {
				t.Fatal(err)
			}
			got, err := os.ReadFile(destination)
			if err != nil {
				t.Fatal(err)
			}
			if string(got) != tt.want {
				t.Fatalf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestRenderFileErrors(t *testing.T) {
	tests := []struct {
		name, source string
		data         any
	}{
		{"syntax", `{{ if }}`, nil},
		{"missing map key", `{{ .missing }}`, map[string]any{}},
		{"execution", `{{ .Name }}`, struct{}{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			src, dst := filepath.Join(dir, "in.typ"), filepath.Join(dir, "out.typ")
			if err := os.WriteFile(src, []byte(tt.source), 0o600); err != nil {
				t.Fatal(err)
			}
			if err := RenderFile(src, dst, tt.data); err == nil {
				t.Fatal("expected error")
			}
			if _, err := os.Stat(dst); !os.IsNotExist(err) {
				t.Fatalf("unexpected destination state: %v", err)
			}
		})
	}
}

func TestTypstText(t *testing.T) {
	for _, char := range []string{"\\", "#", "*", "_", "[", "]", "$", "<", ">", "`", "~", "@", "=", "-", "+", "/"} {
		if got, want := TypstText(char), "\\"+char; got != want {
			t.Errorf("TypstText(%q) = %q, want %q", char, got, want)
		}
	}
}

func BenchmarkTypstText(b *testing.B) {
	value := "Invoice #2026: [hosting] *service* $150,000"
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = TypstText(value)
	}
}
