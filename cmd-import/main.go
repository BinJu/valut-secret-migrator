package main

import (
	"fmt"
	"os"

	"github.com/BinJu/vault-secret-migrator/client/offline"
	"github.com/BinJu/vault-secret-migrator/impt"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("file name is not parsed")
		return
	}
	fileName := os.Args[1]
	inputStream, fileErr := os.Open(fileName)
	if fileErr != nil {
		fmt.Println("Open file failed: " + fileErr.Error())
		return
	}
	defer inputStream.Close()
	importer := impt.NewImporter(offline.NewVault())
	err := importer.Impt(inputStream)
	if err != nil {
		fmt.Println("ERROR: ", err)
	}
}
