package commands

import (
	"fmt"

	"github.com/Dalistor/gaver/internal/scaffold"
	"github.com/spf13/cobra"
)

var GenCmd = &cobra.Command{
	Use:   "gen",
	Short: "Gera artefatos para um projeto existente",
}

var genModuleCmd = &cobra.Command{
	Use:   "module",
	Short: "Gera a estrutura de um novo módulo de domínio",
	Example: `  gaver gen module --name orders
  gaver gen module --name users`,
	RunE: func(cmd *cobra.Command, args []string) error {
		name, _ := cmd.Flags().GetString("name")

		if err := scaffold.GenerateModule(name); err != nil {
			return err
		}

		fmt.Printf("Módulo %q criado em src/modules/%s/\n", name, name)
		fmt.Printf("Registre em main.go: e.Register(%s.New())\n", name)
		return nil
	},
}

func init() {
	genModuleCmd.Flags().StringP("name", "n", "", "Nome do módulo")
	_ = genModuleCmd.MarkFlagRequired("name")

	GenCmd.AddCommand(genModuleCmd)
}
