package commands

import (
	"fmt"

	"github.com/Dalistor/gaver/core/manifest"
	"github.com/spf13/cobra"
)

var RunCmd = &cobra.Command{
	Use:   "run",
	Short: "Executa o projeto",
	RunE: func(cmd *cobra.Command, args []string) error {
		m, err := manifest.Load()
		if err != nil {
			return err
		}

		if m.Commands.Run == "" {
			return fmt.Errorf("nenhum comando de run definido no gaver.json")
		}

		fmt.Printf("Executando %q...\n", m.Name)
		return runShell(m.Commands.Run)
	},
}
