package client

//go:generate
type Vault interface {
	List(path string) ([]string, error)
	Read(key string) (map[string]string, error)
	Write(path string, value string) error
	Delete(key string) error
}
