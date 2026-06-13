package exports

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Dalistor/gaver/core/manifest"
)

func dir(rootDir string) string {
	return filepath.Join(rootDir, ".gaver", "exports")
}

// Save persiste os exports de um módulo em rootDir/.gaver/exports/<name>.json.
func Save(rootDir, name string, exp manifest.Exports) error {
	if len(exp) == 0 {
		return nil
	}
	if err := os.MkdirAll(dir(rootDir), 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(exp, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir(rootDir), name+".json"), data, 0600)
}

// Load carrega os exports salvos de um módulo. Retorna nil sem erro se não existir.
func Load(rootDir, name string) (manifest.Exports, error) {
	data, err := os.ReadFile(filepath.Join(dir(rootDir), name+".json"))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var exp manifest.Exports
	if err := json.Unmarshal(data, &exp); err != nil {
		return nil, fmt.Errorf("exports inválidos de %q: %w", name, err)
	}
	return exp, nil
}

// ToEnvSlice converte os exports de um módulo em pares KEY=VALUE para injeção de ambiente.
// Convenção: {MODULO}_{CHAVE}=protocol://address
//
//	{MODULO}_{CHAVE}_SCHEMA=schema (se declarado)
func ToEnvSlice(moduleName string, exp manifest.Exports) []string {
	prefix := normalize(moduleName)
	var envs []string
	for key, entry := range exp {
		k := normalize(key)
		envs = append(envs, fmt.Sprintf("%s_%s=%s://%s", prefix, k, entry.Protocol, entry.Address))
		if entry.Schema != "" {
			envs = append(envs, fmt.Sprintf("%s_%s_SCHEMA=%s", prefix, k, entry.Schema))
		}
	}
	return envs
}

func normalize(s string) string {
	return strings.ToUpper(strings.NewReplacer("-", "_", ".", "_").Replace(s))
}
