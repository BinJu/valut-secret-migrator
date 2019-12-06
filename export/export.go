package export

import (
	"fmt"
	"io"
	"strings"

	"github.com/BinJu/vault-secret-migrator/client"
	"github.com/BinJu/vault-secret-migrator/record"
)

type Exporter interface {
	Export(path string, writer io.Writer) error
}

type exporter struct {
	client client.Vault
}

func NewExporter(client client.Vault) Exporter {
	return &exporter{client}
}
func (e *exporter) Export(path string, writer io.Writer) error {
	secretsList, err := e.client.List(path)
	if err != nil {
		return err
	}

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
			kv := record.VaultSecret{Path: realPath, Value: value}
			fmt.Fprint(writer, kv.String())
		}

	}

	for _, dir := range dirs {
		e.Export(dir, writer)
	}
	return nil
}
