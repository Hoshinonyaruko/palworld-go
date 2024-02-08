package status

import (
	"log"
	"os"
	"path/filepath"
	"strconv"

	"gopkg.in/ini.v1"
)

var (
	cfg *ini.File
)

func init() {
	// 获取当前工作目录
	currentDir, err := os.Getwd()
	if err != nil {
		log.Fatalf("Error getting current directory: %v", err)
	}

	// 定义配置文件路径为当前目录下的config.ini
	configPath := filepath.Join(currentDir, "config.ini")

	// 初始化时加载或创建配置文件
	cfg, err = ini.LoadSources(ini.LoadOptions{IgnoreInlineComment: true}, configPath)
	if err != nil {
		log.Printf("Fail to read file: %v", err)
		cfg = ini.Empty()
		// 尝试创建文件，因为可能是文件不存在导致的错误
		cfg.SaveTo(configPath)
	}
}

// saveConfig 保存更改到配置文件
func saveConfig() {
	// 在init中定义的configPath是局部变量，这里需要重新获取或定义全局变量
	currentDir, err := os.Getwd()
	if err != nil {
		log.Printf("Error getting current directory for saving config: %v", err)
		return
	}
	configPath := filepath.Join(currentDir, "config.ini")

	err = cfg.SaveTo(configPath)
	if err != nil {
		log.Printf("Fail to save config: %v", err)
	}
}

// SetMemoryIssueDetected 设置内存问题检测标志
func SetMemoryIssueDetected(flag bool) {
	cfg.Section("").Key("MemoryIssueDetected").SetValue(strconv.FormatBool(flag))
	saveConfig()
}

// GetMemoryIssueDetected 获取内存问题检测标志的当前值
func GetMemoryIssueDetected() bool {
	flag, err := cfg.Section("").Key("MemoryIssueDetected").Bool()
	if err != nil {
		return false
	}
	return flag
}

// SetsuccessReadGameWorldSettings 设置成功读取游戏世界设置标志
func SetsuccessReadGameWorldSettings(flag bool) {
	cfg.Section("").Key("SuccessReadGameWorldSettings").SetValue(strconv.FormatBool(flag))
	saveConfig()
}

// GetsuccessReadGameWorldSettings 获取成功读取游戏世界设置标志的当前值
func GetsuccessReadGameWorldSettings() bool {
	flag, err := cfg.Section("").Key("SuccessReadGameWorldSettings").Bool()
	if err != nil {
		return false
	}
	return flag
}

// SetManualServerShutdown 设置手动关闭服务器的状态
func SetManualServerShutdown(flag bool) {
	cfg.Section("").Key("ManualServerShutdown").SetValue(strconv.FormatBool(flag))
	saveConfig()
}

// GetManualServerShutdown 获取手动关闭服务器的状态
func GetManualServerShutdown() bool {
	flag, err := cfg.Section("").Key("ManualServerShutdown").Bool()
	if err != nil {
		return false
	}
	return flag
}

func SetGlobalPid(pid int) {
	cfg.Section("").Key("GlobalPid").SetValue(strconv.Itoa(pid))
	saveConfig()
}

func GetGlobalPid() int {
	pid, err := cfg.Section("").Key("GlobalPid").Int()
	if err != nil {
		return 0
	}
	return pid
}

func SetGlobalSubPid(pid int) {
	cfg.Section("").Key("GlobalSubPid").SetValue(strconv.Itoa(pid))
	saveConfig()
}

func GetGlobalSubPid() int {
	pid, err := cfg.Section("").Key("GlobalSubPid").Int()
	if err != nil {
		return 0
	}
	return pid
}
