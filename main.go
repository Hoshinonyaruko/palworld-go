package main

import (
	"fmt"
)

func main() {
	// 读取或创建配置
	config := readConfig()

	// 打印配置以确认
	fmt.Printf("当前配置: %#v\n", config)

	// 设置监控和自动重启
	supervisor := NewSupervisor(config)
	go supervisor.Start()

	// 设置备份任务
	backupTask := NewBackupTask(config)
	go backupTask.Schedule()

	// 设置内存检查任务
	memoryCheckTask := NewMemoryCheckTask(config, backupTask)
	go memoryCheckTask.Schedule()

	// 主循环，等待用户输入或退出信号
	select {}

}
