# TypstRender

TypstRender adalah library Go dari PT Rvintech Digital Indonesia untuk mengisi template Typst (`.typ`) dengan data Go dan mengompilasinya menjadi PDF menggunakan Typst CLI resmi. Library menjalankan executable yang dipasang terpisah; ini bukan compiler Typst pure-Go.

## Fitur

- Variabel, properti bertingkat, kondisi, dan iterasi dengan `text/template`.
- Data map, struct, slice, array, dan `nil`.
- Helper escaping teks Typst (`typstText`).
- Pembatalan context, timeout, dan lokasi binary Typst kustom.
- File staging unik, validasi PDF, pembersihan, dan penggantian output atomik.
- Hanya standard library Go; tanpa dependency module.

## Persyaratan dan instalasi

- Go 1.27 atau lebih baru.
- Typst CLI resmi harus dipasang terpisah dan tersedia sebagai `typst` di `PATH` (atau atur `Options.BinaryPath`). Lihat [panduan instalasi Typst](https://github.com/typst/typst#installation).

```sh
go get github.com/Rvintech-Digital-Indonesia/TypstRender
```

## Mulai cepat

Buat `invoice.typ`:

```typst
#set page(paper: "a4")
= Invoice
Pelanggan: {{ typstText .Customer.Name }}
{{ range .Items }}
- {{ typstText .Name }}: {{ .Price }}
{{ end }}
```

Contoh pemanggilan:

```go
path, err := typstrender.RenderFile(context.Background(), "invoice.typ", "output/invoice.pdf", data)
if err != nil {
    log.Fatal(err)
}
log.Println("Dibuat:", path)
```

## Sintaks template dan keamanan

Template memakai Go `text/template`: `{{ .Nama }}`, properti bertingkat seperti `{{ .Customer.Nama }}`, `{{ if .Aktif }}...{{ else }}...{{ end }}`, dan `{{ range .Items }}...{{ end }}`. Key map yang tidak ada menghasilkan error. `{{ typstText .Nama }}` meng-escape karakter markup untuk konteks teks Typst. Helper ini tidak mengamankan nilai untuk kode, ekspresi, path, atau URL. Template dan nilai yang sengaja dipakai sebagai markup Typst mentah harus berasal dari sumber tepercaya.

Lihat [DOC.id.md](DOC.id.md) untuk API, konfigurasi, error, pengujian, dan batasan.

## Pengujian

```sh
go test ./...
go test -race ./...
go vet ./...
go test -cover ./...
go test -tags=integration ./...
```

Integration test memerlukan Typst CLI. Status lisensi proyek ada di [LICENSE.md](LICENSE.md).
