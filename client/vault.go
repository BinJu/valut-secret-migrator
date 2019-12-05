package client

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

func (v *vault) List(path string) (string, error) {
	return "", nil
}

func (v *vault) Read(path string) (string, error) {
	return "", nil
}

func (v *vault) Write(path string, value string) error {
	return nil
}
