package client

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"os/exec"
)

//go:generate
type Vault interface {
	List(path string) (string, error)
	Read(path string) (string, error)
	Write(path string, value string) error
}

type vault struct {
}

func NewVault() Vault {
	return &vault{}
}

type VaultError struct {
	StdOut string
	StdErr string
	Err error
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

func (v *vault) List(path string) (string, error) {
	return vaultCmd("list", path)
}

func (v *vault) Read(path string) (string, error) {
	output, err := vaultCmd("read", "-format=json", path)
	if err != nil {
		return "", err
	}
	valueData := vaultData{}
	err = json.Unmarshal([]byte(output), &valueData)
	if err != nil {
		return "", err
	}

	jsonData, err := json.Marshal(valueData.Data)
	if err != nil {
		return "", err
	}
	return string(jsonData), nil
}

func (v *vault) Write(path string, value string) error {
	in := bytes.NewBufferString(value)
	_, err := vaultCmdWithInput("write", in, path, "-")
	return err
}
