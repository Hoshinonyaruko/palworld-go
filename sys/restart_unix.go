//go:build linux || darwin
// +build linux darwin

package sys

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"syscall"

	"github.com/hoshinonyaruko/palworld-go/config"
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
		executableName + "\n"

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

// windows
func setConsoleTitleWindows(title string) error {
	fmt.Printf("\033]0;%s\007", title)
	return nil
}

func KillProcess() error {
	var cmd *exec.Cmd

	if runtime.GOOS == "windows" {
		// Windows: 直接指定要结束的进程名称
		cmd = exec.Command("taskkill", "/IM", "PalServer-Win64-Test-Cmd.exe", "/F")
	} else {
		// 非Windows: 使用pkill命令和进程名称
		cmd = exec.Command("pkill", "-f", "PalServer-Linux-Test")
	}

	return cmd.Run()
}

// RunViaBatch 函数接受配置，程序路径和参数数组
func RunViaBatch(config config.Config, exepath string, args []string) error {
	return nil
}
