# Dokumentasi API TypstRender

## 1. Ringkasan

TypstRender adalah library Go dari PT Rvintech Digital Indonesia yang mengisi template sumber Typst dengan data Go lalu mengompilasinya menjadi PDF menggunakan Typst CLI resmi. Library ini membungkus compiler eksternal, bukan compiler pure-Go.

## 2. Fitur

- Variabel, field bertingkat, `if`, `range`, map, struct, slice, dan array melalui `text/template`.
- Pembatalan context dan timeout kompilasi yang dapat dikonfigurasi.
- Temporary file per render, penggunaan concurrent, validasi header PDF, dan penggantian output atomik.
- Tidak ada dependency modul Go pihak ketiga.

## 3. Persyaratan sistem

Go 1.27+ dan Typst CLI yang terpasang. Modul menggunakan standard library Go; pembuatan PDF tetap membutuhkan executable Typst terpisah.

## 4. Instalasi

```sh
go get github.com/Rvintech-Digital-Indonesia/TypstRender
```

Pasang CLI dari [panduan resmi Typst](https://typst.app/open-source/#download), lalu periksa dengan `typst --version`.

## 5. Mulai cepat

```go
path, err := typstrender.RenderFile(ctx, "templates/invoice.typ", "output/invoice.pdf", data)
if err != nil {
    log.Fatal(err)
}
fmt.Println(path)
```

`ctx` wajib diisi. Direktori output dibuat otomatis. Path hasil bersifat absolut dan bersih. Saat berhasil, file tujuan yang ada akan diganti; saat gagal, file tujuan sebelumnya tetap ada.

## 6. Sintaks template dinamis

```typst
= Invoice
Customer: {{ typstText .Customer.Name }}
{{ range .Items }}
- {{ typstText .Name }}: {{ .Price }}
{{ end }}
{{ if .Paid }}Paid{{ else }}Payment pending{{ end }}
```

Key map yang dirujuk harus tersedia (`missingkey=error`). `typstText` meng-escape karakter markup hanya untuk teks Typst. Fungsi ini tidak mengamankan nilai untuk kode, ekspresi, path, atau URL. `text/template` tidak melakukan escaping otomatis; template statis dan nilai yang memuat sintaks Typst mentah harus tepercaya.

## 7. Referensi API

```go
func RenderFile(ctx context.Context, inputPath, outputPath string, data any) (string, error)
func New(options Options) *Renderer
func (r *Renderer) RenderFile(ctx context.Context, inputPath, outputPath string, data any) (string, error)
func IsTemplateInput(path string) bool
```

`RenderFile` menerima input `.typ` dan `.typst`, serta mewajibkan output `.pdf`. `data` boleh nil. Template sementara dibuat di samping sumber agar import dan asset relatif tetap ditemukan. Template asli tidak diubah.

## 8. Konfigurasi

```go
type Options struct {
    BinaryPath string        // default: "typst" dari PATH
    Timeout    time.Duration // default: 30 detik jika nol atau negatif
}
```

Contoh: `typstrender.New(typstrender.Options{BinaryPath: "/usr/local/bin/typst", Timeout: 45*time.Second})`.

## 9. Penanganan error

Gunakan `errors.Is` dengan `ErrInvalidInput`, `ErrTemplateNotFound`, `ErrTemplateExecution`, `ErrBinaryNotFound`, `ErrCompileFailed`, dan `ErrOutputInvalid`. Pembatalan dan deadline mempertahankan `context.Canceled` serta `context.DeadlineExceeded`. Error compiler menyertakan stderr yang dibatasi ukurannya. Error filesystem mempertahankan penyebab asli.

## 10. Pengujian

```sh
go test ./...
go test -race ./...
go vet ./...
go test -cover ./...
go test -tags integration ./...
go test -bench=. ./...
```

Unit test menggunakan executable palsu. Integration test memakai Typst CLI asli dan dilewati jika CLI tidak tersedia.

## 11. Batasan dan keamanan

- Typst harus dipasang terpisah dan tersedia saat runtime.
- Helper hanya mendukung konteks teks Typst, bukan semua konteks bahasanya.
- Template tidak tepercaya dapat menjalankan fitur Typst; proseslah sebagai kode tidak tepercaya. Jangan interpolasi markup mentah yang tidak tepercaya.
- Staging source memerlukan izin tulis pada direktori template untuk mempertahankan resolusi resource relatif.
- Penggantian atomik bergantung pada semantik rename filesystem.

## 12. Lisensi

Ketentuan lisensi belum dipilih pemilik proyek. Jangan mendistribusikan proyek di bawah lisensi asumsi sebelum `LICENSE.md` disetujui.
