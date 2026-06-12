package downloader

import (
	"fmt"
	"os"
	"os/exec"
)

// Clone clona um repositório Git em profundidade 1 para um diretório temporário.
// O caller deve invocar cleanup() via defer para remover o diretório após o uso.
func Clone(url string) (tmpDir string, cleanup func(), err error) {
	tmpDir, err = os.MkdirTemp("", "gaver-*")
	if err != nil {
		return "", nil, fmt.Errorf("falha ao criar diretório temporário: %w", err)
	}

	cleanup = func() { os.RemoveAll(tmpDir) }

	cmd := exec.Command("git", "clone", "--depth=1", url, tmpDir)
	cmd.Stderr = os.Stderr
	if err = cmd.Run(); err != nil {
		cleanup()
		return "", nil, fmt.Errorf("falha ao clonar %s: %w", url, err)
	}

	return tmpDir, cleanup, nil
}

// CloneTo clona um repositório Git em profundidade 1 diretamente para dest.
// Usado por gaver install para materializar sub-módulos no lugar definitivo.
func CloneTo(url, dest string) error {
	cmd := exec.Command("git", "clone", "--depth=1", url, dest)
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("falha ao clonar %s: %w", url, err)
	}
	return nil
}
