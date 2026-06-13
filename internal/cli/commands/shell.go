package commands

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/Dalistor/gaver/core/exports"
	"github.com/Dalistor/gaver/core/manifest"
	"github.com/Dalistor/gaver/core/pidstore"
)

// ── Output ────────────────────────────────────────────────────────────────────

var outputMu sync.Mutex

// prefixWriter prefixa cada linha com "[nome] " e serializa escritas no output.
type prefixWriter struct {
	prefix string
	w      io.Writer
	buf    []byte
}

func (pw *prefixWriter) Write(p []byte) (int, error) {
	var out []byte
	remaining := p
	for len(remaining) > 0 {
		idx := bytes.IndexByte(remaining, '\n')
		if idx == -1 {
			pw.buf = append(pw.buf, remaining...)
			break
		}
		line := append(pw.buf, remaining[:idx+1]...)
		out = append(out, []byte(pw.prefix)...)
		out = append(out, line...)
		pw.buf = pw.buf[:0]
		remaining = remaining[idx+1:]
	}
	if len(out) > 0 {
		outputMu.Lock()
		pw.w.Write(out)
		outputMu.Unlock()
	}
	return len(p), nil
}

// ── Shell runner ──────────────────────────────────────────────────────────────

// platformShell retorna o interpretador de shell adequado ao SO atual.
func platformShell() (string, string) {
	if runtime.GOOS == "windows" {
		return "cmd", "/C"
	}
	return "sh", "-c"
}

// runShellIn executa um comando no dir em foreground (sem prefixo).
func runShellIn(dir, command string) error {
	shell, flag := platformShell()
	c := exec.Command(shell, flag, command)
	c.Dir = dir
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	c.Stdin = os.Stdin
	return c.Run()
}

// startBackgroundIn inicia um processo em background, prefixa logs com o nome do módulo
// e salva o PID em rootDir/.gaver/pids/. extraEnv é injetado sobre os env vars do processo atual.
func startBackgroundIn(rootDir, name, dir, command string, extraEnv []string) error {
	shell, flag := platformShell()
	c := exec.Command(shell, flag, command)
	c.Dir = dir
	c.Stdout = &prefixWriter{prefix: fmt.Sprintf("[%s] ", name), w: os.Stdout}
	c.Stderr = &prefixWriter{prefix: fmt.Sprintf("[%s] ", name), w: os.Stderr}
	if len(extraEnv) > 0 {
		c.Env = append(os.Environ(), extraEnv...)
	}
	if err := c.Start(); err != nil {
		return err
	}
	return pidstore.Save(rootDir, name, c.Process.Pid)
}

// buildEnv constrói o slice de env vars a injetar em um módulo:
// - Env estáticas declaradas no ModuleRef
// - Exports dos módulos listados em EnvFrom (carregados de .gaver/exports/)
func buildEnv(rootDir string, mod manifest.ModuleRef) ([]string, error) {
	var env []string
	for k, v := range mod.Env {
		env = append(env, fmt.Sprintf("%s=%s", k, v))
	}
	for _, depName := range mod.EnvFrom {
		exp, err := exports.Load(rootDir, depName)
		if err != nil {
			return nil, fmt.Errorf("módulo %q: erro ao carregar exports de %q: %w", mod.Name, depName, err)
		}
		if len(exp) == 0 {
			fmt.Printf("aviso: módulo %q não possui exports registrados para injetar em %q\n", depName, mod.Name)
			continue
		}
		env = append(env, exports.ToEnvSlice(depName, exp)...)
	}
	return env, nil
}

// ── Cascade (init / build / exec) ────────────────────────────────────────────

// runCascade executa cmdKey no módulo em dir e em todos os sub-módulos instalados,
// respeitando depends_on e executando por níveis.
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

	levels, err := computeLevels(m.Modules)
	if err != nil {
		return fmt.Errorf("módulo %q: %w", m.Name, err)
	}
	for _, level := range levels {
		if err := execLevel(absDir, level, parallel, func(modDir string, _ manifest.ModuleRef) error {
			return runCascade(modDir, cmdKey, parallel)
		}); err != nil {
			return err
		}
	}
	return nil
}

// ── Network run ───────────────────────────────────────────────────────────────

// startNetwork inicia recursivamente todos os run commands em background,
// respeitando depends_on e aguardando health checks antes de subir o próximo nível.
// extraEnv são as vars de ambiente injetadas pelo módulo pai neste módulo.
func startNetwork(dir, rootDir string, parallel bool, extraEnv []string) error {
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return err
	}
	m, err := manifest.LoadFrom(absDir)
	if err != nil {
		return err
	}

	levels, err := computeLevels(m.Modules)
	if err != nil {
		return fmt.Errorf("módulo %q: %w", m.Name, err)
	}
	for _, level := range levels {
		if err := execLevel(absDir, level, parallel, func(modDir string, mod manifest.ModuleRef) error {
			modEnv, err := buildEnv(rootDir, mod)
			if err != nil {
				return err
			}
			if err := startNetwork(modDir, rootDir, parallel, modEnv); err != nil {
				return err
			}
			if err := waitForHealth(mod); err != nil {
				return err
			}
			// Após health check: persiste exports para que dependentes possam carregar via env_from
			subManifest, err := manifest.LoadFrom(modDir)
			if err != nil {
				return err
			}
			if len(subManifest.Exports) > 0 {
				if err := exports.Save(rootDir, mod.Name, subManifest.Exports); err != nil {
					fmt.Printf("[%s] aviso: não foi possível salvar exports: %v\n", mod.Name, err)
				} else {
					fmt.Printf("[%s] exports registrados\n", mod.Name)
				}
			}
			return nil
		}); err != nil {
			return err
		}
	}

	if cmd, ok := m.Commands["run"]; ok && cmd != "" {
		fmt.Printf("[%s] iniciando...\n", m.Name)
		if err := startBackgroundIn(rootDir, m.Name, absDir, cmd, extraEnv); err != nil {
			return fmt.Errorf("módulo %q: %w", m.Name, err)
		}
	}
	return nil
}

