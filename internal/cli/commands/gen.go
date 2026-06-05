package commands

import (
	"fmt"

	"github.com/spf13/cobra"
)

var GenCmd = &cobra.Command{
	Use:   "gen",
	Short: "Gera artefatos para um projeto existente",
}

var genPipelineCmd = &cobra.Command{
	Use:   "pipeline",
	Short: "Gera definição de pipeline de CI/CD",
	RunE: func(cmd *cobra.Command, args []string) error {
		target, _ := cmd.Flags().GetString("target")
		fmt.Printf("Gerando pipeline para %q...\n", target)
		return nil
	},
}

func init() {
	genPipelineCmd.Flags().String("target", "", "Alvo do pipeline (github-actions, gitlab-ci)")
	genPipelineCmd.MarkFlagRequired("target")

	GenCmd.AddCommand(genPipelineCmd)
}
