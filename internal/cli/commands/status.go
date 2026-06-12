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
	Short: "Mostra o status dos módulos gerenciados pelo Gaver",
	RunE: func(cmd *cobra.Command, args []string) error {
		absDir, err := filepath.Abs(".")
		if err != nil {
			return err
		}

		pids := pidstore.List(absDir)
		if len(pids) == 0 {
			fmt.Println("Nenhum módulo gerenciado. Use gaver run para iniciar.")
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
