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
	Long: `Envia SIGTERM a todos os processos gerenciados pelo Gaver e aguarda o encerramento.
Processos que não encerrarem dentro do timeout recebem SIGKILL.

O encerramento gracioso permite que cada processo finalize conexões abertas,
persista estado e execute cleanup antes de ser terminado forçadamente.`,
	Example: `  gaver stop
  gaver stop --timeout 10s
  gaver stop --timeout 1m`,
	RunE: func(cmd *cobra.Command, args []string) error {
		timeout, _ := cmd.Flags().GetDuration("timeout")

		absDir, err := filepath.Abs(".")
		if err != nil {
			return err
		}

		if len(pidstore.List(absDir)) == 0 {
			fmt.Println("Nenhum módulo em execução.")
			fmt.Println("Use 'gaver status' para verificar o estado dos módulos.")
			return nil
		}

		stopAll(absDir, timeout)
		return nil
	},
}

func init() {
	StopCmd.Flags().Duration("timeout", 30*time.Second, "Tempo máximo de espera pelo encerramento gracioso antes de SIGKILL (ex: 10s, 1m)")
}
