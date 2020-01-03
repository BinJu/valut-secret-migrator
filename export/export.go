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

var gExportCount int64 = 0

func NewExporter(client client.Vault) Exporter {
	return &exporter{client}
}

func (e *exporter) Export(path string, writer io.Writer) error {
	dirs := []string{path}
	return e.export_func(dirs, writer)
}

func Count() int64 {
	return gExportCount
}

func (e *exporter) export_func(paths []string, writer io.Writer) error {
	dirs := []string{}
	for _, path := range paths {
		secretsList, err := e.client.List(path)
		if err != nil {
			return err
		}
		secrets := strings.Split(secretsList, "\n")
		for _, secret := range secrets[2:] {
			if secret == "" {
				continue
			}

			realPath := path + "/" + secret
			if realPath[len(realPath)-1] == '/' { //a path
				dirs = append(dirs, realPath[0:len(realPath)-1])
			} else {
				value, err := e.client.Read(realPath)
				if err != nil {
					return err
				}
				gExportCount += 1
				kv := record.VaultSecret{Path: realPath, Value: value}
				_, err = fmt.Fprint(writer, kv.String())
				if err != nil {
					return err
				}
			}
		}
	}
	if len(dirs) > 0 {
		return e.export_func(dirs, writer)
	}
	return nil
}
