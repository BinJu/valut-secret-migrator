package vault_test

import (
	"bytes"
	"errors"
	"testing"

	//  	"github.com/stretchr/testify/require"
	"github.com/BinJu/vault-secret-migrator/export"
	"github.com/BinJu/vault-secret-migrator/record"
	"github.com/stretchr/testify/suite"
)

type MainSuite struct {
	suite.Suite
}

type fakeVault struct{}

func (v *fakeVault) List(path string) (string, error) {
	lineSep := record.RecordSeparator()
	if path == "/concourse/main" {
		return "path1" + lineSep + "path2" + lineSep + "path3/", nil
	} else if path == "/concourse/main/path3" {
		return "path4", nil
	}
	return "", errors.New("unknown path")
}

func (v *fakeVault) Read(path string) (string, error) {
	if path == "/concourse/main/path1" {
		return "secret1||$$\n", nil
	} else if path == "/concourse/main/path2" {
		return "secret2", nil
	} else if path == "/concourse/main/path3/path4" {
		return "secret4", nil
	}
	return "", errors.New("the path does not exist")
}

func (v *fakeVault) Write(path string, value string) error {
	return nil
}

func (ms *MainSuite) TestExport() {
	exporter := export.NewExporter(&fakeVault{})
	buff := bytes.NewBuffer(nil)
	err := exporter.Export("/concourse/main", buff)
	ms.NoError(err)
	data := buff.String()
	secrets, err := record.NewMultiVaultSecretsFromString(data)
	ms.NoError(err)
	ms.Len(secrets, 3)
	ms.Equal("/concourse/main/path1", secrets[0].Path)
	ms.Equal("secret1||$$\n", secrets[0].Value)

	ms.Equal("/concourse/main/path2", secrets[1].Path)
	ms.Equal("secret2", secrets[1].Value)

	ms.Equal("/concourse/main/path3/path4", secrets[2].Path)
	ms.Equal("secret4", secrets[2].Value)
}

func TestSuite(t *testing.T) {
	suite.Run(t, &MainSuite{})
}
