package record_test

import (
	"testing"

	"github.com/BinJu/vault-secret-migrator/record"
	"github.com/stretchr/testify/suite"
)

type VaultSecretSuite struct {
	suite.Suite
}

func (s *VaultSecretSuite) TestVaultSecretSerialization() {
	vaultSecret := record.VaultSecret{Path: "/path1", Value: "secret1"}
	kvString := vaultSecret.String()
	vaultSecret2, err := record.NewVaultSecretFromString(kvString)
	s.NoError(err)
	s.Equal(vaultSecret, *vaultSecret2)
}

func TestSuite(t *testing.T) {
	suite.Run(t, &VaultSecretSuite{})
}
