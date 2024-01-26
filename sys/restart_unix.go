//go:build linux || darwin
// +build linux darwin

package sys

import (
	"os"
	"os/exec"
	"syscall"
)

// UnixRestarter implements the Restarter interface for Unix-like systems.
type UnixRestarter struct{}

// NewRestarter creates a new Restarter appropriate for Unix-like systems.
func NewRestarter() *UnixRestarter {
	return &UnixRestarter{}
}

// Restart restarts the application on Unix-like systems.
func (r *UnixRestarter) Restart(executableName string) error {
	scriptContent := "#!/bin/sh\n" +
		"sleep 1\n" + // Sleep for a bit to allow the main application to exit
		"." + executableName + "\n"

	scriptName := "restart.sh"
	if err := os.WriteFile(scriptName, []byte(scriptContent), 0755); err != nil {
		return err
	}

	cmd := exec.Command("/bin/sh", scriptName)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}

	if err := cmd.Start(); err != nil {
		return err
	}

	// The current process can now exit
	os.Exit(0)

	return nil
}
