package commands

import (
	"fmt"

	"github.com/Dalistor/gaver/core/manifest"
	"github.com/spf13/cobra"
)

var InitCmd = &cobra.Command{
	Use:   "init",
	Short: "Inicializa as dependências do projeto",
	RunE: func(cmd *cobra.Command, args []string) error {
		m, err := manifest.Load()
		if err != nil {
			return err
		}

		if m.Commands.Init == "" {
			return fmt.Errorf("nenhum comando de init definido no gaver.json")
		}

		fmt.Printf("Inicializando %q...\n", m.Name)
		return runShell(m.Commands.Init)
	},
}
