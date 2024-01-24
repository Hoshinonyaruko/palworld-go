package main

import (
	"embed"
	"fmt"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"runtime"
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
			log.Printf("你决定使用rammap清理内存....这会导致游戏卡顿")
			// 提取并保存RAMMap到临时文件
			rammapExecutable, err := extractRAMMapExecutable()
			if err != nil {
				log.Fatalf("无法提取RAMMap可执行文件: %v", err)
			}
			defer os.Remove(rammapExecutable) // 确保程序结束时删除文件

			// 创建定时器，根据配置间隔定期运行RAMMap
			ticker := time.NewTicker(time.Duration(config.MemoryCleanupInterval) * time.Second)
			defer ticker.Stop()

			for range ticker.C {
				runRAMMap(rammapExecutable)
			}
		}
	}

	// 主循环，等待用户输入或退出信号
	select {}

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