// waitForStop bloqueia até SIGINT/SIGTERM e então para todos os módulos com grace period.
func waitForStop(rootDir string) {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh
	fmt.Println("\nParando módulos...")
	stopAll(rootDir, 30*time.Second)
}

// stopAll envia SIGTERM a todos os processos gerenciados, aguarda gracePeriod
// e envia SIGKILL aos que ainda não encerraram.
func stopAll(rootDir string, gracePeriod time.Duration) {
	pids := pidstore.List(rootDir)
	if len(pids) == 0 {
		fmt.Println("Nenhum módulo em execução.")
		return
	}

	procs := make(map[string]*os.Process, len(pids))
	for name, pid := range pids {
		proc, err := os.FindProcess(pid)
		if err != nil {
			pidstore.Remove(rootDir, name)
			continue
		}
		if err := proc.Signal(syscall.SIGTERM); err == nil {
			fmt.Printf("[%s] encerrando (pid %d)...\n", name, pid)
			procs[name] = proc
		} else {
			// Processo já morto — remove PID imediatamente
			pidstore.Remove(rootDir, name)
		}
	}

	deadline := time.After(gracePeriod)
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for len(procs) > 0 {
		select {
		case <-deadline:
			for name, proc := range procs {
				fmt.Printf("[%s] forçando encerramento (SIGKILL)...\n", name)
				proc.Signal(syscall.SIGKILL)
				pidstore.Remove(rootDir, name)
			}
			return
		case <-ticker.C:
			for name, proc := range procs {
				if proc.Signal(syscall.Signal(0)) != nil {
					fmt.Printf("[%s] encerrado\n", name)
					pidstore.Remove(rootDir, name)
					delete(procs, name)
				}
			}
		}
	}
}

// ── Health check ──────────────────────────────────────────────────────────────

// waitForHealth aguarda o módulo declarar-se saudável via HTTP antes de continuar.
// Retorna nil imediatamente se o módulo não declarar health check.
func waitForHealth(mod manifest.ModuleRef) error {
	if mod.Health == nil || mod.Health.URL == "" {
		return nil
	}

	timeout := parseDurationOrDefault(mod.Health.Timeout, 30*time.Second)
	interval := parseDurationOrDefault(mod.Health.Interval, 2*time.Second)

	client := &http.Client{Timeout: interval}
	deadline := time.Now().Add(timeout)

	fmt.Printf("[%s] aguardando health check em %s...\n", mod.Name, mod.Health.URL)

	for time.Now().Before(deadline) {
		resp, err := client.Get(mod.Health.URL)
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode >= 200 && resp.StatusCode < 300 {
				fmt.Printf("[%s] pronto\n", mod.Name)
				return nil
			}
		}
		time.Sleep(interval)
	}

	return fmt.Errorf("módulo %q não ficou saudável após %s", mod.Name, timeout)
}

func parseDurationOrDefault(s string, def time.Duration) time.Duration {
	if s == "" {
		return def
	}
	d, err := time.ParseDuration(s)
	if err != nil {
		return def
	}
	return d
}

// ── Level execution ───────────────────────────────────────────────────────────

// execLevel executa fn para cada módulo em mods, em paralelo ou sequencialmente.
// fn recebe o diretório do módulo e o ModuleRef completo.
func execLevel(parentDir string, mods []manifest.ModuleRef, parallel bool, fn func(string, manifest.ModuleRef) error) error {
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
				if err := fn(modDir, mod); err != nil {
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
		if err := fn(modDir, mod); err != nil {
			return err
		}
	}
	return nil
}

// ── Topological sort ──────────────────────────────────────────────────────────

// computeLevels agrupa módulos em níveis de execução respeitando depends_on.
// Algoritmo de Kahn: nível 0 = sem deps, nível N = deps todos no nível < N.
// Retorna erro se detectar ciclo em depends_on.
func computeLevels(modules []manifest.ModuleRef) ([][]manifest.ModuleRef, error) {
	if len(modules) == 0 {
		return nil, nil
	}

	inDeg := make(map[string]int, len(modules))
	for _, m := range modules {
		inDeg[m.Name] = len(m.DependsOn)
	}

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
			var cyclic []string
			for _, m := range modules {
				if !processed[m.Name] {
					cyclic = append(cyclic, m.Name)
				}
			}
			return nil, fmt.Errorf("ciclo detectado em depends_on entre: %s", strings.Join(cyclic, ", "))
		}
		for _, m := range level {
			processed[m.Name] = true
			for _, dep := range dependents[m.Name] {
				inDeg[dep]--
			}
		}
		levels = append(levels, level)
	}

	return levels, nil
}
