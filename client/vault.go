package client

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
)

//go:generate
type Vault interface {
	List(path string) (string, error)
	Read(path string) (string, error)
	Write(path string, key, value string) error
}

type vault struct {
}

func NewVault() Vault {
	return &vault{}
}

func vaultCmd(cmd string, params ...string) (string, error) {
	stdOut := bytes.NewBuffer(nil)
	stdErr := bytes.NewBuffer(nil)
	path, pathErr := exec.LookPath("vault")
	if pathErr != nil {
		return "", pathErr
	}
	vaultCmd := exec.Cmd{Path: path,
		Env:    os.Environ(),
		Stdin:  os.Stdin,
		Stdout: stdOut,
		Stderr: stdErr,
		Args:   params,
	}
	err := vaultCmd.Run()
	if err != nil {
		return "", err
	}

	if stdErr.Len() > 0 {
		return "", errors.New(stdErr.String())
	}

	return stdOut.String(), nil
}

func (v *vault) List(path string) (string, error) {
	return vaultCmd("list", path)
}

func (v *vault) Read(path string) (string, error) {
	return vaultCmd("read", "-format=json", path)
}

func (v *vault) Write(path string, key, value string) error {
	_, err := vaultCmd("write", path, key+"="+value)
	return err
}
