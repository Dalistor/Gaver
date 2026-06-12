package pidstore

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func pidsDir(rootDir string) string {
	return filepath.Join(rootDir, ".gaver", "pids")
}

func pidFile(rootDir, name string) string {
	return filepath.Join(pidsDir(rootDir), name+".pid")
}

func Save(rootDir, name string, pid int) error {
	if err := os.MkdirAll(pidsDir(rootDir), 0755); err != nil {
		return err
	}
	return os.WriteFile(pidFile(rootDir, name), []byte(strconv.Itoa(pid)), 0644)
}

func Remove(rootDir, name string) {
	os.Remove(pidFile(rootDir, name))
}

// List retorna todos os módulos com PIDs salvos: nome → pid.
func List(rootDir string) map[string]int {
	result := make(map[string]int)
	entries, err := os.ReadDir(pidsDir(rootDir))
	if err != nil {
		return result
	}
	for _, e := range entries {
		if !strings.HasSuffix(e.Name(), ".pid") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(pidsDir(rootDir), e.Name()))
		if err != nil {
			continue
		}
		pid, err := strconv.Atoi(strings.TrimSpace(string(data)))
		if err != nil {
			continue
		}
		name := strings.TrimSuffix(e.Name(), ".pid")
		result[name] = pid
	}
	return result
}
