package verify

import (
	"io"

	"github.com/BinJu/vault-secret-migrator/client"
)

type Verifier interface {
	Verify(input io.Reader) error
}

type lastestVerifier struct {
	client client.Vault
}

func (v *lastestVerifier) Verify(input io.Reader) error {
	return nil
}

type consistencyVerifier struct {
	client client.Vault
}

func (v *consistencyVerifier) latestCheck(inpput io.Reader) error {
	return nil
}
