package commands

import (
	"github.com/spf13/cobra"
)

var InitCmd = &cobra.Command{
	Use:   "init",
	Short: "Inicializa dependências do módulo e de todos os sub-módulos recursivamente",
	Long: `Executa o comando 'init' definido em gaver.json em cascata: primeiro no módulo
raiz e depois em cada sub-módulo instalado, respeitando a ordem de depends_on.

Módulos que não declaram 'commands.init' são ignorados silenciosamente.`,
	Example: `  gaver init
  gaver init --parallel`,
	RunE: func(cmd *cobra.Command, args []string) error {
		parallel, _ := cmd.Flags().GetBool("parallel")
		return runCascade(".", "init", parallel)
	},
}

func init() {
	InitCmd.Flags().BoolP("parallel", "p", false, "Inicializa sub-módulos independentes em paralelo")
}
