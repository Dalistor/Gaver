package commands

import (
	"fmt"
	"strings"

	"github.com/Dalistor/gaver/core/downloader"
	"github.com/Dalistor/gaver/core/registry"
	"github.com/Dalistor/gaver/internal/scaffold"
	"github.com/spf13/cobra"
)

var NewCmd = &cobra.Command{
	Use:   "new",
	Short: "Cria um novo projeto a partir de um repositório de templates",
	Long: `Clona um repositório de templates registrado e gera a estrutura do projeto.

O repositório de templates deve conter os arquivos do template na sua raiz.
Use 'gaver repo add' para registrar repositórios.`,
	Example: `  gaver new --name meu-servico
  gaver new --name meu-site --from minha-org
  gaver new --name processador --from oficial`,
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

		fmt.Printf("Criando projeto %q...\n", name)
		if err := scaffold.Generate(repoDir, name); err != nil {
			return err
		}

		fmt.Printf("\nProjeto %q criado em ./%s\n", name, name)
		fmt.Printf("  cd %s\n", name)
		fmt.Printf("  gaver install   # instala sub-módulos declarados em gaver.json\n")
		fmt.Printf("  gaver init      # inicializa dependências\n")
		fmt.Printf("  gaver run       # executa o projeto\n")
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
		return registry.Repo{}, fmt.Errorf(
			"nenhum repositório de templates registrado\n\nRegistre um com: gaver repo add <nome> <url>",
		)
	}
	if len(cfg.Repositories) == 1 {
		return cfg.Repositories[0], nil
	}
	names := make([]string, len(cfg.Repositories))
	for i, r := range cfg.Repositories {
		names[i] = r.Name
	}
	return registry.Repo{}, fmt.Errorf(
		"múltiplos repositórios registrados — use --from <nome> para escolher\n\nDisponíveis: %s\nListe com: gaver repo list",
		strings.Join(names, ", "),
	)
}

func init() {
	NewCmd.Flags().StringP("name", "n", "", "Nome do projeto")
	NewCmd.Flags().StringP("from", "f", "", "Nome do repositório de templates a usar (obrigatório se houver mais de um registrado)")

	NewCmd.MarkFlagRequired("name")
}
