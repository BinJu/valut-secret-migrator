package main

import (
	"encoding/json"
	"fmt"
	"io"
)

func main() {
	fmt.Println("hello!")
}

type VaultSecret struct {
	Path string `json:"path"`
	Value string `json:"value"`
}


func Export(writer io.Writer) error {
	data := [] VaultSecret {
		{Path: "/concourse/main/userpassword", Value: "password"},
	}
	output, err := json.Marshal(data)
	if err != nil {
		return err
	}

	writer.Write(output)
	return nil
}
