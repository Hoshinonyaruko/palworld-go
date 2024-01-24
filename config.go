package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

type Config struct {
	GamePath                  string   `json:"gamePath"`                  // 游戏可执行文件路径PalServer.exe所处的位置
	GameSavePath              string   `json:"gameSavePath"`              // 游戏存档路径 \PalServer\Pal\Saved\文件夹的完整路径
	BackupPath                string   `json:"backupPath"`                // 备份路径
	Address                   string   `json:"address"`                   // 服务器 IP 地址
	RCONPort                  string   `json:"rconPort"`                  // RCON 端口号
	AdminPassword             string   `json:"adminPassword"`             // RCON 管理员密码
	ProcessName               string   `json:"processName"`               // 进程名称 PalServer
	CheckInterval             int      `json:"checkInterval"`             // 进程存活检查时间（秒）
	ServiceLogFile            string   `json:"serviceLogFile"`            // 日志文件路径
	ServiceErrorFile          string   `json:"serviceErrorFile"`          // 错误日志文件路径
	BackupInterval            int      `json:"backupInterval"`            // 备份间隔（秒）
	MemoryCheckInterval       int      `json:"memoryCheckInterval"`       // 内存占用检测时间（秒）
	MemoryUsageThreshold      float64  `json:"memoryUsageThreshold"`      // 重启阈值（百分比）
	TotalMemoryGB             int      `json:"totalMemoryGB"`             // 当前服务器总内存
	MemoryCleanupInterval     int      `json:"memoryCleanupInterval"`     // 内存清理时间间隔（秒）
	RegularMessages           []string `json:"regularMessages"`           // 定期推送的消息数组
	MessageBroadcastInterval  int      `json:"messageBroadcastInterval"`  // 消息广播周期（秒）
	MaintenanceWarningMessage string   `json:"maintenanceWarningMessage"` // 维护警告消息
}

// 默认配置
var defaultConfig = Config{
	GamePath:                  "",
	GameSavePath:              "",
	BackupPath:                "",
	Address:                   "127.0.0.1:25575",
	AdminPassword:             "default_password",
	ProcessName:               "PalServer",
	CheckInterval:             30, // 30 秒
	RCONPort:                  "25575",
	ServiceLogFile:            "/service.log",                                              // 示例路径
	ServiceErrorFile:          "/service.err",                                              // 示例路径
	BackupInterval:            1800,                                                        // 30 分钟
	MemoryCheckInterval:       30,                                                          // 30 秒
	MemoryUsageThreshold:      80,                                                          // 80%
	TotalMemoryGB:             16,                                                          //16G
	MemoryCleanupInterval:     0,                                                           // 内存清理时间间隔，设为半小时（1800秒）0代表不清理
	RegularMessages:           []string{"", ""},                                            // 默认的定期推送消息数组，初始可为空
	MessageBroadcastInterval:  3600,                                                        // 默认消息广播周期，假设为1小时（3600秒）
	MaintenanceWarningMessage: "server is going to rebot,please relogin at 1minute later.", // 默认的维护警告消息
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
	err = AutoConfigurePaths(&config)
	if err != nil {
		log.Fatalf("配置错误: %v", err)
	}
	var write bool
	// 确保所有必要的配置项都有默认值
	if config.Address == "" {
		config.Address = defaultConfig.Address
		write = true
	}
	if config.AdminPassword == "" {
		config.AdminPassword = defaultConfig.AdminPassword
		write = true
	}
	if config.MemoryCleanupInterval == 0 {
		config.MemoryCleanupInterval = 0
	}
	if config.RegularMessages == nil {
		config.RegularMessages = []string{"", ""}
		write = true
	}
	if config.MessageBroadcastInterval == 0 {
		config.MessageBroadcastInterval = 3600
		write = true
	}
	if config.MaintenanceWarningMessage == "" {
		config.MaintenanceWarningMessage = "服务器即将进行维护,你的存档已保存,请放心,请坐稳扶好,1分钟后重新登录。"
		write = true
	}
	//写回本地
	if write {
		writeConfigToFile(config)
	}

	return config
}

// writeConfigToFile 将配置写回文件
func writeConfigToFile(config Config) {
	configJSON, err := json.MarshalIndent(config, "", "    ")
	if err != nil {
		log.Fatalf("无法序列化配置: %v", err)
	}

	err = os.WriteFile(configFile, configJSON, 0644)
	if err != nil {
		log.Fatalf("无法写入配置文件: %v", err)
	}
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

// AutoConfigurePaths 自动配置路径
func AutoConfigurePaths(config *Config) error {
	currentDir, err := os.Getwd()
	if err != nil {
		return err
	}

	exePath := filepath.Join(currentDir, "PalServer.exe")
	if _, err := os.Stat(exePath); os.IsNotExist(err) {
		log.Println("检测到与PalServer.exe不位于同一路径下,建议将程序放置在PalServer.exe同目录下")
		return nil
	}

	correctGamePath := currentDir
	correctGameSavePath := filepath.Join(currentDir, "Pal\\Saved")

	// 检查路径是否需要更新
	if config.GamePath != correctGamePath || config.GameSavePath != correctGameSavePath {
		config.GamePath = correctGamePath
		config.GameSavePath = correctGameSavePath

		// 将更新后的配置写回文件
		updatedConfig, err := json.MarshalIndent(config, "", "  ")
		if err != nil {
			return err
		}

		err = os.WriteFile("config.json", updatedConfig, 0644)
		if err != nil {
			return err
		}

		log.Println("你的目录配置已被自动修正,请重新运行本程序。")
	} else {
		log.Println("你的目录配置正确。")
	}

	return nil
}
