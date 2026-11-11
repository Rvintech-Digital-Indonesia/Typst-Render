# Dokumentasi API TypstRender

## Gambaran umum dan fitur

TypstRender merender file sumber `.typ` yang berisi sintaks Go `text/template` menjadi PDF dengan menjalankan Typst CLI resmi melalui `os/exec`. Fitur mencakup template statis dan dinamis, field bertingkat, kondisi, iterasi, pembatalan context, timeout, lokasi binary kustom, validasi PDF, dan output staging. Ini bukan compiler Typst pure-Go.

## Persyaratan sistem dan instalasi

Gunakan Go 1.27 atau lebih baru dan pasang Typst CLI resmi secara terpisah. Executable harus tersedia sebagai `typst` di `PATH`, kecuali `Options.BinaryPath` diatur. Module Go ini tidak memiliki dependency eksternal.

```sh
go get github.com/Rvintech-Digital-Indonesia/TypstRender
```

## Mulai cepat

```go
data := map[string]any{"Customer": map[string]any{"Name": "Rovino"}}
path, err := typstrender.RenderFile(ctx, "invoice.typ", "output/invoice.pdf", data)
```

Periksa `err` sebelum menggunakan `path`. Jika berhasil, `path` adalah path absolut yang sudah dibersihkan.

## Sintaks template dinamis

Template menggunakan Go `text/template`: `{{ .Nama }}`, field bertingkat (`{{ .Customer.Nama }}`), `{{ if .Aktif }}...{{ else }}...{{ end }}`, dan `{{ range .Items }}...{{ end }}`. Data dapat berupa `map[string]any`, struct dengan field exported, slice, array, atau nil. Key map yang tidak ada menghasilkan error (`missingkey=error`).

Gunakan `{{ typstText .Nama }}` untuk meng-escape karakter markup Typst umum ketika memasukkan teks yang tidak tepercaya ke konteks teks Typst. Helper ini tidak mengamankan kode, ekspresi, path, atau URL. Template dan nilai markup Typst mentah harus berasal dari sumber tepercaya.

## Referensi API

### `RenderFile`

```go
func RenderFile(ctx context.Context, inputPath, outputPath string, data any) (string, error)
```

Merender template `.typ` dan mengembalikan path PDF absolut. Direktori output dibuat bila perlu. Template sumber tidak diubah; data nil valid untuk dokumen statis. Output lama hanya diganti setelah kompilasi berhasil dan validasi PDF lolos.

### `Options` dan `New`

```go
type Options struct {
    BinaryPath string
    Timeout    time.Duration
}

func New(options Options) *Renderer
func (r *Renderer) RenderFile(ctx context.Context, inputPath, outputPath string, data any) (string, error)
```

`BinaryPath` kosong akan mencari `typst` melalui `PATH`. `Timeout` nol atau negatif memilih default 30 detik. Renderer dapat digunakan ulang secara concurrent.

### `IsTemplateInput`

```go
func IsTemplateInput(path string) bool
```

Mengembalikan apakah path menggunakan ekstensi `.typ` (tanpa membedakan huruf besar/kecil).

### Error

Gunakan `errors.Is` untuk memeriksa `ErrInvalidInput`, `ErrTemplateNotFound`, `ErrTemplateExecution`, `ErrBinaryNotFound`, `ErrCompileFailed`, dan `ErrOutputInvalid`. Error pembatalan dan timeout juga membungkus `context.Canceled` dan `context.DeadlineExceeded`. Error filesystem mempertahankan penyebab aslinya.

## Konfigurasi dan perilaku output

Hanya input `.typ` dan output `.pdf` yang diterima. Path output relatif maupun absolut didukung; hasilnya absolut dan dibersihkan. Typst dijalankan tanpa shell dengan argumen terpisah. Timeout default 30 detik. File sumber sementara dan PDF staging memakai nama unik di samping file terkait; staging sumber dekat input mempertahankan resolusi import dan asset relatif. PDF staging yang valid di-rename menjadi output. Kegagalan membersihkan file staging dan menjaga output lama. Hindari berbagi direktori sumber yang dapat ditulis dengan pengguna tidak tepercaya selama render.

## Penanganan error

```go
path, err := typstrender.RenderFile(ctx, "invoice.typ", "out/invoice.pdf", data)
if err != nil {
    switch {
    case errors.Is(err, typstrender.ErrBinaryNotFound):
        log.Println("Pasang Typst CLI")
    case errors.Is(err, context.DeadlineExceeded):
        log.Println("Rendering melewati batas waktu")
    case errors.Is(err, typstrender.ErrCompileFailed):
        log.Println("Typst menolak dokumen:", err)
    default:
        log.Println("Render gagal:", err)
    }
    return
}
log.Println("Dibuat:", path)
```

## Pengujian

Jalankan `go test ./...`, `go test -race ./...`, `go vet ./...`, dan `go test -cover ./...`. Integration test menggunakan build tag: `go test -tags=integration ./...`; Typst CLI resmi harus tersedia. Jalankan demo dengan `go run ./examples/demo` dari root proyek.

## Batasan

Typst harus dipasang terpisah. Library hanya memvalidasi signature PDF, bukan seluruh spesifikasi PDF. `typstText` hanya untuk konteks teks. Library tidak mengisolasi template atau membatasi akses filesystem Typst; timeout membatasi durasi proses. Pemilik proyek belum memilih lisensi; lihat item lisensi yang belum dicentang di `TODO.md`.
