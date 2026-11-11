// Package typstrender renders Typst templates to PDF using the official Typst
// command-line executable.
package typstrender

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Rvintech-Digital-Indonesia/TypstRender/internal/compiler"
	"github.com/Rvintech-Digital-Indonesia/TypstRender/internal/filesystem"
	tpl "github.com/Rvintech-Digital-Indonesia/TypstRender/internal/template"
)

// Sentinel errors classify common rendering failures. Use errors.Is to inspect
// wrapped errors returned by RenderFile.
var (
	ErrInvalidInput      = errors.New("invalid input")
	ErrTemplateNotFound  = errors.New("template not found")
	ErrTemplateExecution = errors.New("template execution failed")
	ErrBinaryNotFound    = compiler.ErrBinaryNotFound
	ErrCompileFailed     = compiler.ErrCompileFailed
	ErrOutputInvalid     = compiler.ErrOutputInvalid
)

const defaultTimeout = 30 * time.Second

// Options configures a Renderer. An empty BinaryPath uses the Typst executable
// found on PATH. A zero or negative Timeout selects the 30 second default.
type Options struct {
	BinaryPath string
	Timeout    time.Duration
}

// Renderer safely renders templates concurrently. Each call uses unique
// temporary files and does not mutate the renderer's configuration.
type Renderer struct {
	binaryPath string
	timeout    time.Duration
}

// New constructs a renderer with the supplied options.
func New(options Options) *Renderer {
	timeout := options.Timeout
	if timeout <= 0 {
		timeout = defaultTimeout
	}
	return &Renderer{binaryPath: options.BinaryPath, timeout: timeout}
}

// RenderFile applies data to a Typst template, compiles it to PDF, and returns
// the cleaned absolute output path. The output directory is created as needed;
// an existing output file is atomically replaced after successful compilation.
func (r *Renderer) RenderFile(ctx context.Context, inputPath, outputPath string, data any) (string, error) {
	if ctx == nil {
		return "", fmt.Errorf("%w: context is nil", ErrInvalidInput)
	}
	if inputPath == "" || outputPath == "" {
		return "", fmt.Errorf("%w: input and output paths are required", ErrInvalidInput)
	}
	if r == nil {
		r = New(Options{})
	}

	inputAbs, err := filesystem.ResolveInput(inputPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "", fmt.Errorf("%w: %s: %w", ErrTemplateNotFound, inputPath, err)
		}
		return "", fmt.Errorf("%w: input: %w", ErrInvalidInput, err)
	}
	outputAbs, err := filesystem.PrepareOutput(outputPath)
	if err != nil {
		return "", fmt.Errorf("%w: output: %w", ErrInvalidInput, err)
	}

	templateFile, cleanupTemplate, err := filesystem.CreateTemplateStage(inputAbs)
	if err != nil {
		return "", fmt.Errorf("create template staging file: %w", err)
	}
	defer cleanupTemplate()
	if err := tpl.RenderFile(inputAbs, templateFile, data); err != nil {
		if errors.Is(err, tpl.ErrParse) || errors.Is(err, tpl.ErrExecute) {
			return "", fmt.Errorf("%w: %w", ErrTemplateExecution, err)
		}
		return "", fmt.Errorf("render template: %w", err)
	}

	stageOutput, cleanupOutput, err := filesystem.CreateOutputStage(outputAbs)
	if err != nil {
		return "", fmt.Errorf("prepare output staging file: %w", err)
	}
	defer cleanupOutput()

	compileCtx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()
	if err := compiler.Compile(compileCtx, r.binaryPath, templateFile, stageOutput); err != nil {
		return "", err
	}
	if err := compiler.ValidatePDF(stageOutput); err != nil {
		return "", err
	}
	if err := os.Rename(stageOutput, outputAbs); err != nil {
		return "", fmt.Errorf("commit output %q: %w", outputAbs, err)
	}
	return filepath.Clean(outputAbs), nil
}

// RenderFile renders a Typst template using the default renderer configuration.
func RenderFile(ctx context.Context, inputPath, outputPath string, data any) (string, error) {
	return New(Options{}).RenderFile(ctx, inputPath, outputPath, data)
}

// IsTemplateInput reports whether path has a supported Typst template extension.
func IsTemplateInput(path string) bool {
	return strings.EqualFold(filepath.Ext(path), ".typ")
}
