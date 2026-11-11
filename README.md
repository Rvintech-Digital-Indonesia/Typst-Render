# TypstRender

TypstRender is a Go library by PT Rvintech Digital Indonesia that fills Typst (`.typ`) templates with Go data and compiles them to PDF with the official Typst CLI. It invokes the installed executable; it is not a pure-Go Typst compiler.

## Features

- Go `text/template` variables, nested values, conditionals, and iteration.
- Maps, structs, slices, arrays, and nil data.
- Typst text escaping helper (`typstText`).
- Context cancellation, a configurable timeout, and custom Typst binary path.
- Unique staging files, PDF validation, cleanup, and atomic output replacement.
- Standard library only; no Go module dependencies.

## Requirements and installation

- Go 1.27 or newer.
- The official Typst CLI installed separately and available as `typst` on `PATH` (or set `Options.BinaryPath`). See [Typst installation](https://github.com/typst/typst#installation).

```sh
go get github.com/Rvintech-Digital-Indonesia/TypstRender
```

## Quick start

Create `invoice.typ`:

```typst
#set page(paper: "a4")
= Invoice
Customer: {{ typstText .Customer.Name }}
{{ range .Items }}
- {{ typstText .Name }}: {{ .Price }}
{{ end }}
```

Render it:

```go
package main

import (
    "context"
    "log"

    "github.com/Rvintech-Digital-Indonesia/TypstRender"
)

func main() {
    data := map[string]any{
        "Customer": map[string]any{"Name": "Rovino"},
        "Items": []map[string]any{{"Name": "Hosting", "Price": 150000}},
    }
    path, err := typstrender.RenderFile(context.Background(), "invoice.typ", "output/invoice.pdf", data)
    if err != nil {
        log.Fatal(err)
    }
    log.Println("Created", path)
}
```

## Template syntax and safety

Templates use Go `text/template`: `{{ .Name }}`, nested fields such as `{{ .Customer.Name }}`, `{{ if .Paid }}...{{ else }}...{{ end }}`, and `{{ range .Items }}...{{ end }}`. Missing map keys are errors. `{{ typstText .Name }}` escapes markup characters for Typst text context. This helper does not make values safe in code, expressions, paths, or URLs. Templates and any values intentionally treated as raw Typst markup must come from trusted sources.

See [DOC.md](DOC.md) for the API, configuration, errors, testing, and limitations.

## Testing

```sh
go test ./...
go test -race ./...
go vet ./...
go test -cover ./...
go test -tags=integration ./...
```

Integration tests require the Typst CLI. See [LICENSE.md](LICENSE.md) for the project's license status.
