package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"syscall"

	"github.com/Dalistor/gaver/core/pidstore"
	"github.com/spf13/cobra"
)

var StatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Mostra PID e status de todos os módulos gerenciados",
	Long: `Lista todos os módulos iniciados por 'gaver run' com seus respectivos PIDs
e estado atual (rodando / parado).

Um módulo aparece como "parado" se o processo não existe mais — por exemplo,
se travou ou foi encerrado externamente. Use 'gaver stop' para limpar os
registros de módulos parados.`,
	Example: `  gaver status`,
	RunE: func(cmd *cobra.Command, args []string) error {
		absDir, err := filepath.Abs(".")
		if err != nil {
			return err
		}

		pids := pidstore.List(absDir)
		if len(pids) == 0 {
			fmt.Println("Nenhum módulo em execução.")
			fmt.Println("Use 'gaver run' para iniciar a rede de módulos.")
			return nil
		}

		fmt.Printf("%-24s %-10s %s\n", "MÓDULO", "PID", "STATUS")
		fmt.Println("─────────────────────────────────────────────")
		for name, pid := range pids {
			status := statusLabel(pid)
			fmt.Printf("%-24s %-10d %s\n", name, pid, status)
		}
		return nil
	},
}

func statusLabel(pid int) string {
	proc, err := os.FindProcess(pid)
	if err != nil {
		return "parado"
	}
	if err := proc.Signal(syscall.Signal(0)); err != nil {
		return "parado"
	}
	return "rodando"
}
