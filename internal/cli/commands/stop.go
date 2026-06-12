package commands

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/Dalistor/gaver/core/pidstore"
	"github.com/spf13/cobra"
)

var StopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Para todos os módulos em execução com encerramento gracioso",
	RunE: func(cmd *cobra.Command, args []string) error {
		timeout, _ := cmd.Flags().GetDuration("timeout")

		absDir, err := filepath.Abs(".")
		if err != nil {
			return err
		}

		if len(pidstore.List(absDir)) == 0 {
			fmt.Println("Nenhum módulo em execução.")
			return nil
		}

		stopAll(absDir, timeout)
		return nil
	},
}

func init() {
	StopCmd.Flags().Duration("timeout", 30*time.Second, "Tempo máximo para encerramento gracioso antes de SIGKILL")
}
