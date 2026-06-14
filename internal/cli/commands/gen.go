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
	Short: "Gera código e estruturas a partir de templates",
	Long: `Gera boilerplate para projetos existentes usando templates do repositório
de templates registrado. Deve ser executado na raiz de um projeto Gaver.`,
	Example: `  gaver gen module --name orders
  gaver gen module --name payments --from oficial`,
}

var genModuleCmd = &cobra.Command{
	Use:   "module",
	Short: "Gera a estrutura de um novo módulo de domínio no projeto atual",
	Long: `Clona o repositório de templates e gera os arquivos do módulo em src/modules/<nome>/.

O template de módulo deve estar em modules/<nome>/ dentro do repositório de templates.
Após gerar, registre o módulo na aplicação conforme as instruções do template.`,
	Example: `  gaver gen module --name orders
  gaver gen module --name payments --from oficial`,
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
	genModuleCmd.Flags().StringP("name", "n", "", "Nome do módulo a gerar")
	genModuleCmd.Flags().StringP("from", "f", "", "Nome do repositório de templates a usar (obrigatório se houver mais de um registrado)")
	_ = genModuleCmd.MarkFlagRequired("name")

	GenCmd.AddCommand(genModuleCmd)
}
