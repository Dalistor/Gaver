package commands

import (
	"fmt"

	"github.com/Dalistor/gaver/core/downloader"
	"github.com/Dalistor/gaver/core/registry"
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
	Example: `  gaver gen module --name orders --from oficial
  gaver gen module --name users`,
	RunE: func(cmd *cobra.Command, args []string) error {
		name, _ := cmd.Flags().GetString("name")
		fromRepo, _ := cmd.Flags().GetString("from")

		cfg, err := registry.Load()
		if err != nil {
			return err
		}

		repo, err := resolveRepo(cfg, fromRepo)
		if err != nil {
			return err
		}

		fmt.Printf("Baixando templates de %q...\n", repo.URL)
		repoDir, cleanup, err := downloader.Clone(repo.URL)
		if err != nil {
			return err
		}
		defer cleanup()

		if err := scaffold.GenerateModule(repoDir, name); err != nil {
			return err
		}

		fmt.Printf("Módulo %q criado em src/modules/%s/\n", name, name)
		fmt.Printf("Registre em main.go: e.Register(%s.New())\n", name)
		return nil
	},
}

func init() {
	genModuleCmd.Flags().StringP("name", "n", "", "Nome do módulo")
	genModuleCmd.Flags().StringP("from", "f", "", "Nome do repositório de templates a usar")
	_ = genModuleCmd.MarkFlagRequired("name")

	GenCmd.AddCommand(genModuleCmd)
}
