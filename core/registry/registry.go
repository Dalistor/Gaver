package registry

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

type Repo struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

type Config struct {
	Repositories []Repo `json:"repositories"`
}

func configPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".gaver", "config.json"), nil
}

func Load() (*Config, error) {
	path, err := configPath()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return &Config{}, nil
	}
	if err != nil {
		return nil, err
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("config inválida: %w", err)
	}
	return &cfg, nil
}

func (c *Config) Save() error {
	path, err := configPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func (c *Config) Add(name, url string) error {
	for _, r := range c.Repositories {
		if r.Name == name {
			return fmt.Errorf("repositório %q já existe", name)
		}
	}
	c.Repositories = append(c.Repositories, Repo{Name: name, URL: url})
	return c.Save()
}

func (c *Config) Remove(name string) error {
	for i, r := range c.Repositories {
		if r.Name == name {
			c.Repositories = append(c.Repositories[:i], c.Repositories[i+1:]...)
			return c.Save()
		}
	}
	return fmt.Errorf("repositório %q não encontrado", name)
}

func (c *Config) Get(name string) (Repo, error) {
	for _, r := range c.Repositories {
		if r.Name == name {
			return r, nil
		}
	}
	return Repo{}, fmt.Errorf("repositório %q não encontrado", name)
}
