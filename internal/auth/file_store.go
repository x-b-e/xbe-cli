package auth

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
)

type fileStore struct {
	path string
}

type fileConfig struct {
	Tokens map[string]string `json:"tokens"`
}

func (s *fileStore) Get(baseURL string) (string, error) {
	config, err := s.load()
	if err != nil {
		return "", err
	}
	value, ok := config.Tokens[baseURL]
	if !ok || value == "" {
		return "", ErrNotFound
	}
	return value, nil
}

func (s *fileStore) Set(baseURL, token string) error {
	config, err := s.load()
	if err != nil && !errors.Is(err, ErrNotFound) {
		return err
	}
	if config.Tokens == nil {
		config.Tokens = map[string]string{}
	}
	config.Tokens[baseURL] = strings.TrimSpace(token)
	return s.save(config)
}

func (s *fileStore) Delete(baseURL string) error {
	config, err := s.load()
	if err != nil {
		return err
	}
	if config.Tokens == nil {
		return ErrNotFound
	}
	if _, ok := config.Tokens[baseURL]; !ok {
		return ErrNotFound
	}
	delete(config.Tokens, baseURL)
	return s.save(config)
}

func (s *fileStore) load() (fileConfig, error) {
	content, err := os.ReadFile(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			return fileConfig{}, ErrNotFound
		}
		return fileConfig{}, err
	}
	var config fileConfig
	if err := json.Unmarshal(content, &config); err != nil {
		return fileConfig{}, err
	}
	return config, nil
}

func (s *fileStore) save(config fileConfig) error {
	if err := os.MkdirAll(filepath.Dir(s.path), 0o700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.path, data, 0o600)
}
