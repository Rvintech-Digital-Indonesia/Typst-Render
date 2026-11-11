//go:build integration

package typstrender

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestIntegrationCompileBasicAndInvoice(t *testing.T) {
	if _, err := exec.LookPath("typst"); err != nil {
		t.Skip("Typst CLI is not installed")
	}
	root := "."
	for _, name := range []string{"basic", "invoice"} {
		t.Run(name, func(t *testing.T) {
			out := filepath.Join(t.TempDir(), name+".pdf")
			var data any
			if name == "invoice" {
				data = map[string]any{
					"Customer": map[string]any{"Name": `Rvin #1 [ACME] _** $<> ~ @invoice`, "Address": "Mojokerto"},
					"Items":    []map[string]any{{"Name": "Hosting", "Price": 150000}},
					"Paid":     true,
				}
			}
			path, err := RenderFile(context.Background(), filepath.Join(root, "testdata/templates", name+".typ"), out, data)
			if err != nil {
				t.Fatal(err)
			}
			if info, err := os.Stat(path); err != nil || info.Size() == 0 {
				t.Fatalf("missing or empty PDF: %v", err)
			}
		})
	}
}

func TestIntegrationRelativeImportsAndAssets(t *testing.T) {
	if _, err := exec.LookPath("typst"); err != nil {
		t.Skip("Typst CLI is not installed")
	}
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "shared.typ"), []byte("= Included content\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "tiny.svg"), []byte(`<svg xmlns="http://www.w3.org/2000/svg" width="10" height="10"><circle cx="5" cy="5" r="4"/></svg>`), 0o600); err != nil {
		t.Fatal(err)
	}
	source := filepath.Join(dir, "relative.typ")
	if err := os.WriteFile(source, []byte(`#include "shared.typ"
#image("tiny.svg", width: 10pt)
`), 0o600); err != nil {
		t.Fatal(err)
	}
	out, err := RenderFile(context.Background(), source, filepath.Join(dir, "output.pdf"), nil)
	if err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(out)
	if err != nil || info.Size() == 0 {
		t.Fatalf("missing or empty PDF: %v", err)
	}
}
