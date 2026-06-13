package manifest

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// HealthCheck define como verificar se um módulo está pronto para receber tráfego.
type HealthCheck struct {
	URL      string `json:"url"`
	Timeout  string `json:"timeout,omitempty"`  // ex: "30s"
	Interval string `json:"interval,omitempty"` // ex: "2s"
}

// ExportEntry descreve um endpoint ou canal que este módulo expõe a outros módulos.
type ExportEntry struct {
	Protocol string `json:"protocol"`         // grpc, http, unix, amqp, etc.
	Address  string `json:"address"`           // host:port ou caminho do socket
	Schema   string `json:"schema,omitempty"`  // .proto, openapi.yaml, etc.
}

// Exports é o mapa de canais exportados pelo módulo.
// Chave é o nome do export (ex: "functions", "stream", "events").
type Exports map[string]ExportEntry

// ModuleRef referencia um sub-módulo declarado em gaver.json.
type ModuleRef struct {
	Name      string            `json:"name"`
	Source    string            `json:"source"`
	DependsOn []string          `json:"depends_on,omitempty"`
	Health    *HealthCheck      `json:"health,omitempty"`
	Env       map[string]string `json:"env,omitempty"`      // env vars estáticas injetadas neste módulo
	EnvFrom   []string          `json:"env_from,omitempty"` // módulos cujos exports serão injetados como env vars
}

type Manifest struct {
	Name     string            `json:"name"`
	Version  string            `json:"version"`
	Type     string            `json:"type"`
	Platform string            `json:"platform,omitempty"`
	Parent   string            `json:"parent,omitempty"`
	Exports  Exports           `json:"exports,omitempty"`
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

	if err := m.Validate(); err != nil {
		return nil, err
	}

	return &m, nil
}

// Validate verifica consistência do manifesto e compatibilidade de plataforma.
func (m *Manifest) Validate() error {
	var errs []string

	if strings.TrimSpace(m.Name) == "" {
		errs = append(errs, "\"name\" é obrigatório")
	}

	// Plataforma: avisa se declarada e incompatível com o SO atual
	if m.Platform != "" && m.Platform != "any" {
		current := runtime.GOOS
		// Normaliza: "macos"/"osx" → "darwin"
		declared := strings.ToLower(m.Platform)
		if declared == "macos" || declared == "osx" {
			declared = "darwin"
		}
		if declared != current {
			errs = append(errs, fmt.Sprintf(
				"módulo declarado para plataforma %q mas rodando em %q", m.Platform, current,
			))
		}
	}

	// Módulos: nomes únicos e depends_on válidos
	seen := make(map[string]bool, len(m.Modules))
	known := make(map[string]bool, len(m.Modules))
	for _, mod := range m.Modules {
		known[mod.Name] = true
	}
	for _, mod := range m.Modules {
		if mod.Name == "" {
			errs = append(errs, "todos os módulos devem ter um \"name\"")
			continue
		}
		if seen[mod.Name] {
			errs = append(errs, fmt.Sprintf("módulo %q duplicado em modules[]", mod.Name))
		}
		seen[mod.Name] = true

		if mod.Source != "" {
			valid := strings.HasPrefix(mod.Source, "https://") ||
				strings.HasPrefix(mod.Source, "git@") ||
				strings.HasPrefix(mod.Source, "ssh://")
			if !valid {
				errs = append(errs, fmt.Sprintf(
					"módulo %q: source deve ser URL remota (https:// ou git@host:path), não caminho local", mod.Name,
				))
			}
		}

		for _, dep := range mod.DependsOn {
			if !known[dep] {
				errs = append(errs, fmt.Sprintf(
					"módulo %q depende de %q que não está declarado em modules[]", mod.Name, dep,
				))
			}
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("gaver.json inválido:\n  - %s", strings.Join(errs, "\n  - "))
	}
	return nil
}
