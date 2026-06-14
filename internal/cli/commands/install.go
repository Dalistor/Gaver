package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/Dalistor/gaver/core/downloader"
	"github.com/Dalistor/gaver/core/lockfile"
	"github.com/Dalistor/gaver/core/manifest"
	"github.com/spf13/cobra"
)

// maxInstallDepth limita a recursão de módulos para prevenir ciclos ou hierarquias malformadas.
const maxInstallDepth = 20

var InstallCmd = &cobra.Command{
	Use:   "install",
	Short: "Clona e instala todos os sub-módulos declarados em gaver.json",
	Long: `Lê os módulos declarados em gaver.json, clona cada repositório remoto em
modules/<nome>/ e repete o processo recursivamente para os sub-módulos de cada um.

Usa gaver.lock para garantir instalações reproduzíveis: se o lock existir, instala
exatamente o commit registrado. Ao instalar um módulo novo, registra o commit atual
no lock. Commite o gaver.lock para que outros ambientes instalem as mesmas versões.`,
	Example: `  gaver install`,
	RunE: func(cmd *cobra.Command, args []string) error {
		absDir, err := filepath.Abs(".")
		if err != nil {
			return err
		}
		return installModules(absDir, 0)
	},
}

func installModules(dir string, depth int) error {
	if depth >= maxInstallDepth {
		return fmt.Errorf("profundidade máxima de módulos (%d) atingida em %s — verifique se há ciclos na hierarquia", maxInstallDepth, dir)
	}

	m, err := manifest.LoadFrom(dir)
	if err != nil {
		return err
	}

	lf, err := lockfile.Load(dir)
	if err != nil {
		return fmt.Errorf("falha ao ler gaver.lock em %s: %w", dir, err)
	}

	lockChanged := false

	for _, mod := range m.Modules {
		modDir := filepath.Join(dir, "modules", mod.Name)

		if _, err := os.Stat(modDir); !os.IsNotExist(err) {
			fmt.Printf("[%s] sub-módulo %q já instalado\n", m.Name, mod.Name)
			if err := installModules(modDir, depth+1); err != nil {
				return err
			}
			continue
		}

		if mod.Source == "" {
			fmt.Printf("[%s] aviso: sub-módulo %q sem source definido — ignorando\n", m.Name, mod.Name)
			continue
		}

		if err := os.MkdirAll(filepath.Join(dir, "modules"), 0755); err != nil {
			return err
		}

		// Usa commit do lock se disponível; caso contrário clona HEAD
		lockedCommit := ""
		if entry, ok := lf.Get(mod.Name); ok && entry.Source == mod.Source {
			lockedCommit = entry.Commit
			fmt.Printf("[%s] instalando %q @ %s...\n", m.Name, mod.Name, lockedCommit[:8])
		} else {
			fmt.Printf("[%s] instalando %q de %s...\n", m.Name, mod.Name, mod.Source)
		}

		if err := downloader.CloneTo(mod.Source, modDir, lockedCommit); err != nil {
			return err
		}

		// Registra ou confirma o commit no lock
		commit, err := downloader.HeadCommit(modDir)
		if err != nil {
			fmt.Printf("[%s] aviso: não foi possível obter commit de %q: %v\n", m.Name, mod.Name, err)
		} else {
			lf.Set(mod.Name, mod.Source, commit)
			lockChanged = true
			fmt.Printf("[%s] %q fixado em %s\n", m.Name, mod.Name, commit[:8])
		}

		if err := installModules(modDir, depth+1); err != nil {
			return err
		}
	}

	if lockChanged {
		if err := lockfile.Save(dir, lf); err != nil {
			return fmt.Errorf("falha ao salvar gaver.lock: %w", err)
		}
	}

	return nil
}
