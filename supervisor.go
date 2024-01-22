package main

import (
	"bytes"
	"fmt"
	"log"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

type Supervisor struct {
	Config     Config
	RconClient RconClient
}

func NewSupervisor(config Config) *Supervisor {
	return &Supervisor{Config: config}
}

func (s *Supervisor) Start() {
	ticker := time.NewTicker(time.Duration(s.Config.CheckInterval) * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		if !s.isServiceRunning() {
			s.restartService()
		} else {
			fmt.Printf("当前正常运行中~\n")
		}
	}
}

func (s *Supervisor) isServiceRunning() bool {
	var cmd *exec.Cmd

	if runtime.GOOS == "windows" {
		cmd = exec.Command("tasklist")
	} else {
		// Unix/Linux，假设'pgrep'可用
		cmd = exec.Command("pgrep", "-f", s.Config.ProcessName)
	}

	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return false
	}

	if runtime.GOOS == "windows" {
		// Windows，检查任务列表输出中是否包含进程名
		return strings.Contains(out.String(), s.Config.ProcessName)
	}

	// Unix/Linux，假如'pgrep'找到了进程，它会返回成功
	return true
}

func (s *Supervisor) restartService() {
	// 构建游戏服务器的启动命令
	var command string
	if runtime.GOOS == "windows" {
		command = filepath.Join(s.Config.GamePath, s.Config.ProcessName+".exe")
	} else {
		command = filepath.Join(s.Config.GamePath, s.Config.ProcessName)
	}

	// 执行启动命令
	cmd := exec.Command(command)
	cmd.Dir = s.Config.GamePath // 设置工作目录为游戏路径

	// 启动进程
	if err := cmd.Start(); err != nil {
		log.Printf("Failed to restart game server: %v", err)
	} else {
		log.Printf("Game server restarted successfully")
	}
}
