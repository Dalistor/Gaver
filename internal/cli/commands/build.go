package commands

import (
	"fmt"

	"github.com/Dalistor/gaver/core/manifest"
	"github.com/spf13/cobra"
)

var BuildCmd = &cobra.Command{
	Use:   "build",
	Short: "Compila o projeto",
	RunE: func(cmd *cobra.Command, args []string) error {
		m, err := manifest.Load()
		if err != nil {
			return err
		}

		if m.Commands.Build == "" {
			return fmt.Errorf("nenhum comando de build definido no gaver.json")
		}

		fmt.Printf("Compilando %q...\n", m.Name)
		return runShell(m.Commands.Build)
	},
}
