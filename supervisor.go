package main

import (
	"bytes"
	"fmt"
	"log"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/hoshinonyaruko/palworld-go/config"
	"github.com/hoshinonyaruko/palworld-go/mod"
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
			s.restartService()
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

func (s *Supervisor) restartService() {
	var exePath string
	var args []string

	if runtime.GOOS == "windows" {
		if s.Config.CommunityServer {
			exePath = filepath.Join(s.Config.SteamPath, "Steam.exe")
			args = []string{"-applaunch", "2394010"}
		} else if s.Config.UseDll {
			err := mod.CheckAndWriteFiles(filepath.Join(s.Config.GamePath, "Pal", "Binaries", "Win64"))
			if err != nil {
				log.Printf("Failed to write files: %v", err)
				return
			}
			exePath = filepath.Join(s.Config.GamePath, "Pal", "Binaries", "Win64", "PalServerInject.exe")
			args = []string{
				"-RconEnabled=True",
				fmt.Sprintf("-AdminPassword=%s", s.Config.WorldSettings.AdminPassword),
				fmt.Sprintf("-port=%d", s.Config.WorldSettings.PublicPort),
				fmt.Sprintf("-players=%d", s.Config.WorldSettings.ServerPlayerMaxNum),
			}
		} else {
			exePath = filepath.Join(s.Config.GamePath, "Pal", "Binaries", "Win64", "PalServer-Win64-Test-Cmd.exe")
			//exePath = "\"" + exePath + "\""
			args = []string{
				"Pal",
				"-RconEnabled=True",
				fmt.Sprintf("-AdminPassword=%s", s.Config.WorldSettings.AdminPassword),
				fmt.Sprintf("-port=%d", s.Config.WorldSettings.PublicPort),
				fmt.Sprintf("-players=%d", s.Config.WorldSettings.ServerPlayerMaxNum),
			}
		}
	} else {
		// 对于非Windows系统的处理保持不变
		exePath = filepath.Join(s.Config.GamePath, s.Config.ProcessName+".sh")
		args = []string{
			"-RconEnabled=True",
			fmt.Sprintf("-AdminPassword=%s", s.Config.WorldSettings.AdminPassword),
			fmt.Sprintf("-port=%d", s.Config.WorldSettings.PublicPort),
			fmt.Sprintf("-players=%d", s.Config.WorldSettings.ServerPlayerMaxNum),
		}
	}

	args = append(args, s.Config.ServerOptions...) // 添加GameWorldSettings参数

	// 执行启动命令
	log.Printf("启动命令: %s %s", exePath, strings.Join(args, " "))
	if s.Config.UseDll && runtime.GOOS == "windows" {
		log.Printf("use bat")
		sys.RunViaBatch(s.Config, exePath, args)
		log.Printf("use bat success")
	} else {
		cmd := exec.Command(exePath, args...)
		cmd.Dir = s.Config.GamePath // 设置工作目录为游戏路径
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

// func constructGameLaunchArguments(settings *GameWorldSettings) []string {
// 	var args []string

// 	sValue := reflect.ValueOf(settings).Elem()
// 	sType := sValue.Type()

// 	for i := 0; i < sType.NumField(); i++ {
// 		field := sType.Field(i)
// 		fieldValue := sValue.Field(i)

// 		jsonTag := firstToUpper(strings.Split(field.Tag.Get("json"), ",")[0]) // 获取json标签的第一部分，并将首字母转换为大写

// 		var arg string
// 		switch fieldValue.Kind() {
// 		case reflect.String:
// 			arg = fmt.Sprintf("%s=%s", jsonTag, fieldValue.String())
// 		case reflect.Float64:
// 			arg = fmt.Sprintf("%s=%s", jsonTag, strconv.FormatFloat(fieldValue.Float(), 'f', 6, 64))
// 		case reflect.Int:
// 			arg = fmt.Sprintf("%s=%d", jsonTag, fieldValue.Int())
// 		case reflect.Bool:
// 			arg = fmt.Sprintf("%s=%t", jsonTag, fieldValue.Bool())
// 		}

// 		args = append(args, arg)
// 	}

// 	return args
// }
