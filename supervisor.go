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

	"github.com/hoshinonyaruko/palworld-go/config"
	"github.com/hoshinonyaruko/palworld-go/mod"
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
	ticker := time.NewTicker(time.Duration(s.Config.CheckInterval) * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		if !s.isServiceRunning() {
			s.restartService()
		} else {
			fmt.Printf("当前正常运行中~\n")
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
	var cmd *exec.Cmd

	if runtime.GOOS == "windows" {
		cmd = exec.Command("tasklist")
	} else {
		// Unix/Linux，假设'pgrep'可用
		cmd = exec.Command("pgrep", "-f", s.Config.ProcessName+".sh")
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
			exePath = filepath.Join(s.Config.GamePath, s.Config.ProcessName+".exe")
			args = []string{
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
	} else {
		cmd := exec.Command(exePath, args...)
		cmd.Dir = s.Config.GamePath // 设置工作目录为游戏路径

		// 启动进程
		if err := cmd.Start(); err != nil {
			log.Printf("Failed to restart game server: %v", err)
		} else {
			log.Printf("Game server restarted successfully")
		}
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
