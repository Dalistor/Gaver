package lockfile

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"time"
)

const filename = "gaver.lock"

// Entry descreve um sub-módulo fixado em um commit específico.
type Entry struct {
	Source string `json:"source"`
	Commit string `json:"commit"`
}

// Lockfile registra os commits exatos de cada sub-módulo instalado.
type Lockfile struct {
	LockedAt string            `json:"locked_at"`
	Modules  map[string]Entry  `json:"modules"`
}

func Load(dir string) (*Lockfile, error) {
	data, err := os.ReadFile(filepath.Join(dir, filename))
	if errors.Is(err, os.ErrNotExist) {
		return &Lockfile{Modules: make(map[string]Entry)}, nil
	}
	if err != nil {
		return nil, err
	}
	var lf Lockfile
	if err := json.Unmarshal(data, &lf); err != nil {
		return nil, err
	}
	if lf.Modules == nil {
		lf.Modules = make(map[string]Entry)
	}
	return &lf, nil
}

func Save(dir string, lf *Lockfile) error {
	lf.LockedAt = time.Now().UTC().Format(time.RFC3339)
	data, err := json.MarshalIndent(lf, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, filename), data, 0644)
}

func (lf *Lockfile) Set(name, source, commit string) {
	lf.Modules[name] = Entry{Source: source, Commit: commit}
}

func (lf *Lockfile) Get(name string) (Entry, bool) {
	e, ok := lf.Modules[name]
	return e, ok
}
