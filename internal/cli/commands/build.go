package commands

import (
	"github.com/spf13/cobra"
)

var BuildCmd = &cobra.Command{
	Use:   "build",
	Short: "Compila o módulo e todos os sub-módulos recursivamente",
	Long: `Executa o comando 'build' definido em gaver.json em cascata: primeiro no módulo
raiz e depois em cada sub-módulo instalado, respeitando a ordem de depends_on.

Módulos que não declaram 'commands.build' são ignorados silenciosamente.`,
	Example: `  gaver build
  gaver build --parallel`,
	RunE: func(cmd *cobra.Command, args []string) error {
		parallel, _ := cmd.Flags().GetBool("parallel")
		return runCascade(".", "build", parallel)
	},
}

func init() {
	BuildCmd.Flags().BoolP("parallel", "p", false, "Compila sub-módulos independentes em paralelo")
}
