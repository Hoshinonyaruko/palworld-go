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
	var exePath string
	var args []string

	// 构造游戏启动参数
	//gameArgs := constructGameLaunchArguments(s.Config.WorldSettings)

	if runtime.GOOS == "windows" {
		exePath = filepath.Join(s.Config.GamePath, s.Config.ProcessName+".exe")
		args = []string{
			"-useperfthreads",
			"-NoAsyncLoadingThread",
			"-UseMultithreadForDS",
			"RconEnabled=True",
			fmt.Sprintf("AdminPassword=%s", s.Config.WorldSettings.AdminPassword),
			fmt.Sprintf("port=%d", s.Config.WorldSettings.PublicPort),
			fmt.Sprintf("players=%d", s.Config.WorldSettings.ServerPlayerMaxNum),
		}
		//args = append(args, gameArgs...) // 添加GameWorldSettings参数
	} else {
		exePath = filepath.Join(s.Config.GamePath, s.Config.ProcessName+".sh")
		args = []string{ // Linux下可能需要不同的参数
			fmt.Sprintf("--port=%d", s.Config.WorldSettings.PublicPort),
			fmt.Sprintf("--players=%d", s.Config.WorldSettings.ServerPlayerMaxNum),
		}
	}

	// 执行启动命令
	log.Printf("启动命令: %s %s", exePath, strings.Join(args, " "))
	cmd := exec.Command(exePath, args...)
	cmd.Dir = s.Config.GamePath // 设置工作目录为游戏路径

	// 启动进程
	if err := cmd.Start(); err != nil {
		log.Printf("Failed to restart game server: %v", err)
	} else {
		log.Printf("Game server restarted successfully")
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
