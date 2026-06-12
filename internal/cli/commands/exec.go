package commands

import (
	"github.com/spf13/cobra"
)

var ExecCmd = &cobra.Command{
	Use:   "exec <command>",
	Short: "Executa um comando customizado definido em gaver.json recursivamente",
	Example: `  gaver exec migrate
  gaver exec seed --parallel`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		parallel, _ := cmd.Flags().GetBool("parallel")
		return runCascade(".", args[0], parallel)
	},
}

func init() {
	ExecCmd.Flags().BoolP("parallel", "p", false, "Executa sub-módulos em paralelo")
}
