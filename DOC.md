# TypstRender API documentation

## Overview and features

TypstRender renders `.typ` source files containing Go `text/template` actions to PDF by invoking the official Typst CLI with `os/exec`. It supports static and dynamic templates, nested fields, conditionals, iteration, context cancellation, timeouts, custom binary paths, PDF validation, and staged output. It is not a pure-Go Typst compiler.

## Requirements and installation

Use Go 1.27 or newer and install the official Typst CLI separately. The executable must be on `PATH` as `typst`, unless `Options.BinaryPath` supplies its path. The module has no external dependencies.

```sh
go get github.com/Rvintech-Digital-Indonesia/TypstRender
```

## Quick start

```go
data := map[string]any{"Customer": map[string]any{"Name": "Rovino"}}
path, err := typstrender.RenderFile(ctx, "invoice.typ", "output/invoice.pdf", data)
```

Check `err` before using `path`. On success, the returned path is absolute and cleaned.

## Dynamic template syntax

Templates use Go `text/template`: `{{ .Name }}`, nested fields (`{{ .Customer.Name }}`), `{{ if .Paid }}...{{ else }}...{{ end }}`, and `{{ range .Items }}...{{ end }}`. Data can be `map[string]any`, structs with exported fields, slices, arrays, or nil. Missing map keys are errors (`missingkey=error`).

Use `{{ typstText .Name }}` to escape common Typst markup characters when inserting untrusted text in a Typst text context. This does not make values safe in code, expressions, paths, or URLs. Templates and raw Typst markup values must come from trusted sources.

## API reference

### `RenderFile`

```go
func RenderFile(ctx context.Context, inputPath, outputPath string, data any) (string, error)
```

Renders a `.typ` template and returns the resulting absolute PDF path. It creates the output directory if needed, preserves the source template, and accepts nil data for static documents. Existing output is replaced only after successful compilation and PDF validation.

### `Options` and `New`

```go
type Options struct {
    BinaryPath string
    Timeout    time.Duration
}

func New(options Options) *Renderer
func (r *Renderer) RenderFile(ctx context.Context, inputPath, outputPath string, data any) (string, error)
```

An empty `BinaryPath` resolves `typst` through `PATH`. A zero or negative `Timeout` selects 30 seconds. A renderer can be reused concurrently.

### `IsTemplateInput`

```go
func IsTemplateInput(path string) bool
```

Reports whether a path uses the supported `.typ` extension (case-insensitive).

### Errors

Use `errors.Is` with `ErrInvalidInput`, `ErrTemplateNotFound`, `ErrTemplateExecution`, `ErrBinaryNotFound`, `ErrCompileFailed`, or `ErrOutputInvalid`. Cancellation and timeout also wrap `context.Canceled` and `context.DeadlineExceeded`. Filesystem errors retain their causes.

## Configuration and output behavior

Only `.typ` inputs and `.pdf` outputs are accepted. Relative and absolute output paths are supported; returned paths are absolute and cleaned. Typst runs without a shell and receives separate arguments. The default timeout is 30 seconds. Unique temporary source and PDF files are staged beside their respective files; staging source beside the input preserves relative imports and assets. A valid staged PDF is renamed into place. Failures remove staging files and preserve an existing destination. Avoid sharing writable source directories with untrusted users during rendering.

## Error handling

```go
path, err := typstrender.RenderFile(ctx, "invoice.typ", "out/invoice.pdf", data)
if err != nil {
    switch {
    case errors.Is(err, typstrender.ErrBinaryNotFound):
        log.Println("Install Typst CLI")
    case errors.Is(err, context.DeadlineExceeded):
        log.Println("Rendering timed out")
    case errors.Is(err, typstrender.ErrCompileFailed):
        log.Println("Typst rejected the document:", err)
    default:
        log.Println("Render failed:", err)
    }
    return
}
log.Println("Created", path)
```

## Testing

Run `go test ./...`, `go test -race ./...`, `go vet ./...`, and `go test -cover ./...`. Integration tests use a build tag: `go test -tags=integration ./...`. They require the official Typst CLI. Run the standalone demo with `go run ./examples/demo` from the project root.

## Limitations

Typst must be installed separately. The library validates the PDF signature, not the full PDF specification. `typstText` is only for text context. The library does not sandbox templates or restrict Typst filesystem access; the process timeout limits execution time. The project owner has not selected a license; see the unchecked licensing items in `TODO.md`.
