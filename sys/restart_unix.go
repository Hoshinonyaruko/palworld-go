//go:build linux || darwin
// +build linux darwin

package sys

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"

	"github.com/hoshinonyaruko/palworld-go/config"
	"github.com/hoshinonyaruko/palworld-go/status"
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

func KillProcess(config config.Config) error {
	var cmd *exec.Cmd

	if runtime.GOOS == "windows" {
		// Windows: 直接指定要结束的进程名称
		cmd = exec.Command("taskkill", "/IM", "PalServer-Win64-Shipping-Cmd.exe", "/F")
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

// linux
func RestartService(config config.Config) {
	var exePath string
	var args []string

	// 对于非Windows系统的处理保持不变
	exePath = filepath.Join(config.GamePath, config.ProcessName+".sh")
	args = []string{
		"-RconEnabled=True",
		fmt.Sprintf("-AdminPassword=%s", config.WorldSettings.AdminPassword),
		fmt.Sprintf("-port=%d", config.WorldSettings.PublicPort),
		fmt.Sprintf("-players=%d", config.WorldSettings.ServerPlayerMaxNum),
	}

	if config.CommunityServer {
		args = append(args, "-publiclobby")
	}

	// 如果RCON启用，则添加RCON参数
	if config.WorldSettings.RconEnabled {
		args = append(args, "-rcon")
	}
	// 添加GameWorldSettings参数
	args = append(args, config.ServerOptions...)

	if config.GameService && config.GameServiceName != "" {
		cmd := exec.Command("sudo", "systemctl", "restart", config.GameServiceName)
		// 启动进程
		if err := cmd.Start(); err != nil {
			log.Printf("Failed to restart game server: %v", err)
		} else {
			log.Printf("Game server restarted successfully")
		}
	} else {
		// 执行启动命令
		log.Printf("启动命令: %s %s", exePath, strings.Join(args, " "))

		cmd := exec.Command(exePath, args...)
		cmd.Dir = config.GamePath // 设置工作目录为游戏路径

		// 启动进程
		if err := cmd.Start(); err != nil {
			log.Printf("Failed to restart game server: %v", err)
		} else {
			log.Printf("Game server restarted successfully")
		}

		// 获取并打印 PID
		log.Printf("Game server started successfully with PID %d", cmd.Process.Pid)
		status.SetGlobalPid(cmd.Process.Pid)
	}

}
