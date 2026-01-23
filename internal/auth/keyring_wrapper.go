package auth

import (
	"github.com/zalando/go-keyring"
)

// Wrapper functions for keyring operations to allow mocking and abstraction

func saveToKeyring(service, account, password string) error {
	return keyring.Set(service, account, password)
}

func loadFromKeyring(service, account string) (string, error) {
	return keyring.Get(service, account)
}

func deleteFromKeyring(service, account string) error {
	return keyring.Delete(service, account)
}
