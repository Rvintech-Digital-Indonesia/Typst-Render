# TypstRender API 文档

## 概述与功能

TypstRender 通过 `os/exec` 调用官方 Typst CLI，将包含 Go `text/template` 语法的 `.typ` 源文件编译为 PDF。它支持静态和动态模板、嵌套字段、条件、迭代、context 取消、超时、自定义二进制路径、PDF 验证和暂存输出。它不是纯 Go Typst 编译器。

## 系统要求与安装

需要 Go 1.27 或更高版本，并单独安装官方 Typst CLI。除非设置 `Options.BinaryPath`，否则应能从 `PATH` 找到 `typst`。该 Go 模块没有外部依赖。

```sh
go get github.com/Rvintech-Digital-Indonesia/TypstRender
```

## 快速开始

```go
data := map[string]any{"Customer": map[string]any{"Name": "Rovino"}}
path, err := typstrender.RenderFile(ctx, "invoice.typ", "output/invoice.pdf", data)
```

使用 `path` 前先检查 `err`。成功时返回清理后的绝对路径。

## 动态模板语法

模板使用 Go `text/template`，不会进行 HTML 自动转义。支持 `{{ .Name }}`、嵌套字段（`{{ .Customer.Name }}`）、`{{ if .Paid }}...{{ else }}...{{ end }}` 和 `{{ range .Items }}...{{ end }}`。数据可以是 `map[string]any`、含导出字段的 struct、slice、array 或 nil。缺少 map 键时会报错（`missingkey=error`）。

在 Typst 文本上下文插入不可信文本时，可使用 `{{ typstText .Name }}` 转义常见标记字符。此函数不能保护代码、表达式、路径或 URL。模板和原始 Typst 标记应来自可信来源。

## API 参考

### `RenderFile`

```go
func RenderFile(ctx context.Context, inputPath, outputPath string, data any) (string, error)
```

渲染 `.typ` 模板并返回 PDF 的绝对路径。必要时创建输出目录。不会修改输入模板；静态文档可以传 nil 数据。只有编译成功且 PDF 验证通过后才替换已有输出。

### `Options` 和 `New`

```go
type Options struct {
    BinaryPath string
    Timeout    time.Duration
}

func New(options Options) *Renderer
func (r *Renderer) RenderFile(ctx context.Context, inputPath, outputPath string, data any) (string, error)
```

`BinaryPath` 为空时从 `PATH` 查找 `typst`。`Timeout` 为零或负数时使用 30 秒默认值。Renderer 可以并发复用。

### `IsTemplateInput`

```go
func IsTemplateInput(path string) bool
```

判断路径是否使用支持的 `.typ` 扩展名（不区分大小写）。

### 错误

可用 `errors.Is` 检查 `ErrInvalidInput`、`ErrTemplateNotFound`、`ErrTemplateExecution`、`ErrBinaryNotFound`、`ErrCompileFailed` 和 `ErrOutputInvalid`。取消和超时也分别包装 `context.Canceled` 与 `context.DeadlineExceeded`。文件系统错误保留原始原因。

## 配置与输出行为

仅接受 `.typ` 输入和 `.pdf` 输出。支持相对或绝对输出路径，返回路径为清理后的绝对路径。Typst 不通过 shell 启动，参数分别传递。默认超时为 30 秒。唯一命名的临时源文件和 PDF 放在对应文件旁边；源文件暂存于输入文件旁，以保留相对导入和资源解析。有效 PDF 暂存文件会重命名到目标位置。失败时清理暂存文件并保留已有目标文件。渲染时避免与不可信用户共享可写源目录。

## 错误处理

```go
path, err := typstrender.RenderFile(ctx, "invoice.typ", "out/invoice.pdf", data)
if err != nil {
    switch {
    case errors.Is(err, typstrender.ErrBinaryNotFound):
        log.Println("请安装 Typst CLI")
    case errors.Is(err, context.DeadlineExceeded):
        log.Println("渲染超时")
    case errors.Is(err, typstrender.ErrCompileFailed):
        log.Println("Typst 无法编译文档：", err)
    default:
        log.Println("渲染失败：", err)
    }
    return
}
log.Println("已生成：", path)
```

## 测试

运行 `go test ./...`、`go test -race ./...`、`go vet ./...` 和 `go test -cover ./...`。集成测试使用构建标签：`go test -tags=integration ./...`，要求安装官方 Typst CLI。在项目根目录运行 `go run ./examples/demo` 可启动示例。

## 限制

必须单独安装 Typst。该库仅验证 PDF 文件签名，不检查完整 PDF 规范。`typstText` 仅适用于文本上下文。除进程超时外，本库不会隔离模板或限制 Typst 文件系统访问。项目所有者尚未选择许可证；请查看 `TODO.md` 中尚未勾选的许可证事项。
