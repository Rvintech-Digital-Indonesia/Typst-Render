// Package filesystem contains filesystem operations used by the public API.
package filesystem

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ResolveInput validates a Typst source path and returns its absolute path.
func ResolveInput(path string) (string, error) {
	if path == "" {
		return "", fmt.Errorf("path is empty")
	}
	if !strings.EqualFold(filepath.Ext(path), ".typ") {
		return "", fmt.Errorf("unsupported template extension %q", filepath.Ext(path))
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("resolve input path: %w", err)
	}
	info, err := os.Stat(abs)
	if err != nil {
		return "", err
	}
	if !info.Mode().IsRegular() {
		return "", fmt.Errorf("input is not a regular file")
	}
	return filepath.Clean(abs), nil
}

// PrepareOutput validates an output filename and creates its parent directory.
func PrepareOutput(path string) (string, error) {
	if path == "" {
		return "", fmt.Errorf("path is empty")
	}
	if !strings.EqualFold(filepath.Ext(path), ".pdf") {
		return "", fmt.Errorf("output extension must be .pdf")
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("resolve output path: %w", err)
	}
	abs = filepath.Clean(abs)
	if info, err := os.Stat(abs); err == nil && info.IsDir() {
		return "", fmt.Errorf("output path is a directory")
	} else if err != nil && !os.IsNotExist(err) {
		return "", err
	}
	if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
		return "", err
	}
	return abs, nil
}

// CreateTemplateStage creates a unique Typst file beside the original source.
// Keeping it beside the template preserves relative imports and asset paths.
func CreateTemplateStage(source string) (string, func(), error) {
	f, err := os.CreateTemp(filepath.Dir(source), ".typstrender-*.typ")
	if err != nil {
		return "", nil, err
	}
	name := f.Name()
	if err := f.Close(); err != nil {
		_ = os.Remove(name)
		return "", nil, err
	}
	return name, func() { _ = os.Remove(name) }, nil
}

// CreateOutputStage reserves a unique output file beside the destination. Typst
// can overwrite the empty file, and leaving it reserved avoids a filename race.
func CreateOutputStage(destination string) (string, func(), error) {
	f, err := os.CreateTemp(filepath.Dir(destination), ".typstrender-*.pdf")
	if err != nil {
		return "", nil, err
	}
	name := f.Name()
	if err := f.Close(); err != nil {
		_ = os.Remove(name)
		return "", nil, err
	}
	return name, func() { _ = os.Remove(name) }, nil
}
