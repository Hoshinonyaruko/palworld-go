package main

import (
	"encoding/json"
	"fmt"
	"os"
)

type Config struct {
	GamePath             string  `json:"gamePath"`             // 游戏可执行文件路径PalServer.exe所处的位置
	GameSavePath         string  `json:"gameSavePath"`         // 游戏存档路径 \PalServer\Pal\Saved\文件夹的完整路径
	BackupPath           string  `json:"backupPath"`           // 备份路径
	Address              string  `json:"address"`              // 服务器 IP 地址
	RCONPort             string  `json:"rconPort"`             // RCON 端口号
	AdminPassword        string  `json:"adminPassword"`        // RCON 管理员密码
	ProcessName          string  `json:"processName"`          // 进程名称 PalServer
	CheckInterval        int     `json:"checkInterval"`        // 进程存活检查时间（秒）
	ServiceLogFile       string  `json:"serviceLogFile"`       // 日志文件路径
	ServiceErrorFile     string  `json:"serviceErrorFile"`     // 错误日志文件路径
	BackupInterval       int     `json:"backupInterval"`       // 备份间隔（秒）
	MemoryCheckInterval  int     `json:"memoryCheckInterval"`  // 内存占用检测时间（秒）
	MemoryUsageThreshold float64 `json:"memoryUsageThreshold"` // 重启阈值（百分比）
	TotalMemoryGB        int     `json:"totalMemoryGB"`        // 当前服务器总内存
}

// 默认配置
var defaultConfig = Config{
	GamePath:             "",
	GameSavePath:         "",
	BackupPath:           "",
	Address:              "127.0.0.1:25575",
	AdminPassword:        "default_password",
	ProcessName:          "PalServer",
	CheckInterval:        30, // 30 秒
	RCONPort:             "25575",
	ServiceLogFile:       "/service.log", // 示例路径
	ServiceErrorFile:     "/service.err", // 示例路径
	BackupInterval:       1800,           // 30 分钟
	MemoryCheckInterval:  30,             // 30 秒
	MemoryUsageThreshold: 80,             // 80%
	TotalMemoryGB:        16,             //16G
}

// 配置文件路径
const configFile = "config.json"

// readConfig 尝试读取配置文件，如果失败则创建默认配置
func readConfig() Config {
	var config Config

	// 读取配置文件
	data, err := os.ReadFile(configFile)
	if err != nil {
		fmt.Println("无法读取配置文件, 正在创建默认配置...")
		createDefaultConfig()
		return defaultConfig
	}

	// 反序列化JSON到结构体
	err = json.Unmarshal(data, &config)
	if err != nil {
		fmt.Println("配置解析失败, 正在使用默认配置...")
		return defaultConfig
	}

	// 确保所有必要的配置项都有值
	if config.Address == "" {
		config.Address = defaultConfig.Address
	}
	if config.AdminPassword == "" {
		config.AdminPassword = defaultConfig.AdminPassword
	}

	return config
}

// createDefaultConfig 创建一个带有默认值的配置文件
func createDefaultConfig() {
	data, err := json.MarshalIndent(defaultConfig, "", "    ")
	if err != nil {
		fmt.Println("无法创建默认配置文件:", err)
		os.Exit(1)
	}

	err = os.WriteFile(configFile, data, 0666)
	if err != nil {
		fmt.Println("无法写入默认配置文件:", err)
		os.Exit(1)
	}

	fmt.Println("默认配置文件已创建:", configFile)
}
