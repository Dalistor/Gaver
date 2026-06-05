package commands

import (
	"os"
	"os/exec"
)

func runShell(command string) error {
	c := exec.Command("sh", "-c", command)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	c.Stdin = os.Stdin
	return c.Run()
}
