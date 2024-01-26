package main

import (
	"embed"
	"fmt"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"syscall"
	"time"
)

//go:embed RAMMap64.exe
var rammapFS embed.FS

func main() {
	// 读取或创建配置
	config := readConfig()

	// 打印配置以确认
	fmt.Printf("当前配置: %#v\n", config)
	fmt.Printf("作者 早苗狐 答疑群:587997911\n")

	// 设置监控和自动重启
	supervisor := NewSupervisor(config)
	go supervisor.Start()

	if !supervisor.isServiceRunning() {
		supervisor.restartService()
	} else {
		fmt.Printf("当前服务端正常运行中,守护和内存助手已启动\n")
	}

	// 设置备份任务
	backupTask := NewBackupTask(config)
	go backupTask.Schedule()

	// 设置推送任务
	palworldBroadcast := NewpalworldBroadcast(config)
	go palworldBroadcast.Schedule()

	// 设置内存检查任务
	memoryCheckTask := NewMemoryCheckTask(config, backupTask)
	go memoryCheckTask.Schedule()

	if runtime.GOOS == "windows" {
		if config.MemoryCleanupInterval != 0 {
			log.Printf("你决定使用rammap清理内存....这不会导致游戏卡顿")

			// 提取并保存RAMMap到临时文件
			rammapExecutable, err := extractRAMMapExecutable()
			if err != nil {
				log.Fatalf("无法提取RAMMap可执行文件: %v", err)
			}
			defer os.Remove(rammapExecutable) // 确保程序结束时删除文件

			// 创建定时器，根据配置间隔定期运行RAMMap
			ticker := time.NewTicker(time.Duration(config.MemoryCleanupInterval) * time.Second)
			go func() {
				defer ticker.Stop()
				for range ticker.C {
					runRAMMap(rammapExecutable)
				}
			}()
		}
	}

	if runtime.GOOS == "windows" {
		// 创建一个定时器，每10秒触发一次，保存游戏设置
		saveSettingsTicker := time.NewTicker(10 * time.Second)
		go func() {
			defer saveSettingsTicker.Stop()
			for range saveSettingsTicker.C {
				// 定时保存配置
				config := readConfigv2()
				err := writeGameWorldSettings(&config, config.WorldSettings)
				if err != nil {
					fmt.Println("Error writing game world settings:", err)
				} else {
					fmt.Println("Game world settings saved successfully.")
				}
			}
		}()
	}

	// 设置信号捕获
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// 等待信号
	<-sigChan
	if runtime.GOOS == "windows" {
		// 接收到退出信号，写回配置
		err := writeGameWorldSettings(&config, config.WorldSettings)
		if err != nil {
			// 处理写回错误
			fmt.Println("Error writing game world settings:", err)
		} else {
			fmt.Println("Success writing game world settings")
		}
	}

	// 正常退出程序
	os.Exit(0)

}

// extractRAMMapExecutable 从嵌入的文件系统中提取RAMMap并写入临时文件
func extractRAMMapExecutable() (string, error) {
	rammapData, err := fs.ReadFile(rammapFS, "RAMMap64.exe")
	if err != nil {
		return "", err
	}

	tmpFile, err := os.CreateTemp("", "RAMMap64-*.exe")
	if err != nil {
		return "", err
	}
	defer tmpFile.Close()

	if _, err := tmpFile.Write(rammapData); err != nil {
		return "", err
	}

	return tmpFile.Name(), nil
}

func runRAMMap(rammapExecutable string) {
	log.Printf("正在使用rammap清理内存....")
	// 调用RAMMap的命令
	cmd := exec.Command(rammapExecutable, "-Ew")
	err := cmd.Run()
	if err != nil {
		log.Printf("运行RAMMap时发生错误: %v", err)
	}
}
