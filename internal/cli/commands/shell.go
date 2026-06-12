package commands

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"

	"github.com/Dalistor/gaver/core/manifest"
	"github.com/Dalistor/gaver/core/pidstore"
)

// runShellIn executa um comando de shell no diretório dir, em foreground.
func runShellIn(dir, command string) error {
	c := exec.Command("sh", "-c", command)
	c.Dir = dir
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	c.Stdin = os.Stdin
	return c.Run()
}

// startBackgroundIn inicia um processo em background e salva o PID em rootDir.
func startBackgroundIn(rootDir, name, dir, command string) error {
	c := exec.Command("sh", "-c", command)
	c.Dir = dir
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	if err := c.Start(); err != nil {
		return err
	}
	return pidstore.Save(rootDir, name, c.Process.Pid)
}

// runCascade executa cmdKey no módulo em dir e em todos os sub-módulos instalados,
// respeitando depends_on e executando por níveis (paralelo ou sequencial).
func runCascade(dir, cmdKey string, parallel bool) error {
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return err
	}
	m, err := manifest.LoadFrom(absDir)
	if err != nil {
		return err
	}

	if cmd, ok := m.Commands[cmdKey]; ok && cmd != "" {
		fmt.Printf("[%s] %s\n", m.Name, cmdKey)
		if err := runShellIn(absDir, cmd); err != nil {
			return fmt.Errorf("módulo %q: %w", m.Name, err)
		}
	}

	levels := computeLevels(m.Modules)
	for _, level := range levels {
		if err := execLevel(absDir, level, parallel, func(modDir string) error {
			return runCascade(modDir, cmdKey, parallel)
		}); err != nil {
			return err
		}
	}
	return nil
}

// startNetwork inicia recursivamente todos os run commands da hierarquia em background,
// respeitando depends_on: sub-módulos de nível mais baixo sobem primeiro.
// rootDir é o diretório raiz onde os PIDs são armazenados.
func startNetwork(dir, rootDir string, parallel bool) error {
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return err
	}
	m, err := manifest.LoadFrom(absDir)
	if err != nil {
		return err
	}

	// Sobe sub-módulos por nível (depends_on garante a ordem)
	levels := computeLevels(m.Modules)
	for _, level := range levels {
		if err := execLevel(absDir, level, parallel, func(modDir string) error {
			return startNetwork(modDir, rootDir, parallel)
		}); err != nil {
			return err
		}
	}

	// Depois sobe o próprio módulo
	if cmd, ok := m.Commands["run"]; ok && cmd != "" {
		fmt.Printf("[%s] iniciando...\n", m.Name)
		if err := startBackgroundIn(rootDir, m.Name, absDir, cmd); err != nil {
			return fmt.Errorf("módulo %q: %w", m.Name, err)
		}
	}
	return nil
}

// waitForStop bloqueia até receber SIGINT/SIGTERM e então para todos os módulos.
func waitForStop(rootDir string) {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh
	fmt.Println("\nParando módulos...")
	stopAll(rootDir)
}

// stopAll envia SIGTERM para todos os processos com PID salvo em rootDir.
func stopAll(rootDir string) {
	pids := pidstore.List(rootDir)
	if len(pids) == 0 {
		fmt.Println("Nenhum módulo em execução.")
		return
	}
	for name, pid := range pids {
		proc, err := os.FindProcess(pid)
		if err != nil {
			continue
		}
		if err := proc.Signal(syscall.SIGTERM); err == nil {
			fmt.Printf("[%s] parado (pid %d)\n", name, pid)
		}
		pidstore.Remove(rootDir, name)
	}
}

// execLevel executa fn para cada módulo em mods, em paralelo ou sequencialmente,
// respeitando a ordem dentro do nível.
func execLevel(parentDir string, mods []manifest.ModuleRef, parallel bool, fn func(string) error) error {
	if parallel {
		var wg sync.WaitGroup
		errs := make(chan error, len(mods))
		for _, mod := range mods {
			wg.Add(1)
			go func() {
				defer wg.Done()
				modDir := filepath.Join(parentDir, "modules", mod.Name)
				if _, err := os.Stat(filepath.Join(modDir, "gaver.json")); os.IsNotExist(err) {
					fmt.Printf("aviso: sub-módulo %q não instalado — execute gaver install\n", mod.Name)
					return
				}
				if err := fn(modDir); err != nil {
					errs <- err
				}
			}()
		}
		wg.Wait()
		close(errs)
		for err := range errs {
			return err
		}
		return nil
	}

	for _, mod := range mods {
		modDir := filepath.Join(parentDir, "modules", mod.Name)
		if _, err := os.Stat(filepath.Join(modDir, "gaver.json")); os.IsNotExist(err) {
			fmt.Printf("aviso: sub-módulo %q não instalado — execute gaver install\n", mod.Name)
			continue
		}
		if err := fn(modDir); err != nil {
			return err
		}
	}
	return nil
}

// computeLevels agrupa módulos em níveis de execução respeitando depends_on.
// Nível 0 = sem dependências, nível 1 = depende só do nível 0, etc.
// Detecta ciclos e os coloca em um nível único ao final.
func computeLevels(modules []manifest.ModuleRef) [][]manifest.ModuleRef {
	if len(modules) == 0 {
		return nil
	}

	inDeg := make(map[string]int, len(modules))
	for _, m := range modules {
		inDeg[m.Name] = len(m.DependsOn)
	}

	// dependents[x] = lista de módulos que dependem de x
	dependents := make(map[string][]string)
	for _, m := range modules {
		for _, dep := range m.DependsOn {
			dependents[dep] = append(dependents[dep], m.Name)
		}
	}

	processed := make(map[string]bool, len(modules))
	var levels [][]manifest.ModuleRef

	for len(processed) < len(modules) {
		var level []manifest.ModuleRef
		for _, m := range modules {
			if !processed[m.Name] && inDeg[m.Name] == 0 {
				level = append(level, m)
			}
		}
		if len(level) == 0 {
			// Ciclo detectado — coloca todos os restantes juntos
			for _, m := range modules {
				if !processed[m.Name] {
					level = append(level, m)
				}
			}
			levels = append(levels, level)
			break
		}
		for _, m := range level {
			processed[m.Name] = true
			for _, dep := range dependents[m.Name] {
				inDeg[dep]--
			}
		}
		levels = append(levels, level)
	}

	return levels
}
