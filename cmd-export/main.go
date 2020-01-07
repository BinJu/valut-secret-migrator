package main

import (
	"fmt"
	"os"

	"github.com/BinJu/vault-secret-migrator/client/offline"
	"github.com/BinJu/vault-secret-migrator/export"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("path is not specified")
		os.Exit(1)
	}
	exporter := export.NewExporter(offline.NewVault())
	err := exporter.Export(os.Args[1], os.Stdout)

	if err != nil {
		fmt.Println("Error:", err)
	}
}
