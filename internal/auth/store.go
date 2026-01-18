package auth

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/zalando/go-keyring"
)

var ErrNotFound = errors.New("token not found")

const serviceName = "xbe-cli"

// TokenSource describes where a token was loaded from.
type TokenSource string

const (
	TokenSourceFlag     TokenSource = "flag"
	TokenSourceEnv      TokenSource = "env"
	TokenSourceKeychain TokenSource = "keychain"
	TokenSourceFile     TokenSource = "file"
	TokenSourceNone     TokenSource = "none"
)

// ResolveToken returns the token for a base URL, honoring flag/env/store precedence.
func ResolveToken(baseURL, flagToken string) (string, TokenSource, error) {
	if strings.TrimSpace(flagToken) != "" {
		return strings.TrimSpace(flagToken), TokenSourceFlag, nil
	}
	if token, ok := EnvToken(); ok {
		return token, TokenSourceEnv, nil
	}

	store := DefaultStore()
	normalized := NormalizeBaseURL(baseURL)

	if token, source, err := store.Get(normalized); err == nil {
		return token, source, nil
	} else if errors.Is(err, ErrNotFound) {
		return "", TokenSourceNone, ErrNotFound
	} else {
		return "", TokenSourceNone, err
	}
}

// EnvToken returns a token from environment variables, if present.
func EnvToken() (string, bool) {
	if value := strings.TrimSpace(os.Getenv("XBE_TOKEN")); value != "" {
		return value, true
	}
	if value := strings.TrimSpace(os.Getenv("XBE_API_TOKEN")); value != "" {
		return value, true
	}
	return "", false
}

// NormalizeBaseURL ensures consistent token keys.
func NormalizeBaseURL(baseURL string) string {
	return strings.TrimRight(strings.TrimSpace(baseURL), "/")
}

// Store abstracts token storage.
type Store interface {
	Get(baseURL string) (string, TokenSource, error)
	Set(baseURL, token string) error
	Delete(baseURL string) error
}

// DefaultStore returns a combined keychain+file store.
func DefaultStore() Store {
	return combinedStore{
		keyring: keyringStore{service: serviceName},
		file:    newFileStore(),
	}
}

type combinedStore struct {
	keyring keyringStore
	file    *fileStore
}

func (s combinedStore) Get(baseURL string) (string, TokenSource, error) {
	key := tokenKey(baseURL)
	value, err := s.keyring.Get(key)
	if err == nil {
		return value, TokenSourceKeychain, nil
	}
	if errors.Is(err, keyring.ErrNotFound) {
		value, fileErr := s.file.Get(baseURL)
		if fileErr != nil {
			return "", TokenSourceNone, fileErr
		}
		return value, TokenSourceFile, nil
	}
	value, fileErr := s.file.Get(baseURL)
	if fileErr == nil {
		return value, TokenSourceFile, nil
	}
	if errors.Is(fileErr, ErrNotFound) {
		return "", TokenSourceNone, ErrNotFound
	}
	return "", TokenSourceNone, fileErr
}

func (s combinedStore) Set(baseURL, token string) error {
	key := tokenKey(baseURL)
	if err := s.keyring.Set(key, token); err == nil {
		return nil
	}
	return s.file.Set(baseURL, token)
}

func (s combinedStore) Delete(baseURL string) error {
	key := tokenKey(baseURL)
	keyringErr := s.keyring.Delete(key)
	fileErr := s.file.Delete(baseURL)
	if keyringErr == nil || errors.Is(keyringErr, keyring.ErrNotFound) {
		if fileErr == nil || errors.Is(fileErr, ErrNotFound) {
			return nil
		}
		return fileErr
	}
	if fileErr == nil || errors.Is(fileErr, ErrNotFound) {
		return keyringErr
	}
	return fmt.Errorf("keychain error: %v; file error: %w", keyringErr, fileErr)
}

type keyringStore struct {
	service string
}

func (s keyringStore) Get(key string) (string, error) {
	return keyring.Get(s.service, key)
}

func (s keyringStore) Set(key, value string) error {
	return keyring.Set(s.service, key, value)
}

func (s keyringStore) Delete(key string) error {
	return keyring.Delete(s.service, key)
}

func tokenKey(baseURL string) string {
	return "token:" + baseURL
}

func newFileStore() *fileStore {
	configDir := os.Getenv("XDG_CONFIG_HOME")
	if strings.TrimSpace(configDir) == "" {
		userConfigDir, err := os.UserConfigDir()
		if err == nil {
			configDir = userConfigDir
		}
	}
	if strings.TrimSpace(configDir) == "" {
		configDir = filepath.Join(os.Getenv("HOME"), ".config")
	}
	return &fileStore{path: filepath.Join(configDir, "xbe", "config.json")}
}
