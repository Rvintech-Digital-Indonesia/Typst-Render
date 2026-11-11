package typstrender

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func fakeTypst(t *testing.T, script string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "typst-fake")
	if err := os.WriteFile(path, []byte("#!/bin/sh\n"+script+"\n"), 0o700); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestRenderFileStagesAndReplacesOutput(t *testing.T) {
	dir := t.TempDir()
	input := filepath.Join(dir, "invoice.typ")
	original := []byte("= {{ .Name }}")
	if err := os.WriteFile(input, original, 0o600); err != nil {
		t.Fatal(err)
	}
	output := filepath.Join(dir, "nested", "invoice.pdf")
	if err := os.MkdirAll(filepath.Dir(output), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(output, []byte("old"), 0o600); err != nil {
		t.Fatal(err)
	}
	bin := fakeTypst(t, `printf '%s' '%PDF-1.7 generated' > "$3"`)
	got, err := New(Options{BinaryPath: bin}).RenderFile(context.Background(), input, output, map[string]any{"Name": "Rvin"})
	if err != nil {
		t.Fatal(err)
	}
	if !filepath.IsAbs(got) || filepath.Clean(got) != got {
		t.Fatalf("path not clean and absolute: %q", got)
	}
	pdf, err := os.ReadFile(output)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(string(pdf), "%PDF-") {
		t.Fatalf("unexpected output %q", pdf)
	}
	unchanged, err := os.ReadFile(input)
	if err != nil {
		t.Fatal(err)
	}
	if string(unchanged) != string(original) {
		t.Fatal("source template was changed")
	}
}

func TestRenderFailuresPreserveExistingOutputAndCleanup(t *testing.T) {
	dir := t.TempDir()
	input, output := filepath.Join(dir, "in.typ"), filepath.Join(dir, "out.pdf")
	if err := os.WriteFile(input, []byte("hello"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(output, []byte("previous"), 0o600); err != nil {
		t.Fatal(err)
	}
	bin := fakeTypst(t, `echo diagnostic >&2; exit 1`)
	_, err := New(Options{BinaryPath: bin}).RenderFile(context.Background(), input, output, nil)
	if !errors.Is(err, ErrCompileFailed) || !strings.Contains(err.Error(), "diagnostic") {
		t.Fatalf("unexpected error: %v", err)
	}
	got, err := os.ReadFile(output)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "previous" {
		t.Fatalf("existing output changed: %q", got)
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), ".typstrender-") {
			t.Fatalf("staging file left behind: %s", e.Name())
		}
	}
}

func TestRenderCancellationAndTimeout(t *testing.T) {
	dir := t.TempDir()
	input := filepath.Join(dir, "in.typ")
	if err := os.WriteFile(input, []byte("ok"), 0o600); err != nil {
		t.Fatal(err)
	}
	bin := fakeTypst(t, `sleep 2; printf '%s' '%PDF-1.7' > "$3"`)
	t.Run("timeout", func(t *testing.T) {
		_, err := New(Options{BinaryPath: bin, Timeout: 30 * time.Millisecond}).RenderFile(context.Background(), input, filepath.Join(dir, "timeout.pdf"), nil)
		if !errors.Is(err, context.DeadlineExceeded) {
			t.Fatalf("expected deadline error, got %v", err)
		}
	})
	t.Run("cancellation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		_, err := New(Options{BinaryPath: bin}).RenderFile(ctx, input, filepath.Join(dir, "cancel.pdf"), nil)
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("expected cancellation error, got %v", err)
		}
	})
}

func TestRenderValidationErrors(t *testing.T) {
	dir := t.TempDir()
	input := filepath.Join(dir, "in.typ")
	if err := os.WriteFile(input, []byte("hello"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := RenderFile(context.Background(), filepath.Join(dir, "missing.typ"), filepath.Join(dir, "out.pdf"), nil); !errors.Is(err, ErrTemplateNotFound) {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, err := RenderFile(context.Background(), filepath.Join(dir, "in.txt"), filepath.Join(dir, "out.pdf"), nil); !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, err := RenderFile(context.Background(), filepath.Join(dir, "in.typ"), filepath.Join(dir, "out.txt"), nil); !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, err := New(Options{BinaryPath: filepath.Join(dir, "missing-bin")}).RenderFile(context.Background(), input, filepath.Join(dir, "out.pdf"), nil); !errors.Is(err, ErrBinaryNotFound) {
		t.Fatalf("expected missing binary error, got %v", err)
	}
	if _, err := RenderFile(nil, input, filepath.Join(dir, "out.pdf"), nil); !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected nil context error, got %v", err)
	}
	if !IsTemplateInput("invoice.typ") || IsTemplateInput("invoice.typst") || IsTemplateInput("invoice.pdf") {
		t.Fatal("unexpected template extension classification")
	}
}

func TestInvalidCompilerOutputIsRejected(t *testing.T) {
	dir := t.TempDir()
	input := filepath.Join(dir, "in.typ")
	if err := os.WriteFile(input, []byte("source"), 0o600); err != nil {
		t.Fatal(err)
	}
	output := filepath.Join(dir, "out.pdf")
	bin := fakeTypst(t, `printf 'not-pdf' > "$3"`)
	if _, err := New(Options{BinaryPath: bin}).RenderFile(context.Background(), input, output, nil); !errors.Is(err, ErrOutputInvalid) {
		t.Fatalf("expected invalid PDF error, got %v", err)
	}
	if _, err := os.Stat(output); !os.IsNotExist(err) {
		t.Fatalf("invalid output was committed: %v", err)
	}
}

func TestConcurrentRendering(t *testing.T) {
	dir := t.TempDir()
	input := filepath.Join(dir, "in.typ")
	if err := os.WriteFile(input, []byte("{{ .Name }}"), 0o600); err != nil {
		t.Fatal(err)
	}
	bin := fakeTypst(t, `printf '%s' '%PDF-1.7' > "$3"`)
	r := New(Options{BinaryPath: bin})
	for i := 0; i < 8; i++ {
		i := i
		t.Run("parallel", func(t *testing.T) {
			t.Parallel()
			path := filepath.Join(dir, "out", string(rune('a'+i))+".pdf")
			if _, err := r.RenderFile(context.Background(), input, path, map[string]any{"Name": i}); err != nil {
				t.Fatal(err)
			}
		})
	}
}
