// Package compiler invokes the official Typst command-line compiler.
package compiler

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

// Sentinel errors classify compiler failures.
var (
	ErrBinaryNotFound = errors.New("typst binary not found")
	ErrCompileFailed  = errors.New("typst compilation failed")
	ErrOutputInvalid  = errors.New("invalid output")
)

// Compile invokes Typst without a shell and includes compiler diagnostics when
// the process exits unsuccessfully.
func Compile(ctx context.Context, binaryPath, inputPath, outputPath string) error {
	if binaryPath == "" {
		binaryPath = "typst"
	}
	resolved, err := exec.LookPath(binaryPath)
	if err != nil {
		return fmt.Errorf("%w: %q: %w", ErrBinaryNotFound, binaryPath, err)
	}
	cmd := exec.CommandContext(ctx, resolved, "compile", inputPath, outputPath)
	var stderr limitedBuffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return fmt.Errorf("typst compile deadline exceeded: %w", context.DeadlineExceeded)
		}
		if errors.Is(ctx.Err(), context.Canceled) {
			return fmt.Errorf("typst compile canceled: %w", context.Canceled)
		}
		message := strings.TrimSpace(stderr.String())
		if message != "" {
			return fmt.Errorf("%w: %s: %w", ErrCompileFailed, message, err)
		}
		return fmt.Errorf("%w: %w", ErrCompileFailed, err)
	}
	return nil
}

// ValidatePDF checks that the compiler produced a readable file with a PDF header.
func ValidatePDF(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("%w: open compiler output: %w", ErrOutputInvalid, err)
	}
	defer f.Close()
	var header [5]byte
	n, err := io.ReadFull(f, header[:])
	if err != nil || n != len(header) || !bytes.Equal(header[:], []byte("%PDF-")) {
		if err != nil {
			return fmt.Errorf("%w: PDF header missing: %w", ErrOutputInvalid, err)
		}
		return fmt.Errorf("%w: file does not start with %%PDF-", ErrOutputInvalid)
	}
	return nil
}

const maxDiagnosticBytes = 32 * 1024

type limitedBuffer struct {
	buffer bytes.Buffer
}

func (b *limitedBuffer) Write(p []byte) (int, error) {
	n := len(p)
	remaining := maxDiagnosticBytes - b.buffer.Len()
	if remaining > 0 {
		if len(p) > remaining {
			p = p[:remaining]
		}
		_, _ = b.buffer.Write(p)
	}
	return n, nil
}

func (b *limitedBuffer) String() string { return b.buffer.String() }
