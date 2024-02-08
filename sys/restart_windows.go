//go:build windows
// +build windows

package sys

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"
	"unsafe"

	"github.com/hoshinonyaruko/palworld-go/config"
	"github.com/hoshinonyaruko/palworld-go/mod"
	"github.com/hoshinonyaruko/palworld-go/status"
	"gopkg.in/ini.v1"
)

// WindowsRestarter implements the Restarter interface for Windows systems.
type WindowsRestarter struct{}

// NewRestarter creates a new Restarter appropriate for Windows systems.
func NewRestarter() *WindowsRestarter {
	return &WindowsRestarter{}
}

func (r *WindowsRestarter) Restart(executablePath string) error {
	// Separate the directory and the executable name
	execDir, execName := filepath.Split(executablePath)

	// Including -faststart parameter in the script that starts the executable
	scriptContent := "@echo off\n" +
		"pushd " + strconv.Quote(execDir) + "\n" +
		// Add the -faststart parameter here
		"start \"\" " + strconv.Quote(execName) + " -faststart\n" +
		"popd\n"

	scriptName := "restart.bat"
	if err := os.WriteFile(scriptName, []byte(scriptContent), 0755); err != nil {
		return err
	}

	cmd := exec.Command("cmd.exe", "/C", scriptName)

	if err := cmd.Start(); err != nil {
		return err
	}

	// The current process can now exit
	os.Exit(0)

	// This return statement will never be reached
	return nil
}

// windows
func setConsoleTitleWindows(title string) error {
	kernel32, err := syscall.LoadDLL("kernel32.dll")
	if err != nil {
		return err
	}
	proc, err := kernel32.FindProc("SetConsoleTitleW")
	if err != nil {
		return err
	}
	p0, err := syscall.UTF16PtrFromString(title)
	if err != nil {
		return err
	}
	r1, _, err := proc.Call(uintptr(unsafe.Pointer(p0)))
	if r1 == 0 {
		return err
	}
	return nil
}

func KillProcess() error {
	pid := status.GetGlobalPid()
	subPid := status.GetGlobalSubPid()

	fmt.Printf("获取到当前服务端进程pid:%v\n", pid)
	if pid == 0 {
		return fmt.Errorf("invalid PID: %d", pid)
	}

	// 结束主进程
	err := killByPid(pid)
	if err != nil {
		return err
	}

	// 如果存在SUBPID，则结束SUBPID
	if subPid != 0 {
		fmt.Printf("获取到子进程pid:%v\n", subPid)
		err := killByPid(subPid)
		if err != nil {
			return err
		}
	}

	return nil
}

func killByPid(pid int) error {
	var cmd *exec.Cmd

	if runtime.GOOS == "windows" {
		// Windows: 使用taskkill和PID结束进程
		cmd = exec.Command("taskkill", "/PID", strconv.Itoa(pid), "/F")
	} else {
		// 非Windows: 使用kill命令和PID结束进程
		cmd = exec.Command("kill", "-9", strconv.Itoa(pid))
	}

	cmd.SysProcAttr = &syscall.SysProcAttr{}
	if runtime.GOOS == "windows" {
		// 在Windows上隐藏命令行窗口
		cmd.SysProcAttr.HideWindow = true
	}

	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to kill process with PID %d: %v", pid, err)
	}

	fmt.Printf("成功结束进程 PID %d\n", pid)
	return nil
}

