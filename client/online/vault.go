package online

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/BinJu/vault-secret-migrator/client"
)

type VaultError struct {
	err  error
	desp string
}

func (v *VaultError) Error() string {
	return fmt.Sprintf("[Vault Error] %s. The original error: %v", v.desp, v.err)
}

func (v *VaultError) RootError() error {
	return v.err
}

func NewVaultError(desp string, err error) *VaultError {
	return &VaultError{err: err, desp: desp}
}

type vault struct {
	InstanceAddr  string
	PathPrefix    string
	SkipSSLVerify bool
	Token         string
}

func NewVault(instanceAddr string, skipSSLVerify bool, token string) client.Vault {
	return &vault{
		InstanceAddr:  instanceAddr,
		PathPrefix:    "/v1",
		SkipSSLVerify: skipSSLVerify,
		Token:         token,
	}
}

func (v *vault) List(path string) ([]string, error) {
	data, err := v.vault("LIST", path, nil)
	if err != nil {
		return nil, NewVaultError("failed to list the path: "+path, err)
	}

	var listResult vaultListResponse
	if err := json.Unmarshal(data, &listResult); err != nil {
		return nil, NewVaultError("failed to decode json result: "+string(data), err)
	}
	return listResult.Data.Keys, nil
}

func (v *vault) Read(key string) (map[string]string, error) {
	data, err := v.vault("GET", key, nil)
	if err != nil {
		return nil, NewVaultError("failed to read the key: "+key, err)
	}

	var result vaultReadResponse
	err = json.Unmarshal(data, &result)
	if err != nil {
		return nil, NewVaultError("failed to decode response: "+string(data), err)
	}

	return result.Data, nil
}

func (v *vault) Write(key string, value string) error {
	_, err := v.vault("POST", key, []byte(value))
	if err != nil {
		return NewVaultError("failed to write secret: "+key+" = "+value, err)
	}
	return nil
}

func (v *vault) Delete(key string) error {
	return errors.New("unimplemented")
}

func (v *vault) vault(requestCmd string, path string, data []byte) ([]byte, error) {
	transport := http.DefaultTransport.(*http.Transport)
	transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: v.SkipSSLVerify}
	client := &http.Client{Transport: transport}

	var buff io.Reader
	if data != nil {
		buff = bytes.NewBuffer(data)
	}

	req, err := http.NewRequest(requestCmd, v.InstanceAddr+v.PathPrefix+path, buff)
	if err != nil {
		return nil, fmt.Errorf("failed to initial a request: %v", err)
	}
	req.Header.Add("X-Vault-Token", v.Token) //"7b35793c-a809-7406-b14c-73611506626a")
	req.Header.Add("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send the request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return nil, fmt.Errorf("%d", resp.StatusCode)
	}
	output, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read data from response body: %v", err)
	}
	return output, nil
}

type vaultListData struct {
	Keys []string `json:"keys"`
}

type vaultListResponse struct {
	Data vaultListData `json:"data"`
}

type vaultReadResponse struct {
	Data map[string]string `json:"data"`
}
