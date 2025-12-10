package keyring

import (
	"github.com/zalando/go-keyring"
)

const (
	serviceName = "pact"
	tokenKey    = "github_token"
)

// SetToken stores the GitHub token in the OS keychain
func SetToken(token string) error {
	return keyring.Set(serviceName, tokenKey, token)
}

// GetToken retrieves the GitHub token from the OS keychain
func GetToken() (string, error) {
	return keyring.Get(serviceName, tokenKey)
}

// DeleteToken removes the GitHub token from the OS keychain
func DeleteToken() error {
	return keyring.Delete(serviceName, tokenKey)
}

// HasToken checks if a token exists in the keychain
func HasToken() bool {
	_, err := GetToken()
	return err == nil
}

// SetSecret stores a secret in the OS keychain
func SetSecret(name, value string) error {
	return keyring.Set(serviceName, name, value)
}

// GetSecret retrieves a secret from the OS keychain
func GetSecret(name string) (string, error) {
	return keyring.Get(serviceName, name)
}

// DeleteSecret removes a secret from the OS keychain
func DeleteSecret(name string) error {
	return keyring.Delete(serviceName, name)
}

// HasSecret checks if a secret exists in the keychain
func HasSecret(name string) bool {
	_, err := GetSecret(name)
	return err == nil
}
