package commands

import (
	"fmt"

	"github.com/Dalistor/gaver/internal/scaffold"
	"github.com/spf13/cobra"
)

var NewCmd = &cobra.Command{
	Use:   "new",
	Short: "Cria um novo projeto Gaver",
	Example: `  gaver new --type api --name meu-projeto
  gaver new --type webapp --name meu-site --database postgres`,
	RunE: func(cmd *cobra.Command, args []string) error {
		projectType, _ := cmd.Flags().GetString("type")
		name, _ := cmd.Flags().GetString("name")
		database, _ := cmd.Flags().GetString("database")

		fmt.Printf("Criando projeto %q...\n", name)

		if err := scaffold.Generate(projectType, name, database); err != nil {
			return err
		}

		fmt.Printf("\nProjeto criado em ./%s\n", name)
		fmt.Printf("  cd %s\n", name)
		fmt.Printf("  gaver init\n")
		fmt.Printf("  gaver run\n")
		return nil
	},
}

func init() {
	NewCmd.Flags().StringP("type", "t", "", "Tipo do projeto (api, webapp, agent, cli)")
	NewCmd.Flags().StringP("name", "n", "", "Nome do projeto")
	NewCmd.Flags().StringP("database", "d", "", "Banco de dados (postgres, mysql, sqlite)")

	NewCmd.MarkFlagRequired("type")
	NewCmd.MarkFlagRequired("name")
}
