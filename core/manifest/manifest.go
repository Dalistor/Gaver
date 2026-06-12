package manifest

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// ModuleRef referencia um sub-módulo declarado em gaver.json.
type ModuleRef struct {
	Name      string   `json:"name"`
	Source    string   `json:"source"`
	DependsOn []string `json:"depends_on,omitempty"`
}

type Manifest struct {
	Name     string            `json:"name"`
	Version  string            `json:"version"`
	Type     string            `json:"type"`
	Platform string            `json:"platform,omitempty"`
	Parent   string            `json:"parent,omitempty"`
	Modules  []ModuleRef       `json:"modules,omitempty"`
	Commands map[string]string `json:"commands"`
}

func Load() (*Manifest, error) {
	return LoadFrom(".")
}

func LoadFrom(dir string) (*Manifest, error) {
	data, err := os.ReadFile(filepath.Join(dir, "gaver.json"))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("gaver.json não encontrado em %s — este diretório é um projeto Gaver?", dir)
		}
		return nil, err
	}

	var m Manifest
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("gaver.json inválido em %s: %w", dir, err)
	}

	return &m, nil
}
