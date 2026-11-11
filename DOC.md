# TypstRender API Documentation

## 1. Overview

TypstRender is a Go library by PT Rvintech Digital Indonesia that fills Typst source templates with Go data and compiles the result to PDF using the official Typst CLI. It is a wrapper around the external compiler, not a pure-Go compiler.

## 2. Features

- Go `text/template` variables, nested fields, `if`, `range`, maps, structs, slices, and arrays.
- Context cancellation and a configurable compilation timeout.
- Per-render temporary files, concurrent use, PDF header validation, and atomic output replacement.
- No third-party Go module dependencies.

## 3. System requirements

Go 1.27+ and an installed Typst CLI. The library module uses the Go standard library; PDF generation requires the separate Typst executable.

## 4. Installation

```sh
go get github.com/Rvintech-Digital-Indonesia/TypstRender
```

Install the CLI using the [official Typst guide](https://typst.app/open-source/#download), then check `typst --version`.

## 5. Quick start

```go
path, err := typstrender.RenderFile(ctx, "templates/invoice.typ", "output/invoice.pdf", data)
if err != nil {
    log.Fatal(err)
}
fmt.Println(path)
```

`ctx` is required. Output directories are created automatically. The returned path is absolute and cleaned. On success, an existing destination is replaced; on failure, an existing destination stays intact.

## 6. Dynamic template syntax

```typst
= Invoice
Customer: {{ typstText .Customer.Name }}
{{ range .Items }}
- {{ typstText .Name }}: {{ .Price }}
{{ end }}
{{ if .Paid }}Paid{{ else }}Payment pending{{ end }}
```

Map keys that are referenced must exist (`missingkey=error`). `typstText` escapes markup characters for Typst text markup only. It does not make a value safe for code, expressions, paths, or URLs. Static templates and values that contain raw Typst syntax must be trusted; `text/template` itself does not escape output.

## 7. API reference

```go
func RenderFile(ctx context.Context, inputPath, outputPath string, data any) (string, error)
func New(options Options) *Renderer
func (r *Renderer) RenderFile(ctx context.Context, inputPath, outputPath string, data any) (string, error)
func IsTemplateInput(path string) bool
```

`RenderFile` accepts `.typ` and `.typst` inputs and requires a `.pdf` output. `data` may be nil. Templates are staged beside the original source so relative imports and assets continue to resolve. The source template is not modified.

## 8. Configuration

```go
type Options struct {
    BinaryPath string        // default: "typst" resolved from PATH
    Timeout    time.Duration // default: 30 seconds when zero or negative
}
```

Example: `typstrender.New(typstrender.Options{BinaryPath: "/usr/local/bin/typst", Timeout: 45*time.Second})`.

## 9. Error handling

Use `errors.Is` with `ErrInvalidInput`, `ErrTemplateNotFound`, `ErrTemplateExecution`, `ErrBinaryNotFound`, `ErrCompileFailed`, and `ErrOutputInvalid`. Context cancellation and deadlines preserve `context.Canceled` and `context.DeadlineExceeded`. Compiler errors include bounded stderr diagnostics. Filesystem errors retain their original cause.

## 10. Testing

```sh
go test ./...
go test -race ./...
go vet ./...
go test -cover ./...
go test -tags integration ./...
go test -bench=. ./...
```

Unit tests use a fake executable. Integration tests use the actual Typst CLI and skip when it is unavailable.

## 11. Limitations and security

- Typst must be installed separately and available at runtime.
- The helper supports Typst text markup, not every Typst language context.
- Rendering untrusted templates can execute Typst features and must be treated as untrusted-code processing. Do not interpolate untrusted raw markup.
- Temporary source staging requires write access to the template directory to preserve relative resource resolution.
- Atomic replacement depends on filesystem rename semantics.

## 12. License

License terms have not been selected by the project owner. Do not redistribute this project under an assumed license until `LICENSE.md` is approved.
