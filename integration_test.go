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
					"Customer": map[string]any{"Name": "Rvin", "Address": "Mojokerto"},
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
