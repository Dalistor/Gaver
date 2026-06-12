package commands

import (
	"fmt"

	"github.com/Dalistor/gaver/core/registry"
	"github.com/spf13/cobra"
)

var RepoCmd = &cobra.Command{
	Use:   "repo",
	Short: "Gerencia repositórios de templates Gaver-compatíveis",
}

var repoAddCmd = &cobra.Command{
	Use:     "add <name> <url>",
	Short:   "Registra um repositório de templates",
	Example: `  gaver repo add oficial https://github.com/Dalistor/gaver-templates`,
	Args:    cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		name, url := args[0], args[1]
		cfg, err := registry.Load()
		if err != nil {
			return err
		}
		if err := cfg.Add(name, url); err != nil {
			return err
		}
		fmt.Printf("Repositório %q adicionado.\n", name)
		return nil
	},
}

var repoListCmd = &cobra.Command{
	Use:   "list",
	Short: "Lista os repositórios registrados",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := registry.Load()
		if err != nil {
			return err
		}
		if len(cfg.Repositories) == 0 {
			fmt.Println("Nenhum repositório registrado.")
			fmt.Println("Use: gaver repo add <name> <url>")
			return nil
		}
		for _, r := range cfg.Repositories {
			fmt.Printf("  %-20s %s\n", r.Name, r.URL)
		}
		return nil
	},
}

var repoRemoveCmd = &cobra.Command{
	Use:     "remove <name>",
	Short:   "Remove um repositório registrado",
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

func init() {
	RepoCmd.AddCommand(repoAddCmd)
	RepoCmd.AddCommand(repoListCmd)
	RepoCmd.AddCommand(repoRemoveCmd)
}
