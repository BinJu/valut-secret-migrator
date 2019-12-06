package record

import (
	"bytes"
	"errors"
	"runtime"
	"strings"
)

type VaultSecret struct {
	Path  string `json:"path"`
	Value string `json:"value"`
}

func (v *VaultSecret) String() string {
	bytes := bytes.NewBufferString(v.Path)
	bytes.WriteString("||")
	value := strings.ReplaceAll(v.Value, "||", "\\|\\|")
	value = strings.ReplaceAll(value, "$$\n", "\\$\\$\n")
	bytes.WriteString(value)
	bytes.WriteString("$$\n")
	return bytes.String()
}

func NewVaultSecretFromString(secretKV string) (*VaultSecret, error) {
	kv := strings.Split(secretKV, "||")
	if len(kv) != 2 {
		return nil, errors.New("wrong format of secret kv")
	}
	if strings.HasSuffix(kv[1], "$$\n") {
		kv[1] = kv[1][0 : len(kv[1])-3]
	}
	value := strings.ReplaceAll(kv[1], "\\|\\|", "||")
	value = strings.ReplaceAll(value, "\\$\\$\n", "$$\n")
	vaultSecret := VaultSecret{Path: kv[0], Value: value}
	return &vaultSecret, nil
}

func NewMultiVaultSecretsFromString(text string) ([]*VaultSecret, error) {
	secrets := []*VaultSecret{}
	for {
		idx := strings.Index(text, "$$\n")
		if idx > 0 {
			secret, err := NewVaultSecretFromString(text[0:idx])
			if err != nil {
				return secrets, err
			}
			secrets = append(secrets, secret)
			if idx+3 >= len(text) {
				break
			} else {
				text = text[idx+3:]
			}
		} else {
			break
		}
	}
	return secrets, nil
}

func RecordSeparator() string {
	if runtime.GOOS == "windows" {
		return "\r\n"
	} else {
		return "\n"
	}
}
