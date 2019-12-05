package vault_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/BinJu/vault-secret-migrator/export"
	"testing"
	//  	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type MainSuite struct {
	suite.Suite
}

type fakeVault struct {}

func (v *fakeVault) List(path string) (string, error) {
	if path == "/concourse" {
		return "path1\npath2\npath3/", nil
	} else if path == "/concourse/path3"{
		return "path4\npath5/", nil
	} else if path == "/concourse/path3/path5" {
		return "path6", nil
	}
	return "", errors.New("unknown path")
}

func (v *fakeVault) Read(path string) (string, error) {
	if path == "/concourse/path1" {
		return "secret1", nil
	} else if path == "/concourse/path2" {
		return "secret2", nil
	} else if path == "/concourse/path3/path4" {
		return "secret4", nil
	} else if path == "/concourse/path3/path5/path6" {
		return "secret6", nil
	}
	return "", errors.New("the path does not exist")
}

func (v *fakeVault) Write(path string, value string) error {
	return nil
}

func (ms *MainSuite) TestExport() {
	b := bytes.NewBufferString("[")

	exporter := export.NewExporter(&fakeVault{})
	err := exporter.Export("/concourse", b)
	ms.NoError(err)
	secrets := [] export.VaultSecret{}
	err = json.Unmarshal(append(b.Bytes(), ']'), &secrets)
	ms.NoError(err)
	ms.Len(secrets, 4)
	ms.Equal("/concourse/path1", secrets[0].Path)
	ms.Equal("secret1", secrets[0].Value)
	ms.Equal("/concourse/path2", secrets[1].Path)
	ms.Equal("secret2", secrets[1].Value)

	ms.Equal("/concourse/path3/path4", secrets[2].Path)
	ms.Equal("secret4", secrets[2].Value)

	ms.Equal("/concourse/path3/path5/path6", secrets[3].Path)
	ms.Equal("secret6", secrets[3].Value)
}

func TestSuite(t *testing.T) {
	suite.Run(t, &MainSuite{})
}
