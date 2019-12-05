package export

import (
	"encoding/json"
	"github.com/BinJu/vault-secret-migrator/client"
	"io"
	"strings"
)

type VaultSecret struct {
	Path string `json:"path"`
	Value string `json:"value"`
}

type Exporter interface {
	Export(path string, writer io.Writer) error
}

type exporter struct {
	client client.Vault
}

func NewExporter(client client.Vault) Exporter {
	return &exporter{client}
}
func (e *exporter)Export(path string, writer io.Writer) error {
	secretsList, err := e.client.List(path)
	if err != nil {
		return err
	}

	data := [] VaultSecret {}
	dirs := []string{}

	secrets := strings.Split(secretsList, "\n")
	for _, secret := range secrets {
		realPath := path + "/" + secret
		if realPath[len(realPath)-1] == '/' { //a path
			dirs = append(dirs, realPath[0:len(realPath)-1])
		} else {
			value, err := e.client.Read(realPath)
			if err != nil {
				return err
			}
			kv := VaultSecret{Path: realPath, Value: value}
			data = append(data, kv)
		}

	}

	output, err := json.Marshal(data)
	if err != nil {
		return err
	}

	writer.Write(output[1: len(output)-2])
	for _, dir := range dirs {
		e.Export(dir, writer)
	}
	return nil
}