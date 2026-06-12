package commands

import (
	"github.com/spf13/cobra"
)

var BuildCmd = &cobra.Command{
	Use:   "build",
	Short: "Compila o módulo e seus sub-módulos recursivamente",
	RunE: func(cmd *cobra.Command, args []string) error {
		parallel, _ := cmd.Flags().GetBool("parallel")
		return runCascade(".", "build", parallel)
	},
}

func init() {
	BuildCmd.Flags().BoolP("parallel", "p", false, "Compila sub-módulos em paralelo")
}
