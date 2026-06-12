package commands

import (
	"fmt"
	"path/filepath"

	"github.com/Dalistor/gaver/core/manifest"
	"github.com/spf13/cobra"
)

var RunCmd = &cobra.Command{
	Use:   "run",
	Short: "Executa o módulo. Com sub-módulos, inicia a rede inteira em background",
	RunE: func(cmd *cobra.Command, args []string) error {
		parallel, _ := cmd.Flags().GetBool("parallel")

		absDir, err := filepath.Abs(".")
		if err != nil {
			return err
		}

		m, err := manifest.LoadFrom(absDir)
		if err != nil {
			return err
		}

		// Módulo simples: roda em foreground
		if len(m.Modules) == 0 {
			runCmd, ok := m.Commands["run"]
			if !ok || runCmd == "" {
				return fmt.Errorf("nenhum comando run definido em gaver.json")
			}
			fmt.Printf("[%s] run\n", m.Name)
			return runShellIn(absDir, runCmd)
		}

		// Modo rede: sobe todos em background respeitando depends_on
		if err := startNetwork(absDir, absDir, parallel); err != nil {
			return err
		}

		fmt.Println("\nRede iniciada. Pressione Ctrl+C para parar.")
		waitForStop(absDir)
		return nil
	},
}

func init() {
	RunCmd.Flags().BoolP("parallel", "p", false, "Inicia sub-módulos independentes em paralelo")
}
