package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/Dalistor/gaver/core/downloader"
	"github.com/Dalistor/gaver/core/manifest"
	"github.com/spf13/cobra"
)

var InstallCmd = &cobra.Command{
	Use:   "install",
	Short: "Baixa e instala todos os sub-módulos declarados em gaver.json recursivamente",
	RunE: func(cmd *cobra.Command, args []string) error {
		absDir, err := filepath.Abs(".")
		if err != nil {
			return err
		}
		return installModules(absDir)
	},
}

func installModules(dir string) error {
	m, err := manifest.LoadFrom(dir)
	if err != nil {
		return err
	}

	for _, mod := range m.Modules {
		modDir := filepath.Join(dir, "modules", mod.Name)

		if _, err := os.Stat(modDir); !os.IsNotExist(err) {
			fmt.Printf("[%s] sub-módulo %q já instalado\n", m.Name, mod.Name)
			if err := installModules(modDir); err != nil {
				return err
			}
			continue
		}

		if mod.Source == "" {
			fmt.Printf("[%s] aviso: sub-módulo %q sem source definido — ignorando\n", m.Name, mod.Name)
			continue
		}

		fmt.Printf("[%s] instalando %q de %s...\n", m.Name, mod.Name, mod.Source)
		if err := os.MkdirAll(filepath.Join(dir, "modules"), 0755); err != nil {
			return err
		}
		if err := downloader.CloneTo(mod.Source, modDir); err != nil {
			return err
		}

		if err := installModules(modDir); err != nil {
			return err
		}
	}

	return nil
}
