package main

import (
	"fmt"
	"github.com/BinJu/vault-secret-migrator/client"
	"github.com/BinJu/vault-secret-migrator/impt"
	"os"
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
	importer := impt.NewImporter(client.NewVault())
	err := importer.Impt(inputStream)
	if err != nil {
		fmt.Println("ERROR: ", err)
	}
}
