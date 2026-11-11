package compiler

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func makeCompiler(t *testing.T, body string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "typst-fake")
	if err := os.WriteFile(path, []byte("#!/bin/sh\n"+body+"\n"), 0o700); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestCompileSuccessAndValidation(t *testing.T) {
	compiler := makeCompiler(t, `printf '%s' '%PDF-1.7 fake' > "$3"`)
	out := filepath.Join(t.TempDir(), "out.pdf")
	if err := Compile(context.Background(), compiler, "in.typ", out); err != nil {
		t.Fatal(err)
	}
	if err := ValidatePDF(out); err != nil {
		t.Fatal(err)
	}
}

func TestCompileErrors(t *testing.T) {
	compiler := makeCompiler(t, `echo 'bad source' >&2; exit 7`)
	err := Compile(context.Background(), compiler, "in.typ", "out.pdf")
	if !errors.Is(err, ErrCompileFailed) || !strings.Contains(err.Error(), "bad source") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCompileNonzeroWithoutDiagnostics(t *testing.T) {
	compiler := makeCompiler(t, `exit 2`)
	err := Compile(context.Background(), compiler, "in.typ", "out.pdf")
	if !errors.Is(err, ErrCompileFailed) || !strings.Contains(err.Error(), "code 2") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestMissingBinary(t *testing.T) {
	err := Compile(context.Background(), filepath.Join(t.TempDir(), "missing"), "in.typ", "out.pdf")
	if !errors.Is(err, ErrBinaryNotFound) {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidatePDFRejectsInvalidOutput(t *testing.T) {
	path := filepath.Join(t.TempDir(), "out.pdf")
	if err := os.WriteFile(path, []byte("not pdf"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := ValidatePDF(path); !errors.Is(err, ErrOutputInvalid) {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidatePDFMissingFile(t *testing.T) {
	err := ValidatePDF(filepath.Join(t.TempDir(), "missing.pdf"))
	if !errors.Is(err, ErrOutputInvalid) || !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestLimitedBufferCapsDiagnostics(t *testing.T) {
	var b limitedBuffer
	input := strings.Repeat("x", maxDiagnosticBytes+100)
	n, err := b.Write([]byte(input))
	if err != nil || n != len(input) {
		t.Fatalf("Write returned (%d, %v)", n, err)
	}
	if len(b.String()) != maxDiagnosticBytes {
		t.Fatalf("diagnostic length = %d, want %d", len(b.String()), maxDiagnosticBytes)
	}
}
