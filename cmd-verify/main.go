package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
)

const pathPrefix = "/v1"

func main() {
	if len(os.Args) < 3 {
		fmt.Println("<SRC VAULT> <DST VAULT>")
		return
	}

	src := "https://127.0.0.1:8201"
	dst := "https://127.0.0.1:8200"

	vaultClient := &vault{
		InstanceAddr:  src,
		PathPrefix:    "/v1",
		SkipSSLVerify: true,
		//Token:         "s.CAAuHWSvkHkdmmCLS123cT03",
		Token: "7b35793c-a809-7406-b14c-73611506626a",
	}

	/*result, err := vaultClient.List("/concourse/main")
	if err != nil {
		fmt.Println("failed to list vault:", err.Error())
		return
	}
	fmt.Println("LIST", result)

	value, err := vaultClient.Read("/concourse/main/concourse_for_k8s_token")
	if err != nil {
		if root := err.RootError(); root.Error() == "404" {
			fmt.Println("NOT FOUND")
			return
		}

		fmt.Println("failed to read vault: ", err.Error())
		return
	}

	fmt.Println("VALUE: ")
	for key, val := range value {
		fmt.Printf("%s = %s\n", key, val)
	}*/

	err := vaultClient.Write("/concourse/main/test_10", []byte(`{"value":"secret-1"}`))
	if err != nil {
		fmt.Println("fail to write:", err)
		return
	}

	migrator := &onlineMigrator{
		source: &vault{
			InstanceAddr:  src,
			PathPrefix:    "/v1",
			SkipSSLVerify: true,
			Token:         "7b35793c-a809-7406-b14c-73611506626a",
		},
		dest: &vault{
			InstanceAddr:  dst,
			PathPrefix:    "/v1",
			SkipSSLVerify: true,
			Token:         "s.CAAuHWSvkHkdmmCLS123cT03",
		},
		dryRun: true,
		debug:  true,
	}
	migErr := migrator.Migrate("/concourse/main", "/concourse/shared")
	if migErr != nil {
		fmt.Println("fail to migrate credentials. error:", migErr)
	}
	fmt.Printf("%d credentials are handled\n", HandleCount())
}

type Vault interface {
	Read(key string) (map[string]string, *VaultError)
	Write(key string, value []byte) *VaultError
	List(path string) ([]string, *VaultError)
	Delete(key string) *VaultError
}

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

func (v *vault) List(path string) ([]string, *VaultError) {
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

func (v *vault) Read(key string) (map[string]string, *VaultError) {
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

func (v *vault) Write(key string, value []byte) *VaultError {
	_, err := v.vault("POST", key, value)
	if err != nil {
		return NewVaultError("failed to write secret", err)
	}
	return nil
}

func (v *vault) Delete(key string) *VaultError {
	return nil
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

var handleCount int64 = 0

func HandleCount() int64 {
	return handleCount
}

type Migrator interface {
	Migrate(paths ...string) error
}

type onlineMigrator struct {
	source Vault
	dest   Vault
	dryRun bool
	debug  bool
}

func (m *onlineMigrator) Migrate(paths ...string) error {
	return m.migrate_func(paths)
}

func (m *onlineMigrator) migrate_func(paths []string) error {
	dirs := []string{}

	for _, p := range paths {
		secretList, err := m.source.List(p)
		if err != nil {
			return err
		}
		for _, secret := range secretList {
			realPath := p + "/" + secret
			if realPath[len(realPath)-1] == '/' {
				dirs = append(dirs, realPath[0:len(realPath)-1])
			} else {
				handleCount = handleCount + 1
				valueSrc, err := m.source.Read(realPath)
				if err != nil {
					return err
				}
				valueDst, err := m.dest.Read(realPath)
				if err != nil {
					rootErr := err.RootError()
					if rootErr != nil && rootErr.Error() == "404" {
						if m.dryRun {
							fmt.Println("[WARN] dest vault instance missed key: " + realPath)
						} else {
							data, err := vaultValueToBytes(valueSrc)
							if err != nil {
								return err
							}
							fmt.Println("[WARN] write credential to dest. key:", realPath)
							writeErr := m.dest.Write(realPath, data)
							if writeErr != nil {
								return err
							}
						}
					} else {
						return err
					}
				} else {
					if m.debug {
						fmt.Println("[DEBUG] comparing key:", realPath)
					}

					if !vaultValueEqual(valueSrc, valueDst) {
						if m.dryRun {
							fmt.Println("[WARN] dest value doesn't equal to source value. key: " + realPath)
						} else {
							data, err := vaultValueToBytes(valueSrc)
							if err != nil {
								return err
							}

							fmt.Println("[WARN] update the credential in dest. key:", realPath)
							err = m.dest.Write(realPath, data)
							if err != nil {
								return err
							}
						}
					}
				}
			}
		}
	}
	if len(dirs) > 0 {
		return m.migrate_func(dirs)
	}
	return nil
}

func vaultValueToBytes(value map[string]string) ([]byte, error) {
	return json.Marshal(value)
}

func vaultValueEqual(val1 map[string]string, val2 map[string]string) bool {
	if len(val1) != len(val2) {
		return false
	}
	for key, val := range val1 {
		if val2[key] != val {
			return false
		}
	}
	return true
}
