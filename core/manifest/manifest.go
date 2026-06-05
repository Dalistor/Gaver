package manifest

import (
	"encoding/json"
	"fmt"
	"os"
)

type Manifest struct {
	Name     string   `json:"name"`
	Version  string   `json:"version"`
	Type     string   `json:"type"`
	Platform string   `json:"platform"`
	Database Database `json:"database"`
	Commands Commands `json:"commands"`
}

type Database struct {
	Type string `json:"type"`
}

type Commands struct {
	Init  string `json:"init"`
	Run   string `json:"run"`
	Build string `json:"build"`
}

func Load() (*Manifest, error) {
	data, err := os.ReadFile("gaver.json")
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("gaver.json não encontrado — este diretório é um projeto Gaver?")
		}
		return nil, err
	}

	var m Manifest
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("gaver.json inválido: %w", err)
	}

	return &m, nil
}
