# TypstRender API 文档

## 1. 项目概述

TypstRender 是 PT Rvintech Digital Indonesia 开发的 Go 库。它使用 Go 数据填充 Typst 源模板，并调用官方 Typst CLI 编译 PDF。它是外部编译器的封装，不是纯 Go 编译器。

## 2. 功能

- 通过 `text/template` 支持变量、嵌套字段、`if`、`range`、map、struct、slice 和 array。
- 支持 context 取消和可配置的编译超时。
- 每次渲染使用独立临时文件，支持并发、PDF 文件头校验和原子替换输出。
- 不依赖第三方 Go 模块。

## 3. 系统要求

需要 Go 1.27+ 和已安装的 Typst CLI。Go 模块使用标准库，但生成 PDF 仍需要单独安装 Typst 可执行文件。

## 4. 安装

```sh
go get github.com/Rvintech-Digital-Indonesia/TypstRender
```

请按[官方 Typst 安装指南](https://typst.app/open-source/#download)安装 CLI，并运行 `typst --version` 检查。

## 5. 快速开始

```go
path, err := typstrender.RenderFile(ctx, "templates/invoice.typ", "output/invoice.pdf", data)
if err != nil {
    log.Fatal(err)
}
fmt.Println(path)
```

必须提供 `ctx`。输出目录会自动创建。返回路径是清理后的绝对路径。成功时会替换已存在的目标文件；失败时保留原目标文件。

## 6. 动态模板语法

```typst
= Invoice
Customer: {{ typstText .Customer.Name }}
{{ range .Items }}
- {{ typstText .Name }}: {{ .Price }}
{{ end }}
{{ if .Paid }}Paid{{ else }}Payment pending{{ end }}
```

被引用的 map key 必须存在（`missingkey=error`）。`typstText` 只为 Typst 文本标记上下文转义字符，不能保护代码、表达式、路径或 URL。`text/template` 不会自动转义；模板和包含原始 Typst 语法的值必须可信。

## 7. API 参考

```go
func RenderFile(ctx context.Context, inputPath, outputPath string, data any) (string, error)
func New(options Options) *Renderer
func (r *Renderer) RenderFile(ctx context.Context, inputPath, outputPath string, data any) (string, error)
func IsTemplateInput(path string) bool
```

`RenderFile` 接受 `.typ` 和 `.typst` 输入，输出必须为 `.pdf`。`data` 可以为 nil。临时 Typst 文件放在原模板旁边，以便相对导入和资源继续解析。原模板不会被修改。

## 8. 配置

```go
type Options struct {
    BinaryPath string        // 默认：从 PATH 查找 "typst"
    Timeout    time.Duration // 零或负数时默认 30 秒
}
```

示例：`typstrender.New(typstrender.Options{BinaryPath: "/usr/local/bin/typst", Timeout: 45*time.Second})`。

## 9. 错误处理

使用 `errors.Is` 检查 `ErrInvalidInput`、`ErrTemplateNotFound`、`ErrTemplateExecution`、`ErrBinaryNotFound`、`ErrCompileFailed` 和 `ErrOutputInvalid`。取消和超时分别保留 `context.Canceled` 与 `context.DeadlineExceeded`。编译错误包含长度受限的 stderr 诊断信息。文件系统错误保留原始原因。

## 10. 测试

```sh
go test ./...
go test -race ./...
go vet ./...
go test -cover ./...
go test -tags integration ./...
go test -bench=. ./...
```

单元测试使用伪造的可执行文件。集成测试调用真正的 Typst CLI；未安装 CLI 时会跳过。

## 11. 限制与安全

- Typst 必须单独安装，并在运行时可用。
- Helper 仅适用于 Typst 文本上下文，不适用于所有语言上下文。
- 不可信模板可能执行 Typst 功能，应按不可信代码处理。不要插入不可信的原始标记。
- 为保留相对资源解析，模板目录需要可写以创建临时源文件。
- 原子替换取决于文件系统的 rename 语义。

## 12. 许可证

项目所有者尚未确定许可证条款。在 `LICENSE.md` 获批前，不要假定许可类型进行分发。
