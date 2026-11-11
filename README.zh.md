# TypstRender

TypstRender 是 PT Rvintech Digital Indonesia 开发的 Go 库，可使用 Go 数据填充 Typst（`.typ`）模板，并调用官方 Typst CLI 编译为 PDF。它运行单独安装的可执行程序，并非纯 Go Typst 编译器。

## 功能

- 使用 `text/template` 支持变量、嵌套字段、条件和迭代。
- 支持 map、struct、slice、array 和 nil 数据。
- 提供 Typst 文本转义函数 `typstText`。
- 支持 context 取消、超时和自定义 Typst 二进制路径。
- 使用唯一暂存文件，验证 PDF，清理临时文件并原子替换输出。
- 仅依赖 Go 标准库，没有 Go 模块依赖。

## 要求与安装

- Go 1.27 或更高版本。
- 需单独安装官方 Typst CLI，并将 `typst` 放入 `PATH`，或配置 `Options.BinaryPath`。请参阅 [Typst 安装说明](https://github.com/typst/typst#installation)。

```sh
go get github.com/Rvintech-Digital-Indonesia/TypstRender
```

## 快速开始

模板示例 `invoice.typ`：

```typst
#set page(paper: "a4")
= Invoice
客户：{{ typstText .Customer.Name }}
{{ range .Items }}
- {{ typstText .Name }}：{{ .Price }}
{{ end }}
```

```go
path, err := typstrender.RenderFile(context.Background(), "invoice.typ", "output/invoice.pdf", data)
if err != nil {
    log.Fatal(err)
}
log.Println("已生成：", path)
```

## 模板语法与安全

模板使用 Go `text/template`，支持 `{{ .Name }}`、`{{ .Customer.Name }}`、`{{ if .Paid }}...{{ else }}...{{ end }}` 和 `{{ range .Items }}...{{ end }}`。缺少 map 键时会报错。`{{ typstText .Name }}` 仅针对 Typst 文本上下文转义标记字符；它不能保护代码、表达式、路径或 URL。模板以及有意作为原始 Typst 标记输出的值必须来自可信来源。

API、配置、错误、测试和限制请参阅 [DOC.zh.md](DOC.zh.md)。

## 测试

```sh
go test ./...
go test -race ./...
go vet ./...
go test -cover ./...
go test -tags=integration ./...
```

集成测试需要 Typst CLI。项目许可状态请参阅 [LICENSE.md](LICENSE.md)。
