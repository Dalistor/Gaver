package commands

import (
	"fmt"

	"github.com/Dalistor/gaver/core/downloader"
	"github.com/Dalistor/gaver/core/registry"
	"github.com/Dalistor/gaver/internal/scaffold"
	"github.com/spf13/cobra"
)

var NewCmd = &cobra.Command{
	Use:   "new",
	Short: "Cria um novo projeto Gaver a partir de um repositório de templates",
	Example: `  gaver new --type api --name meu-projeto --from oficial
  gaver new --type webapp --name meu-site --from minha-org`,
	RunE: func(cmd *cobra.Command, args []string) error {
		projectType, _ := cmd.Flags().GetString("type")
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

		fmt.Printf("Criando projeto %q...\n", name)
		if err := scaffold.Generate(repoDir, projectType, name); err != nil {
			return err
		}

		fmt.Printf("\nProjeto criado em ./%s\n", name)
		fmt.Printf("  cd %s\n", name)
		fmt.Printf("  gaver init\n")
		fmt.Printf("  gaver run\n")
		return nil
	},
}

// resolveRepo retorna o repositório a usar. Se fromRepo estiver vazio e houver
// apenas um repositório registrado, usa-o automaticamente.
func resolveRepo(cfg *registry.Config, fromRepo string) (registry.Repo, error) {
	if fromRepo != "" {
		return cfg.Get(fromRepo)
	}
	if len(cfg.Repositories) == 0 {
		return registry.Repo{}, fmt.Errorf("nenhum repositório registrado. Use: gaver repo add <name> <url>")
	}
	if len(cfg.Repositories) == 1 {
		return cfg.Repositories[0], nil
	}
	return registry.Repo{}, fmt.Errorf("múltiplos repositórios registrados — especifique com --from <nome>")
}

func init() {
	NewCmd.Flags().StringP("type", "t", "", "Tipo do projeto (api, webapp, agent, cli)")
	NewCmd.Flags().StringP("name", "n", "", "Nome do projeto")
	NewCmd.Flags().StringP("from", "f", "", "Nome do repositório de templates a usar")

	NewCmd.MarkFlagRequired("type")
	NewCmd.MarkFlagRequired("name")
}
