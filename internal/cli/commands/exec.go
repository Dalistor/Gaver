package commands

import (
	"github.com/spf13/cobra"
)

var ExecCmd = &cobra.Command{
	Use:   "exec <comando>",
	Short: "Executa um comando customizado em cascata por toda a hierarquia de módulos",
	Long: `Percorre todos os módulos instalados e executa o comando informado em cada um
que o declarar em 'commands', respeitando a ordem de depends_on.

Útil para comandos de operação como migrate, seed, test ou qualquer chave
customizada definida nos gaver.json dos módulos. Módulos que não declararem
o comando são silenciosamente ignorados.`,
	Example: `  gaver exec migrate
  gaver exec seed --parallel
  gaver exec test`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		parallel, _ := cmd.Flags().GetBool("parallel")
		return runCascade(".", args[0], parallel)
	},
}

func init() {
	ExecCmd.Flags().BoolP("parallel", "p", false, "Executa sub-módulos independentes em paralelo")
}
