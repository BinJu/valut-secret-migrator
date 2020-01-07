package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"

	"github.com/BinJu/vault-secret-migrator/client"
	"github.com/BinJu/vault-secret-migrator/client/online"
)

const pathPrefix = "/v1"

type vaultInfo struct {
	Addr            string
	IgnoreSSLVerify bool
	Token           string
}

type rootPaths []string

func (p *rootPaths) String() string {
	return fmt.Sprintf("%v", *p)
}
func (p *rootPaths) Set(val string) error {
	*p = append(*p, val)
	return nil
}

type command struct {
	Source    vaultInfo
	Target    vaultInfo
	DryRun    bool
	Debug     bool
	RootPaths rootPaths
}

func main() {
	var cmd command
	cmdErr := parseCommand(&cmd)
	if cmdErr != nil {
		fmt.Println("[ERROR]", cmdErr)
		return
	}

	migrator := &onlineMigrator{
		source: online.NewVault(cmd.Source.Addr, cmd.Source.IgnoreSSLVerify, cmd.Source.Token), //"7b35793c-a809-7406-b14c-73611506626a"),
		dest:   online.NewVault(cmd.Target.Addr, cmd.Target.IgnoreSSLVerify, cmd.Target.Token), //"s.CAAuHWSvkHkdmmCLS123cT03"),
		dryRun: cmd.DryRun,
		debug:  cmd.Debug,
	}

	migErr := migrator.Migrate(cmd.RootPaths...)
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

func parseCommand(cmd *command) error {
	flag.StringVar(&cmd.Source.Addr, "source-addr", "", "the source vault address")
	flag.BoolVar(&cmd.Source.IgnoreSSLVerify, "source-ssl-verify", true, "ignore the ssl verification to the source vault")
	flag.StringVar(&cmd.Source.Token, "source-token", "", "the token for the source vault")

	flag.StringVar(&cmd.Target.Addr, "target-addr", "", "the target vault address")
	flag.BoolVar(&cmd.Target.IgnoreSSLVerify, "target-ssl-verify", true, "ignore the ssl verification to the target vault")
	flag.StringVar(&cmd.Target.Token, "target-token", "", "the token for the target vault")

	flag.BoolVar(&cmd.DryRun, "dry-run", true, "dry run mode will not migrate data to target")
	flag.BoolVar(&cmd.Debug, "debug", false, "debug flag")

	flag.Var(&cmd.RootPaths, "root-path", "root paths that migration starts with")

	flag.Parse()

	if err := verifyVaultInfo("source", &cmd.Source); err != nil {
		return err
	}

	if err := verifyVaultInfo("target", &cmd.Target); err != nil {
		return err
	}

	if len(cmd.RootPaths) == 0 {
		return errors.New("root-path should not be empty")
	}

	if cmd.Debug && cmd.DryRun {
		fmt.Println("[WARN] dry run mode")
	}
	return nil
}

func verifyVaultInfo(id string, info *vaultInfo) error {
	if info.Addr == "" {
		return errors.New(id + "-addr is empty")
	}
	if info.Token == "" {
		return errors.New(id + "-token is empty")
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
