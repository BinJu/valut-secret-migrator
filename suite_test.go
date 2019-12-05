package main_test

import (
	"bytes"
	"encoding/json"
	"github.com/BinJu/vault-secret-migrator"
	"testing"
	//  	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type MainSuite struct {
	suite.Suite
}

func (ms *MainSuite) TestExport() {
	b := bytes.NewBufferString("")
	err := main.Export(b)
	ms.NoError(err)
	secrets := [] main.VaultSecret{}
	err = json.Unmarshal(b.Bytes(), &secrets)
	ms.NoError(err)
	ms.Len(secrets, 1)
	ms.Equal("/concourse/main/userpassword", secrets[0].Path)
	ms.Equal("password", secrets[0].Value)
}

func TestSuite(t *testing.T) {
	suite.Run(t, &MainSuite{})
}
