package main

import (
	"fmt"
	"github.com/BinJu/vault-secret-migrator/client"
	"github.com/BinJu/vault-secret-migrator/export"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("path is not specified")
		os.Exit(1)
	}
	exporter := export.NewExporter(client.NewVault())
	err := exporter.Export(os.Args[1], os.Stdout)

	if err != nil {
		fmt.Println("Error:", err)
	}
}
