package main

import (
	"bytes"
	"fmt"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/hoshinonyaruko/palworld-go/config"
	"github.com/hoshinonyaruko/palworld-go/status"
	"github.com/hoshinonyaruko/palworld-go/sys"
)

type Supervisor struct {
	Config     config.Config
	RconClient RconClient
}

func NewSupervisor(config config.Config) *Supervisor {
	return &Supervisor{Config: config}
}

func (s *Supervisor) Start() {
	if s.Config.CheckInterval == 0 {
		fmt.Println("CheckInterval 设置为 0，不检查进程存活")
		return // 直接返回，不启动定时器
	}

	ticker := time.NewTicker(time.Duration(s.Config.CheckInterval) * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		// 在尝试重启服务之前检查是否手动关闭了服务器
		if status.GetManualServerShutdown() {
			fmt.Println("检测到服务器已手动关闭，不执行重启操作")
			continue // 跳过本次循环，不执行重启操作
		}

		if !s.isServiceRunning() {
			sys.RestartService(s.Config)
		} else {
			fmt.Println("当前正常运行中~")
		}
		if s.hasDefunct() {
			fmt.Printf("发现僵尸进程，准备清理~\n")
			// 此处只考虑僵尸进程是由自身内存释放导致的，如有其他原因，后续再patch
			sys.RestartApplication()
		}
	}
}

func (s *Supervisor) hasDefunct() bool {
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		return false
	} else {
		cmd = exec.Command("ps", "-ef")
	}
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		fmt.Printf("[ERROR] %v", err)
	}
	return strings.Contains(out.String(), "[PalServer.sh] <defunct>")
}

func (s *Supervisor) isServiceRunning() bool {
	pid := status.GetGlobalPid() // 假设这是从之前存储的地方获取PID
	if pid == 0 {
		return false
	}

	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("tasklist", "/FI", fmt.Sprintf("PID eq %d", pid))
	} else {
		cmd = exec.Command("ps", "-p", strconv.Itoa(pid))
	}

	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return false
	}

	if runtime.GOOS == "windows" {
		// 检查输出中是否有行包含PID，这意味着进程存在
		return strings.Contains(out.String(), strconv.Itoa(pid))
	} else {
		// Unix/Linux，如果`ps`命令找到了PID，它会返回成功
		return true
	}
}
