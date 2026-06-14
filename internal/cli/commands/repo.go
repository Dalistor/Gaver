package commands

import (
	"fmt"
	"strings"

	"github.com/Dalistor/gaver/core/registry"
	"github.com/spf13/cobra"
)

var RepoCmd = &cobra.Command{
	Use:   "repo",
	Short: "Gerencia repositórios de templates registrados",
	Long: `Gerencia a lista de repositórios de templates compatíveis com Gaver,
armazenada em ~/.gaver/config.json.

Um repositório compatível deve ter projetos em projects/<tipo>/ e templates
de módulos em modules/<nome>/. Use 'gaver new' e 'gaver gen module' para
consumir os templates dos repositórios registrados.`,
	Example: `  gaver repo add oficial https://github.com/Dalistor/gaver-templates
  gaver repo list
  gaver repo remove oficial`,
}

var repoAddCmd = &cobra.Command{
	Use:   "add <nome> <url>",
	Short: "Registra um repositório de templates",
	Long: `Adiciona um repositório Git à lista de repositórios de templates em
~/.gaver/config.json. A URL deve ser um repositório Git remoto acessível.`,
	Example: `  gaver repo add oficial https://github.com/Dalistor/gaver-templates
  gaver repo add minha-org git@github.com:minha-org/gaver-templates.git`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		name, url := args[0], args[1]

		if !isValidRepoURL(url) {
			return fmt.Errorf(
				"URL inválida: %q\n\nUse URLs remotas: https://github.com/..., git@github.com:..., ssh://...",
				url,
			)
		}

		cfg, err := registry.Load()
		if err != nil {
			return err
		}
		if err := cfg.Add(name, url); err != nil {
			return err
		}
		fmt.Printf("Repositório %q adicionado.\n", name)
		fmt.Printf("Use 'gaver new --type <tipo> --name <nome> --from %s' para criar um projeto.\n", name)
		return nil
	},
}

var repoListCmd = &cobra.Command{
	Use:   "list",
	Short: "Lista os repositórios registrados",
	Example: `  gaver repo list`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := registry.Load()
		if err != nil {
			return err
		}
		if len(cfg.Repositories) == 0 {
			fmt.Println("Nenhum repositório registrado.")
			fmt.Println("Registre um com: gaver repo add <nome> <url>")
			return nil
		}
		fmt.Printf("  %-20s %s\n", "NOME", "URL")
		fmt.Println("  " + strings.Repeat("─", 60))
		for _, r := range cfg.Repositories {
			fmt.Printf("  %-20s %s\n", r.Name, r.URL)
		}
		return nil
	},
}

var repoRemoveCmd = &cobra.Command{
	Use:   "remove <nome>",
	Short: "Remove um repositório registrado",
	Example: `  gaver repo remove oficial`,
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := registry.Load()
		if err != nil {
			return err
		}
		if err := cfg.Remove(args[0]); err != nil {
			return err
		}
		fmt.Printf("Repositório %q removido.\n", args[0])
		return nil
	},
}

func isValidRepoURL(url string) bool {
	return strings.HasPrefix(url, "https://") ||
		strings.HasPrefix(url, "git@") ||
		strings.HasPrefix(url, "ssh://")
}

func init() {
	RepoCmd.AddCommand(repoAddCmd)
	RepoCmd.AddCommand(repoListCmd)
	RepoCmd.AddCommand(repoRemoveCmd)
}
