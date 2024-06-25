package envconfig

import (
	"context"
	"errors"

	vault "github.com/hashicorp/vault/api"
)

var ErrorNoKeyFound error = errors.New("error no such key in vault found")

const serviceEnvAddr = "http://vault:8200"
const outsideEnvAddr = "http://localhost:8200"

type EnvStorage struct {
	c *vault.Client
}

func connectToVault(addr string) (*vault.Client, error) {
	config := vault.DefaultConfig()
	config.Address = addr
	c, err := vault.NewClient(config)
	if err != nil {
		return nil, err
	}

	c.SetToken("dev-only-token")
	return c, nil
}

func NewEnvStorage() (EnvStorage, error) {
	eSt, err := connectToVault(serviceEnvAddr)
	return EnvStorage{eSt}, err
}

func NewEnvClientStorage() (EnvStorage, error) {
	eSt, err := connectToVault(outsideEnvAddr)
	return EnvStorage{eSt}, err
}

func (es EnvStorage) EnvUpdateVal(where string, key string, val string) error {
	secret, err := es.c.KVv2("secret").Get(context.Background(), where)
	if err != nil {
		if errors.Is(err, vault.ErrSecretNotFound) { // if need to create new secret path
			secret = &vault.KVSecret{Data: make(map[string]interface{})}
		} else { // another err
			return err
		}
	}

	secret.Data[key] = val

	_, err = es.c.KVv2("secret").Put(context.Background(), where, secret.Data)
	return err
}

func (es EnvStorage) EnvGetVal(where string, key string) (string, error) {
	secret, err := es.c.KVv2("secret").Get(context.Background(), where)
	if err != nil {
		return "", err
	}

	rA, ok := secret.Data[key].(string)
	if !ok {
		return "", ErrorNoKeyFound
	}
	return rA, nil
}
