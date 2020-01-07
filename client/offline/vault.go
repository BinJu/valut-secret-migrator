package offline

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/BinJu/vault-secret-migrator/client"
)

type vault struct {
}

func NewVault() client.Vault {
	return &vault{}
}

type VaultError struct {
	StdOut string
	StdErr string
	Err    error
}

type vaultData struct {
	Data map[string]string `json:"data"`
}

func (e *VaultError) Error() string {
	return "STDOUT: " + e.StdOut + "\nSTDERR: " + e.StdErr + "\nERROR: " + e.Err.Error()
}

func vaultCmd(cmd string, params ...string) (string, error) {
	return vaultCmdWithInput(cmd, os.Stdin, params...)
}

func vaultCmdWithInput(cmd string, input io.Reader, params ...string) (string, error) {
	stdOut := bytes.NewBuffer(nil)
	stdErr := bytes.NewBuffer(nil)
	args := []string{cmd}
	args = append(args, params...)
	vaultCmd := exec.Command("vault", args...)
	vaultCmd.Stdout = stdOut
	vaultCmd.Stderr = stdErr
	if input != nil {
		vaultCmd.Stdin = input
	}
	err := vaultCmd.Run()
	if err != nil {
		return "", &VaultError{StdOut: stdOut.String(), StdErr: stdErr.String(), Err: err}
	}

	if stdErr.Len() > 0 {
		return stdOut.String(), &VaultError{StdOut: stdOut.String(), StdErr: stdErr.String(), Err: err}
	}

	return stdOut.String(), nil
}

func (v *vault) List(path string) ([]string, error) {
	data, err := vaultCmd("list", path)
	if err != nil {
		return nil, err
	}
	items := strings.Split(data, "\n")
	if len(items) > 3 {
		items = items[2 : len(items)-1]
	} else {
		items = []string{}
	}
	return items, nil

}

func (v *vault) Read(path string) (map[string]string, error) {
	output, err := vaultCmd("read", "-format=json", path)
	if err != nil {
		return nil, err
	}
	valueData := vaultData{}
	err = json.Unmarshal([]byte(output), &valueData)
	if err != nil {
		return nil, err
	}

	return valueData.Data, nil
}

func (v *vault) Write(path string, value string) error {
	in := bytes.NewBufferString(value)
	_, err := vaultCmdWithInput("write", in, path, "-")
	return err
}

func (v *vault) Delete(key string) error {
	return errors.New("unimplemented")
}