// RunViaBatch 函数接受配置，程序路径和参数数组
func RunViaBatch(config config.Config, exepath string, args []string) error {
	// 获取当前工作目录
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	// 创建批处理脚本内容
	batchScript := `@echo off
	start "" "` + exepath + `" ` + strings.Join(args, " ")

	// 指定批处理文件的路径
	batchFilePath := filepath.Join(cwd, "run_command.bat")

	// 写入批处理脚本到文件
	err = os.WriteFile(batchFilePath, []byte(batchScript), 0644)
	if err != nil {
		return err
	}

	// 执行批处理脚本
	cmd := exec.Command(batchFilePath)
	cmd.Dir = config.GamePath // 设置工作目录为游戏路径
	err = cmd.Run()
	if err != nil {
		log.Printf("RunViaBatch error : %v", err)
	}

	// 等待一段时间，确保C++程序有足够的时间写入PID到文件
	time.Sleep(1 * time.Second)

	err = parsePidFile(config)
	if err != nil {
		log.Fatalf("Error parsing pid.ini: %v", err)
	}
	return nil
}

func parsePidFile(config config.Config) error {
	pidFilePath := filepath.Join(config.GamePath, "pid.ini")

	// 使用ini库加载和解析文件
	cfg, err := ini.Load(pidFilePath)
	if err != nil {
		return fmt.Errorf("failed to read pid.ini: %v", err)
	}

	// 读取PID
	pidString := cfg.Section("").Key("PID").String()
	pid, err := strconv.Atoi(pidString)
	if err != nil {
		return fmt.Errorf("failed to convert PID string to int: %v", err)
	}
	log.Printf("Game server started successfully with PID %d", pid)
	status.SetGlobalPid(pid) // 存储转换后的PID

	// 读取SUBPID
	subPidString := cfg.Section("").Key("SUBPID").String()
	subPid, err := strconv.Atoi(subPidString)
	if err != nil {
		return fmt.Errorf("failed to convert SUBPID string to int: %v", err)
	}
	log.Printf("Subprocess started successfully with SUBPID %d", subPid)
	status.SetGlobalSubPid(subPid) // 存储转换后的SUBPID

	return nil
}

func RestartService(config config.Config) {
	var exePath string
	var args []string

	if config.CommunityServer {
		exePath = filepath.Join(config.SteamPath, "Steam.exe")
		args = []string{"-applaunch", "2394010"}
	} else if config.UseDll {
		err := mod.CheckAndWriteFiles(filepath.Join(config.GamePath, "Pal", "Binaries", "Win64"))
		if err != nil {
			log.Printf("Failed to write files: %v", err)
			return
		}
		exePath = filepath.Join(config.GamePath, "Pal", "Binaries", "Win64", "PalServerInject.exe")
		args = []string{
			"-RconEnabled=True",
			fmt.Sprintf("-AdminPassword=%s", config.WorldSettings.AdminPassword),
			fmt.Sprintf("-port=%d", config.WorldSettings.PublicPort),
			fmt.Sprintf("-players=%d", config.WorldSettings.ServerPlayerMaxNum),
		}
	} else {
		exePath = filepath.Join(config.GamePath, "Pal", "Binaries", "Win64", "PalServer-Win64-Test-Cmd.exe")
		//exePath = "\"" + exePath + "\""
		args = []string{
			"Pal",
			"-RconEnabled=True",
			fmt.Sprintf("-AdminPassword=%s", config.WorldSettings.AdminPassword),
			fmt.Sprintf("-port=%d", config.WorldSettings.PublicPort),
			fmt.Sprintf("-players=%d", config.WorldSettings.ServerPlayerMaxNum),
		}
	}

	args = append(args, config.ServerOptions...) // 添加GameWorldSettings参数

	// 执行启动命令
	log.Printf("启动命令: %s %s", exePath, strings.Join(args, " "))
	if config.UseDll && runtime.GOOS == "windows" {
		log.Printf("use bat")
		RunViaBatch(config, exePath, args)
		log.Printf("use bat success")
	} else {
		cmd := exec.Command(exePath, args...)
		cmd.Dir = config.GamePath // 设置工作目录为游戏路径
		if runtime.GOOS == "windows" {
			// 仅在Windows平台上设置
			cmd.SysProcAttr = &syscall.SysProcAttr{
				CreationFlags: 16,
			}
		}

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
