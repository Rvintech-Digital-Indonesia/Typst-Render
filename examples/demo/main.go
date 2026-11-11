package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/Rvintech-Digital-Indonesia/TypstRender"
)

type invoiceData struct {
	Customer customer
	Items    []item
	Paid     bool
}

type customer struct{ Name, Address string }
type item struct {
	Name  string
	Price int
}

func main() {
	ctx := context.Background()
	renderer := typstrender.New(typstrender.Options{Timeout: 45 * time.Second})
	if err := os.MkdirAll("output", 0o755); err != nil {
		log.Fatal(err)
	}

	staticPath, err := renderer.RenderFile(ctx, "testdata/templates/basic.typ", "output/basic.pdf", nil)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Static PDF:", staticPath)

	mapData := map[string]any{
		"Customer": map[string]any{"Name": "Rovino", "Address": "Mojokerto"},
		"Items": []map[string]any{
			{"Name": "Hosting", "Price": 150000},
			{"Name": "Domain", "Price": 200000},
		},
		"Paid": true,
	}
	path, err := typstrender.RenderFile(ctx, "testdata/templates/invoice.typ", "output/invoice-map.pdf", mapData)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Map invoice PDF:", path)

	structData := invoiceData{
		Customer: customer{Name: "Rvintech", Address: "Indonesia"},
		Items:    []item{{Name: "TypstRender", Price: 350000}},
		Paid:     false,
	}
	path, err = renderer.RenderFile(ctx, "testdata/templates/invoice.typ", "output/invoice-struct.pdf", structData)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Struct invoice PDF:", path)
}
