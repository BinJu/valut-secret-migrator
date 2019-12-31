package impt

import (
	"fmt"
	"github.com/BinJu/vault-secret-migrator/client"
	"github.com/BinJu/vault-secret-migrator/record"
	"io"
)

const buff_size = 1024*1024*4

type Importer interface {
	Impt(input io.Reader) error
}

type importer struct {
	client client.Vault
}

func NewImporter(client client.Vault) Importer {
	return &importer{client}
}

func (i *importer) Impt(input io.Reader) error {
	var last []byte
	var buff[buff_size]byte
	for {
		lastLen := 0
		if last != nil {
			copy(buff[0:], last)
			lastLen = len(last)
		}
		cnt, err := input.Read(buff[lastLen:])
		if err == io.EOF || cnt == 0 {
			break
		}
		secrets, tail, err := record.NewMultiVaultSecretsFromString(string(buff[0:lastLen+cnt]))
		last = tail
		fmt.Printf("DEBUG: read %d rec, left %d\n", len(secrets), len(last))
		for _, sec := range secrets {
			fmt.Print("writing[" + sec.Path + "]")
			err := i.client.Write(sec.Path, sec.Value)
			if err != nil {
				return err
			}
			fmt.Println("...DONE")
		}
	}
	return nil
}