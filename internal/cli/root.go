package cli

import (
	"os"

	"github.com/Dalistor/gaver/internal/cli/commands"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "gaver",
	Short: "Meta-framework para geração de estruturas de trabalho modulares",
	Long: `Gaver unifica scaffolding de projetos, orquestração de pipelines
e integração com agentes IA em uma única interface consistente.`,
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
}
