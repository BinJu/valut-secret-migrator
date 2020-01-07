package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/BinJu/vault-secret-migrator/client"
	"github.com/BinJu/vault-secret-migrator/client/online"
)

const pathPrefix = "/v1"

func main() {
	if len(os.Args) < 3 {
		fmt.Println("<SRC VAULT> <DST VAULT>")
		return
	}

	src := "https://127.0.0.1:8201"
	dst := "https://127.0.0.1:8200"

	migrator := &onlineMigrator{
		source: online.NewVault(src, true, "7b35793c-a809-7406-b14c-73611506626a"),
		dest:   online.NewVault(dst, true, "s.CAAuHWSvkHkdmmCLS123cT03"),
		dryRun: true,
		debug:  true,
	}
	migErr := migrator.Migrate("/concourse/main", "/concourse/shared")
	if migErr != nil {
		fmt.Println("fail to migrate credentials. error:", migErr)
	}
	fmt.Printf("%d credentials are handled\n", HandleCount())
}

type Migrator interface {
	Migrate(paths ...string) error
}

type onlineMigrator struct {
	source client.Vault
	dest   client.Vault
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
					rootErr := err.(*online.VaultError).RootError()
					if rootErr != nil && rootErr.Error() == "404" {
						if m.dryRun {
							fmt.Println("[WARN] dest vault instance missed key: " + realPath)
						} else {
							data, err := vaultValueToBytes(valueSrc)
							if err != nil {
								return err
							}
							fmt.Println("[WARN] write credential to dest. key:", realPath)
							writeErr := m.dest.Write(realPath, string(data))
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
							err = m.dest.Write(realPath, string(data))
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

var handleCount int64 = 0

func HandleCount() int64 {
	return handleCount
}
