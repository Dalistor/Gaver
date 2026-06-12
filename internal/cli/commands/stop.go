package commands

import (
	"fmt"
	"path/filepath"

	"github.com/Dalistor/gaver/core/pidstore"
	"github.com/spf13/cobra"
)

var StopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Para todos os módulos em execução gerenciados pelo Gaver",
	RunE: func(cmd *cobra.Command, args []string) error {
		absDir, err := filepath.Abs(".")
		if err != nil {
			return err
		}

		pids := pidstore.List(absDir)
		if len(pids) == 0 {
			fmt.Println("Nenhum módulo em execução.")
			return nil
		}

		stopAll(absDir)
		return nil
	},
}
