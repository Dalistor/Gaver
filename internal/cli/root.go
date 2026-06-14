package cli

import (
	"os"

	"github.com/Dalistor/gaver/internal/cli/commands"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "gaver",
	Short: "Motor de execução e gerenciador de módulos para projetos de qualquer tipo",
	Long: `Gaver orquestra redes de módulos independentes — cada módulo é um repositório
Git com seu próprio gaver.json que declara comandos, dependências e exports.

O Gaver instala, executa, monitora e encerra módulos em qualquer linguagem,
respeitando a ordem de depends_on e injetando variáveis de ambiente entre módulos.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(commands.NewCmd)
	rootCmd.AddCommand(commands.InitCmd)
	rootCmd.AddCommand(commands.RunCmd)
	rootCmd.AddCommand(commands.BuildCmd)
	rootCmd.AddCommand(commands.GenCmd)
	rootCmd.AddCommand(commands.RepoCmd)
	rootCmd.AddCommand(commands.InstallCmd)
	rootCmd.AddCommand(commands.ExecCmd)
	rootCmd.AddCommand(commands.StopCmd)
	rootCmd.AddCommand(commands.StatusCmd)
}
