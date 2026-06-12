package downloader

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// Clone clona um repositório em profundidade 1 para um diretório temporário.
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

// CloneTo clona um repositório diretamente para dest.
// Se commit não for vazio, faz checkout desse commit exato (clone completo).
// Se commit for vazio, usa --depth=1 (clone rápido do HEAD).
func CloneTo(url, dest, commit string) error {
	var args []string
	if commit == "" {
		args = []string{"clone", "--depth=1", url, dest}
	} else {
		args = []string{"clone", url, dest}
	}

	cmd := exec.Command("git", args...)
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("falha ao clonar %s: %w", url, err)
	}

	if commit != "" {
		checkout := exec.Command("git", "checkout", commit)
		checkout.Dir = dest
		checkout.Stderr = os.Stderr
		if err := checkout.Run(); err != nil {
			return fmt.Errorf("falha ao fazer checkout de %s em %s: %w", commit, url, err)
		}
	}

	return nil
}

// HeadCommit retorna o SHA do commit HEAD do repositório em dir.
func HeadCommit(dir string) (string, error) {
	cmd := exec.Command("git", "rev-parse", "HEAD")
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("falha ao obter commit HEAD em %s: %w", dir, err)
	}
	return strings.TrimSpace(string(out)), nil
}
