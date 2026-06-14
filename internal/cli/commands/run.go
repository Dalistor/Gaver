package commands

import (
	"fmt"
	"path/filepath"

	"github.com/Dalistor/gaver/core/manifest"
	"github.com/spf13/cobra"
)

var RunCmd = &cobra.Command{
	Use:   "run",
	Short: "Executa o módulo ou inicia toda a rede de sub-módulos em background",
	Long: `Comportamento depende da presença de sub-módulos em gaver.json:

  Sem sub-módulos: executa 'commands.run' em foreground (processo principal).

  Com sub-módulos: inicia cada módulo em background respeitando depends_on,
  aguarda o health check de cada nível antes de subir o próximo, e bloqueia
  até Ctrl+C. Ao receber o sinal, encerra todos os processos graciosamente
  (SIGTERM → aguarda 30s → SIGKILL nos que restarem).

Os logs de cada módulo são prefixados com [nome-do-módulo] para facilitar
a identificação em redes com múltiplos serviços.`,
	Example: `  gaver run
  gaver run --parallel`,
	RunE: func(cmd *cobra.Command, args []string) error {
		parallel, _ := cmd.Flags().GetBool("parallel")

		absDir, err := filepath.Abs(".")
		if err != nil {
			return err
		}

		m, err := manifest.LoadFrom(absDir)
		if err != nil {
			return err
		}

		// Módulo simples: roda em foreground
		if len(m.Modules) == 0 {
			runCmd, ok := m.Commands["run"]
			if !ok || runCmd == "" {
				return fmt.Errorf("nenhum comando 'run' definido em gaver.json\n\nAdicione: \"commands\": { \"run\": \"<comando>\" }")
			}
			fmt.Printf("[%s] run\n", m.Name)
			return runShellIn(absDir, runCmd)
		}

		// Modo rede: sobe todos em background respeitando depends_on
		if err := startNetwork(absDir, absDir, parallel, nil); err != nil {
			return err
		}

		fmt.Println("\nRede iniciada. Pressione Ctrl+C para encerrar graciosamente.")
		waitForStop(absDir)
		return nil
	},
}

func init() {
	RunCmd.Flags().BoolP("parallel", "p", false, "Inicia sub-módulos independentes em paralelo")
}
