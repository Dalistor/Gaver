package commands

import (
	"github.com/spf13/cobra"
)

var InitCmd = &cobra.Command{
	Use:   "init",
	Short: "Inicializa as dependências do módulo e seus sub-módulos recursivamente",
	RunE: func(cmd *cobra.Command, args []string) error {
		parallel, _ := cmd.Flags().GetBool("parallel")
		return runCascade(".", "init", parallel)
	},
}

func init() {
	InitCmd.Flags().BoolP("parallel", "p", false, "Executa sub-módulos em paralelo")
}
