package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"unicode"

	"github.com/hoshinonyaruko/palworld-go/status"
	"gopkg.in/ini.v1"
)

type Config struct {
	GamePath                  string             `json:"gamePath"`                  // 游戏可执行文件路径PalServer.exe所处的位置
	GameSavePath              string             `json:"gameSavePath"`              // 游戏存档路径 \PalServer\Pal\Saved\文件夹的完整路径
	BackupPath                string             `json:"backupPath"`                // 备份路径
	SteamPath                 string             `json:"steamPath"`                 // steam路径
	CommunityServer           bool               `json:"communityServer"`           // 社区服务器开关
	UseDll                    bool               `json:"useDll"`                    // dll注入
	Address                   string             `json:"address"`                   // 服务器 IP 地址
	UseHttps                  bool               `json:"usehttps"`                  // 使用 https
	WebuiPort                 string             `json:"webuiPort"`                 // Webui 端口号
	AutolaunchWebui           bool               `json:"autoLaunchWebui"`           // 自动打开webui
	ProcessName               string             `json:"processName"`               // 进程名称 PalServer
	ServerOptions             []string           `json:"serverOptions"`             // 服务器启动参数
	CheckInterval             int                `json:"checkInterval"`             // 进程存活检查时间（秒）
	BackupInterval            int                `json:"backupInterval"`            // 备份间隔（秒）
	RestartInterval           int                `json:"RestartInterval"`           // 自动重启服务器（秒）
	MemoryCheckInterval       int                `json:"memoryCheckInterval"`       // 内存占用检测时间（秒）
	MemoryUsageThreshold      float64            `json:"memoryUsageThreshold"`      // 重启阈值（百分比）
	TotalMemoryGB             int                `json:"totalMemoryGB"`             // 当前服务器总内存
	MemoryCleanupInterval     int                `json:"memoryCleanupInterval"`     // 内存清理时间间隔（秒）
	RegularMessages           []string           `json:"regularMessages"`           // 定期推送的消息数组
	MessageBroadcastInterval  int                `json:"messageBroadcastInterval"`  // 消息广播周期（秒）
	MaintenanceWarningMessage string             `json:"maintenanceWarningMessage"` // 维护警告消息
	WorldSettings             *GameWorldSettings `json:"worldSettings"`             // 帕鲁设定
	Engine                    *Engine            `json:"engine"`                    // 服务端引擎设置
}

// 默认配置
var defaultConfig = Config{
	GamePath:                  "",
	GameSavePath:              "",
	BackupPath:                "",
	SteamPath:                 "",
	CommunityServer:           false,
	Address:                   "127.0.0.1",
	UseHttps:                  false,
	ProcessName:               "PalServer",
	UseDll:                    false,
	ServerOptions:             []string{"-useperfthreads", "-NoAsyncLoadingThread", "-UseMultithreadForDS"},
	CheckInterval:             30,     // 30 秒
	WebuiPort:                 "8000", // Webui 端口号
	AutolaunchWebui:           false,
	BackupInterval:            1800,                                                        // 30 分钟
	MemoryCheckInterval:       60,                                                          // 60 秒
	MemoryUsageThreshold:      90,                                                          // 90%
	TotalMemoryGB:             16,                                                          // 16G
	MemoryCleanupInterval:     1800,                                                        // 内存清理时间间隔，设为半小时（1800秒）0代表不清理
	RestartInterval:           0,                                                           // 自动重启间隔
	RegularMessages:           []string{""},                                                // 默认的定期推送消息数组，初始可为空
	MessageBroadcastInterval:  3600,                                                        // 默认消息广播周期，假设为1小时（3600秒）
	MaintenanceWarningMessage: "server is going to rebot,please relogin at 1minute later.", // 默认的维护警告消息
}

// Engine 默认配置
var defaultEngine = Engine{
	Player: PlayerConfig{
		ConfiguredInternetSpeed: 104857600,
		ConfiguredLanSpeed:      104857600,
	},
	SocketSubsystemEpic: SocketSubsystemEpicConfig{
		MaxClientRate:         104857600,
		MaxInternetClientRate: 104857600,
	},
	EngineConfig: EngineConfig{
		BSmoothFrameRate:        true,
		BUseFixedFrameRate:      false,
		SmoothedFrameRateRange:  FrameRateRange{LowerBound: Bound{Type: "Inclusive", Value: 30.0}, UpperBound: Bound{Type: "Exclusive", Value: 60.0}},
		MinDesiredFrameRate:     30.0,
		FixedFrameRate:          120.0,
		NetClientTicksPerSecond: 120,
	},
}

type GameWorldSettings struct {
	Difficulty                          string  `json:"difficulty"`
	DayTimeSpeedRate                    float64 `json:"dayTimeSpeedRate"`
	NightTimeSpeedRate                  float64 `json:"nightTimeSpeedRate"`
	ExpRate                             float64 `json:"expRate"`
	PalCaptureRate                      float64 `json:"palCaptureRate"`
	PalSpawnNumRate                     float64 `json:"palSpawnNumRate"`
	PalDamageRateAttack                 float64 `json:"palDamageRateAttack"`
	PalDamageRateDefense                float64 `json:"palDamageRateDefense"`
	PlayerDamageRateAttack              float64 `json:"playerDamageRateAttack"`
	PlayerDamageRateDefense             float64 `json:"playerDamageRateDefense"`
	PlayerStomachDecreaceRate           float64 `json:"playerStomachDecreaceRate"`
	PlayerStaminaDecreaceRate           float64 `json:"playerStaminaDecreaceRate"`
	PlayerAutoHPRegeneRate              float64 `json:"playerAutoHPRegeneRate"`
	PlayerAutoHpRegeneRateInSleep       float64 `json:"playerAutoHpRegeneRateInSleep"`
	PalStomachDecreaceRate              float64 `json:"palStomachDecreaceRate"`
	PalStaminaDecreaceRate              float64 `json:"palStaminaDecreaceRate"`
	PalAutoHPRegeneRate                 float64 `json:"palAutoHPRegeneRate"`
	PalAutoHpRegeneRateInSleep          float64 `json:"palAutoHpRegeneRateInSleep"`
	BuildObjectDamageRate               float64 `json:"buildObjectDamageRate"`
	BuildObjectDeteriorationDamageRate  float64 `json:"buildObjectDeteriorationDamageRate"`
	CollectionDropRate                  float64 `json:"collectionDropRate"`
	CollectionObjectHpRate              float64 `json:"collectionObjectHpRate"`
	CollectionObjectRespawnSpeedRate    float64 `json:"collectionObjectRespawnSpeedRate"`
	EnemyDropItemRate                   float64 `json:"enemyDropItemRate"`
	DeathPenalty                        string  `json:"deathPenalty"`
	EnablePlayerToPlayerDamage          bool    `json:"enablePlayerToPlayerDamage"`
	EnableFriendlyFire                  bool    `json:"enableFriendlyFire"`
	EnableInvaderEnemy                  bool    `json:"enableInvaderEnemy"`
	ActiveUNKO                          bool    `json:"activeUNKO"`
	EnableAimAssistPad                  bool    `json:"enableAimAssistPad"`
	EnableAimAssistKeyboard             bool    `json:"enableAimAssistKeyboard"`
	DropItemMaxNum                      int     `json:"dropItemMaxNum"`
	DropItemMaxNum_UNKO                 int     `json:"dropItemMaxNum_UNKO"`
	BaseCampMaxNum                      int     `json:"baseCampMaxNum"`
	BaseCampWorkerMaxNum                int     `json:"baseCampWorkerMaxNum"`
	DropItemAliveMaxHours               float64 `json:"dropItemAliveMaxHours"`
	AutoResetGuildNoOnlinePlayers       bool    `json:"autoResetGuildNoOnlinePlayers"`
	AutoResetGuildTimeNoOnlinePlayers   float64 `json:"autoResetGuildTimeNoOnlinePlayers"`
	GuildPlayerMaxNum                   int     `json:"guildPlayerMaxNum"`
	PalEggDefaultHatchingTime           float64 `json:"palEggDefaultHatchingTime"`
	WorkSpeedRate                       float64 `json:"workSpeedRate"`
	IsMultiplay                         bool    `json:"isMultiplay"`
	IsPvP                               bool    `json:"isPvP"`
	CanPickupOtherGuildDeathPenaltyDrop bool    `json:"canPickupOtherGuildDeathPenaltyDrop"`
	EnableNonLoginPenalty               bool    `json:"enableNonLoginPenalty"`
	EnableFastTravel                    bool    `json:"enableFastTravel"`
	IsStartLocationSelectByMap          bool    `json:"isStartLocationSelectByMap"`
	ExistPlayerAfterLogout              bool    `json:"existPlayerAfterLogout"`
	EnableDefenseOtherGuildPlayer       bool    `json:"enableDefenseOtherGuildPlayer"`
	CoopPlayerMaxNum                    int     `json:"coopPlayerMaxNum"`
	ServerPlayerMaxNum                  int     `json:"serverPlayerMaxNum"`
	ServerName                          string  `json:"serverName"`
	ServerDescription                   string  `json:"serverDescription"`
	AdminPassword                       string  `json:"adminPassword"`
	ServerPassword                      string  `json:"serverPassword"`
	PublicPort                          int     `json:"publicPort"`
	PublicIP                            string  `json:"publicIP"`
	RconEnabled                         bool    `json:"rconEnabled"`
	RconPort                            int     `json:"rconPort"`
	Region                              string  `json:"region"`
	UseAuth                             bool    `json:"useAuth"`
	BanListURL                          string  `json:"banListURL"`
}

// Engine is the top-level structure representing the configuration.
type Engine struct {
	Player              PlayerConfig              `json:"player"`
	SocketSubsystemEpic SocketSubsystemEpicConfig `json:"socketsubsystemepic"`
	EngineConfig        EngineConfig              `json:"engine"`
}

// PlayerConfig represents the configuration for the player.
type PlayerConfig struct {
	ConfiguredInternetSpeed int `json:"ConfiguredInternetSpeed"`
	ConfiguredLanSpeed      int `json:"ConfiguredLanSpeed"`
}

// SocketSubsystemEpicConfig represents the configuration for the socket subsystem.
type SocketSubsystemEpicConfig struct {
	MaxClientRate         int `json:"MaxClientRate"`
	MaxInternetClientRate int `json:"MaxInternetClientRate"`
}

// EngineConfig represents the configuration for the engine.
type EngineConfig struct {
	BSmoothFrameRate        bool           `json:"bSmoothFrameRate"`
	BUseFixedFrameRate      bool           `json:"bUseFixedFrameRate"`
	SmoothedFrameRateRange  FrameRateRange `json:"SmoothedFrameRateRange"`
	MinDesiredFrameRate     float64        `json:"MinDesiredFrameRate"`
	FixedFrameRate          float64        `json:"FixedFrameRate"`
	NetClientTicksPerSecond int            `json:"NetClientTicksPerSecond"`
}

// FrameRateRange represents a range of frame rates.
type FrameRateRange struct {
	LowerBound Bound `json:"LowerBound"`
	UpperBound Bound `json:"UpperBound"`
}

// Bound represents a boundary in a frame rate range.
type Bound struct {
	Type  string  `json:"Type"`
	Value float64 `json:"Value"`
}

// 配置文件路径
const configFile = "config.json"

// readConfig 尝试读取配置文件，如果失败则创建并自动配置默认配置
func ReadConfig() Config {
	var config Config

	data, err := os.ReadFile(configFile)
	if err != nil {
		fmt.Println("无法读取配置文件, 正在创建默认配置...")
		config = createDefaultConfig()
	} else {
		err = json.Unmarshal(data, &config)
		if err != nil {
			fmt.Println("配置解析失败, 正在使用默认配置...")
			config = defaultConfig
		}
	}

	// 自动配置路径
	err = AutoConfigurePaths(&config)
	if err != nil {
		log.Fatalf("路径配置错误: %v", err)
	}

	// 检查并设置默认值
	if checkAndSetDefaults(&config) {
		// 如果配置被修改，写回文件
		writeConfigToFile(config)
	}

	return config
}

// ReadConfigv2 尝试读取配置文件，如果失败则创建并自动配置默认配置
func ReadConfigv2() Config {
	var config Config

	data, err := os.ReadFile(configFile)
	if err != nil {
		fmt.Println("无法读取配置文件, 正在创建默认配置...")
		config = createDefaultConfig()
	} else {
		err = json.Unmarshal(data, &config)
		if err != nil {
			fmt.Println("配置解析失败, 正在使用默认配置...")
			config = defaultConfig
		}
	}

	return config
}

// checkAndSetDefaults 检查并设置默认值，返回是否做了修改
func checkAndSetDefaults(config *Config) bool {
	// 通过反射获取Config的类型和值
	val := reflect.ValueOf(config).Elem()
	typ := val.Type()

	// 记录是否进行了修改
	var modified bool

	// 遍历所有字段
	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		defaultField := reflect.ValueOf(defaultConfig).Field(i)
		fieldType := field.Type()

		// 跳过布尔类型的字段
		if fieldType.Kind() == reflect.Bool {
			continue
		}

		fieldName := typ.Field(i).Name

		// 特殊处理RestartInterval字段
		if fieldName == "RestartInterval" {
			continue
		}

		// 如果字段是零值，设置为默认值
		if isZeroOfUnderlyingType(field.Interface()) {
			field.Set(defaultField)
			modified = true
		}
	}

	// 如果BackupPath为空，则设置为gamePath\backup
	if config.BackupPath == "" {
		config.BackupPath = filepath.Join(config.GamePath, "backup")
		fmt.Printf("未设置备份目录，自动设置为：%s\n", config.BackupPath)
		modified = true
	}
	// 新逻辑：根据GamePath自动设置SteamPath为GamePath的上三级目录
	// 这个写法更健壮
	if config.GamePath != "" {
		// 调用三次filepath.Dir获取上三级目录
		steamPath := filepath.Dir(filepath.Dir(filepath.Dir(config.GamePath)))
		if config.SteamPath != steamPath {
			config.SteamPath = steamPath
			fmt.Printf("SteamPath自动设置为：%s\n", config.SteamPath)
			modified = true
		}
	}

	return modified
}

// isZeroOfUnderlyingType 检查一个值是否为其类型的零值
func isZeroOfUnderlyingType(x interface{}) bool {
	return reflect.DeepEqual(x, reflect.Zero(reflect.TypeOf(x)).Interface())
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

// createDefaultConfig 创建一个带有默认值的配置文件，并返回这个配置
func createDefaultConfig() Config {
	// 序列化默认配置
	data, err := json.MarshalIndent(defaultConfig, "", "    ")
	if err != nil {
		fmt.Println("无法创建默认配置文件:", err)
		os.Exit(1)
	}

	// 将默认配置写入文件
	err = os.WriteFile(configFile, data, 0666)
	if err != nil {
		fmt.Println("无法写入默认配置文件:", err)
		os.Exit(1)
	}

	fmt.Println("默认配置文件已创建:", configFile)

	// 返回默认配置
	return defaultConfig
}

// AutoConfigurePaths 自动配置路径
func AutoConfigurePaths(config *Config) error {
	currentDir, err := os.Getwd()
	if err != nil {
		return err
	}

	// 根据操作系统设置可执行文件的名称
	exeName := "PalServer"
	if runtime.GOOS == "windows" {
		exeName += ".exe"
	} else {
		exeName += ".sh"
	}

	exePath := filepath.Join(currentDir, exeName)
	if _, err := os.Stat(exePath); os.IsNotExist(err) {
		log.Printf("检测到 %s 不位于同一路径下, 建议将程序放置在 %s 同目录下\n", exeName, exeName)
		return nil
	}

	correctGamePath := currentDir
	correctGameSavePath := filepath.Join(currentDir, "Pal", "Saved")

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

	//这里刷新 已经区分不同的操作系统
	gameworldsettings, err := ReadGameWorldSettings(config)
	if err != nil {
		log.Printf("解析游戏parworldsetting出错,错误%v", err)
		status.SetsuccessReadGameWorldSettings(false)
	} else {
		config.WorldSettings = gameworldsettings
		log.Println("从游戏parworldsetting.ini解析配置成功.")
		log.Printf("从游戏parworldsetting.ini解析配置成功.%v", config.WorldSettings)
		status.SetsuccessReadGameWorldSettings(true)
		// 将更新后的配置写回文件
		updatedConfig, err := json.MarshalIndent(config, "", "  ")
		if err != nil {
			return err
		}
		err = os.WriteFile("config.json", updatedConfig, 0644)
		if err != nil {
			return err
		}
	}
	engine, err := ReadEngineSettings(config)
	if err != nil {
		log.Printf("解析游戏engine.ini出错,错误%v", err)

	} else {
		config.Engine = engine
		log.Println("从游戏engine.ini解析配置成功.")
		log.Printf("从游戏engine.ini解析配置成功.%v", config.Engine)

		// 将更新后的配置写回文件
		updatedConfig, err := json.MarshalIndent(config, "", "  ")
		if err != nil {
			return err
		}
		err = os.WriteFile("config.json", updatedConfig, 0644)
		if err != nil {
			return err
		}
	}

	return nil
}

func ReadGameWorldSettings(config *Config) (*GameWorldSettings, error) {
	var iniPath string

	// 根据操作系统选择不同的路径
	switch runtime.GOOS {
	case "windows":
		iniPath = filepath.Join(config.GameSavePath, "Config", "WindowsServer", "PalWorldSettings.ini")
	case "linux":
		iniPath = filepath.Join(config.GameSavePath, "Config", "LinuxServer", "PalWorldSettings.ini")
	default:
		// 对于其他操作系统，暂时还不知道，按linux处理
		iniPath = filepath.Join(config.GameSavePath, "Config", "LinuxServer", "PalWorldSettings.ini")
	}

	// 检查INI文件是否存在，如果不存在则创建
	if _, err := os.Stat(iniPath); os.IsNotExist(err) {
		file, err := os.Create(iniPath)
		if err != nil {
			return nil, err
		}
		file.Close()
		fmt.Printf("创建了新的INI文件:%s\n", iniPath)
	}

	// 加载INI文件
	cfg, err := ini.Load(iniPath)
	if err != nil {
		return nil, err
	}
	var settingsString string

	// 获取section
	sectionName := "/Script/Pal.PalGameWorldSettings"
	section, err := cfg.GetSection(sectionName)
	if err != nil {
		fmt.Printf("初次使用，正在为您自动设置游戏默认参数\n")
		settingsString = "(Difficulty=None,DayTimeSpeedRate=1.000000,NightTimeSpeedRate=1.000000,ExpRate=1.000000,PalCaptureRate=1.000000,PalSpawnNumRate=1.000000,PalDamageRateAttack=1.000000,PalDamageRateDefense=1.000000,PlayerDamageRateAttack=1.000000,PlayerDamageRateDefense=1.000000,PlayerStomachDecreaceRate=1.000000,PlayerStaminaDecreaceRate=1.000000,PlayerAutoHPRegeneRate=1.000000,PlayerAutoHpRegeneRateInSleep=1.000000,PalStomachDecreaceRate=1.000000,PalStaminaDecreaceRate=1.000000,PalAutoHPRegeneRate=1.000000,PalAutoHpRegeneRateInSleep=1.000000,BuildObjectDamageRate=1.000000,BuildObjectDeteriorationDamageRate=1.000000,CollectionDropRate=1.000000,CollectionObjectHpRate=1.000000,CollectionObjectRespawnSpeedRate=1.000000,EnemyDropItemRate=1.000000,DeathPenalty=All,bEnablePlayerToPlayerDamage=False,bEnableFriendlyFire=False,bEnableInvaderEnemy=True,bActiveUNKO=False,bEnableAimAssistPad=True,bEnableAimAssistKeyboard=False,DropItemMaxNum=3000,DropItemMaxNum_UNKO=100,BaseCampMaxNum=128,BaseCampWorkerMaxNum=15,DropItemAliveMaxHours=1.000000,bAutoResetGuildNoOnlinePlayers=False,AutoResetGuildTimeNoOnlinePlayers=72.000000,GuildPlayerMaxNum=20,PalEggDefaultHatchingTime=72.000000,WorkSpeedRate=1.000000,bIsMultiplay=False,bIsPvP=False,bCanPickupOtherGuildDeathPenaltyDrop=False,bEnableNonLoginPenalty=True,bEnableFastTravel=True,bIsStartLocationSelectByMap=True,bExistPlayerAfterLogout=False,bEnableDefenseOtherGuildPlayer=False,CoopPlayerMaxNum=4,ServerPlayerMaxNum=32,ServerName=\"palgo\",ServerDescription=\"https://github.com/Hoshinonyaruko/palworld-go\",AdminPassword=\"useradmin\",ServerPassword=\"\",PublicPort=8211,PublicIP=\"\",RCONEnabled=True,RCONPort=25575,Region=\"\",bUseAuth=True,BanListURL=\"https://api.palworldgame.com/api/banlist.txt\")"
		fmt.Printf("已为您生成默认游戏配置，默认控制台地址:http://127.0.0.1:8000\n")
		fmt.Printf("控制台默认用户名(在ServerName配置)\n")
		fmt.Printf("控制台默认密码(在AdminPassword配置)\n")
		fmt.Printf("登录cookie 24小时有效,若在控制台修改后需立即刷新,删除cookie.db并使用新的用户名密码登录\n")
		// 解析设置字符串
		return parseSettings(settingsString), nil
	}

	// 获取OptionSettings项的值
	optionSettingsKey, err := section.GetKey("OptionSettings")
	if err != nil {
		fmt.Printf("未找到配置设置,使用游戏默认配置\n")
		settingsString = "(Difficulty=None,DayTimeSpeedRate=1.000000,NightTimeSpeedRate=1.000000,ExpRate=1.000000,PalCaptureRate=1.000000,PalSpawnNumRate=1.000000,PalDamageRateAttack=1.000000,PalDamageRateDefense=1.000000,PlayerDamageRateAttack=1.000000,PlayerDamageRateDefense=1.000000,PlayerStomachDecreaceRate=1.000000,PlayerStaminaDecreaceRate=1.000000,PlayerAutoHPRegeneRate=1.000000,PlayerAutoHpRegeneRateInSleep=1.000000,PalStomachDecreaceRate=1.000000,PalStaminaDecreaceRate=1.000000,PalAutoHPRegeneRate=1.000000,PalAutoHpRegeneRateInSleep=1.000000,BuildObjectDamageRate=1.000000,BuildObjectDeteriorationDamageRate=1.000000,CollectionDropRate=1.000000,CollectionObjectHpRate=1.000000,CollectionObjectRespawnSpeedRate=1.000000,EnemyDropItemRate=1.000000,DeathPenalty=All,bEnablePlayerToPlayerDamage=False,bEnableFriendlyFire=False,bEnableInvaderEnemy=True,bActiveUNKO=False,bEnableAimAssistPad=True,bEnableAimAssistKeyboard=False,DropItemMaxNum=3000,DropItemMaxNum_UNKO=100,BaseCampMaxNum=128,BaseCampWorkerMaxNum=15,DropItemAliveMaxHours=1.000000,bAutoResetGuildNoOnlinePlayers=False,AutoResetGuildTimeNoOnlinePlayers=72.000000,GuildPlayerMaxNum=20,PalEggDefaultHatchingTime=72.000000,WorkSpeedRate=1.000000,bIsMultiplay=False,bIsPvP=False,bCanPickupOtherGuildDeathPenaltyDrop=False,bEnableNonLoginPenalty=True,bEnableFastTravel=True,bIsStartLocationSelectByMap=True,bExistPlayerAfterLogout=False,bEnableDefenseOtherGuildPlayer=False,CoopPlayerMaxNum=4,ServerPlayerMaxNum=32,ServerName=\"palgo\",ServerDescription=\"https://github.com/Hoshinonyaruko/palworld-go\",AdminPassword=\"useradmin\",ServerPassword=\"\",PublicPort=8211,PublicIP=\"\",RCONEnabled=True,RCONPort=25575,Region=\"\",bUseAuth=True,BanListURL=\"https://api.palworldgame.com/api/banlist.txt\")"
		fmt.Printf("已为您生成默认游戏配置，默认控制台地址:http://127.0.0.1:8000\n")
		fmt.Printf("控制台默认用户名(在ServerName配置):palgo\n")
		fmt.Printf("控制台默认密码(在AdminPassword配置):useradmin\n")
		fmt.Printf("登录cookie 24小时有效,若在控制台修改后需立即刷新,删除cookie.db并使用新的用户名密码登录\n")
	} else {
		settingsString = optionSettingsKey.String()
	}

	// 解析设置字符串
	return parseSettings(settingsString), nil
}

func firstToUpper(s string) string {
	if s == "" {
		return s
	}
	r := []rune(s)
	r[0] = unicode.ToUpper(r[0])
	return string(r)
}

func parseSettings(settingsString string) *GameWorldSettings {
	// Remove the "(" prefix and the closing ")"
	trimmed := strings.TrimPrefix(settingsString, "(")
	trimmed = strings.TrimSuffix(trimmed, ")")

	// Split the settings into key-value pairs
	keyValuePairs := strings.Split(trimmed, ",")

	settings := &GameWorldSettings{}
	sValue := reflect.ValueOf(settings).Elem()
	sType := sValue.Type()

	for _, pair := range keyValuePairs {
		keyValue := strings.SplitN(pair, "=", 2)
		if len(keyValue) != 2 {
			continue
		}

		key := strings.TrimSpace(keyValue[0])
		value := strings.TrimSpace(keyValue[1])
		log.Printf("加载帕鲁ini,key:%v,value:%v", key, value)

		// 直接移除key中可能存在的前缀'b'
		key = strings.TrimPrefix(key, "b")

		// 特殊规则处理
		if key == "RCONEnabled" {
			key = "RconEnabled"
		} else if key == "RCONPort" {
			key = "RconPort"
		}

		for i := 0; i < sType.NumField(); i++ {
			field := sType.Field(i)
			// 将json标签首字母转换为大写
			jsonTag := firstToUpper(strings.Split(field.Tag.Get("json"), ",")[0]) // 获取json标签的第一部分，忽略后面的选项（如omitempty）
			//log.Printf("调试,jsonTag:%v,key:%v", jsonTag, key)
			if jsonTag == key {
				fieldValue := sValue.Field(i)
				if fieldValue.CanSet() {
					switch fieldValue.Kind() {
					case reflect.String:
						trimmedValue := strings.Trim(value, "\"") // 移除双引号
						fieldValue.SetString(trimmedValue)
					case reflect.Float64:
						if val, err := strconv.ParseFloat(value, 64); err == nil {
							fieldValue.SetFloat(val)
						}
					case reflect.Int:
						if val, err := strconv.Atoi(value); err == nil {
							fieldValue.SetInt(int64(val))
						}
					case reflect.Bool:
						if val, err := strconv.ParseBool(value); err == nil {
							fieldValue.SetBool(val)
						}
					}
				}
			}
		}
	}
	return settings
}

func settingsToString(settings *GameWorldSettings) string {
	var settingsParts []string

	sValue := reflect.ValueOf(settings).Elem()
	sType := sValue.Type()

	for i := 0; i < sValue.NumField(); i++ {
		field := sType.Field(i)
		fieldValue := sValue.Field(i)

		jsonTag := firstToUpper(strings.Split(field.Tag.Get("json"), ",")[0]) // 获取json标签的第一部分，并将首字母转换为大写

		// 特殊规则处理
		if jsonTag == "RconEnabled" {
			jsonTag = "RCONEnabled"
		} else if jsonTag == "RconPort" {
			jsonTag = "RCONPort"
		} else if fieldValue.Kind() == reflect.Bool {
			// 如果字段是布尔类型，并且不是RconEnabled，在jsonTag前加上小写的'b'
			jsonTag = "b" + jsonTag
		}

		var valueString string
		switch fieldValue.Kind() {
		case reflect.String:
			valueString = "\"" + fieldValue.String() + "\"" // 添加双引号
		case reflect.Float64:
			valueString = strconv.FormatFloat(fieldValue.Float(), 'f', 6, 64) // 格式化浮点数，保留6位小数
		case reflect.Int:
			valueString = strconv.FormatInt(fieldValue.Int(), 10)
		case reflect.Bool:
			valueString = strconv.FormatBool(fieldValue.Bool())
		}

		settingsPart := fmt.Sprintf("%s=%s", jsonTag, valueString)
		settingsParts = append(settingsParts, settingsPart)
	}

	return "(" + strings.Join(settingsParts, ",") + ")"
}

func WriteGameWorldSettings(config *Config, settings *GameWorldSettings) error {
	var iniPath string

	// 根据操作系统选择不同的路径
	switch runtime.GOOS {
	case "windows":
		iniPath = filepath.Join(config.GameSavePath, "Config", "WindowsServer", "PalWorldSettings.ini")
	case "linux":
		iniPath = filepath.Join(config.GameSavePath, "Config", "LinuxServer", "PalWorldSettings.ini")
	default:
		iniPath = filepath.Join(config.GameSavePath, "Config", "LinuxServer", "PalWorldSettings.ini")
	}

	// 加载INI文件
	cfg, err := ini.Load(iniPath)
	if err != nil {
		return err
	}

	// 获取或创建section
	sectionName := "/Script/Pal.PalGameWorldSettings"
	section, err := cfg.GetSection(sectionName)
	if err != nil {
		if section, err = cfg.NewSection(sectionName); err != nil {
			return err
		}
	}

	// 使用settingsToString函数生成OptionSettings值
	optionSettingsValue := settingsToString(settings)

	// 获取或创建OptionSettings项，并设置其值
	optionSettingsKey, err := section.GetKey("OptionSettings")
	if err != nil {
		if _, err = section.NewKey("OptionSettings", optionSettingsValue); err != nil {
			return err
		}
	} else {
		optionSettingsKey.SetValue(optionSettingsValue)
	}

	// 保存修改后的INI文件
	return cfg.SaveTo(iniPath)
}

// ReadEngineSettings 读取并解析引擎的INI配置文件
func ReadEngineSettings(config *Config) (*Engine, error) {
	var iniPath string

	// 根据操作系统选择不同的路径
	switch runtime.GOOS {
	case "windows":
		iniPath = filepath.Join(config.GameSavePath, "Config", "WindowsServer", "Engine.ini")
	case "linux":
		iniPath = filepath.Join(config.GameSavePath, "Config", "LinuxServer", "Engine.ini")
	default:
		iniPath = filepath.Join(config.GameSavePath, "Config", "LinuxServer", "Engine.ini")
	}

	// 检查INI文件是否存在，如果不存在则创建
	if _, err := os.Stat(iniPath); os.IsNotExist(err) {
		file, err := os.Create(iniPath)
		if err != nil {
			return nil, err
		}
		file.Close()
		fmt.Printf("创建了新的INI文件: %s\n", iniPath)
	}

	// 加载INI文件
	cfg, err := ini.Load(iniPath)
	if err != nil {
		return nil, err
	}

	// 初始化Engine结构体
	engine := new(Engine)

	sectionName := "/script/engine.player"
	_, err = cfg.GetSection(sectionName)
	if err != nil {
		engine := &defaultEngine
		fmt.Printf("为您加载了网路配置优化参数,提升服务器FPS(包速率)\n")
		return engine, nil
	}
	sectionName = "/script/socketsubsystemepic.epicnetdriver"
	_, err = cfg.GetSection(sectionName)
	if err != nil {
		engine := &defaultEngine
		fmt.Printf("为您加载了网路配置优化参数,提升服务器FPS(包速率)\n")
		return engine, nil
	}
	sectionName = "/script/engine.engine"
	_, err = cfg.GetSection(sectionName)
	if err != nil {
		engine := &defaultEngine
		fmt.Printf("为您加载了网路配置优化参数,提升服务器FPS(包速率)\n")
		return engine, nil
	}

	// 解析各个section到对应的结构体中
	if err := cfg.Section("/script/engine.player").MapTo(&engine.Player); err != nil {
		return nil, err
	}
	if err := cfg.Section("/script/socketsubsystemepic.epicnetdriver").MapTo(&engine.SocketSubsystemEpic); err != nil {
		return nil, err
	}
	if err := cfg.Section("/script/engine.engine").MapTo(&engine.EngineConfig); err != nil {
		return nil, err
	}

	// 手动解析SmoothedFrameRateRange字段
	smoothedFrameRateRangeStr := cfg.Section("/script/engine.engine").Key("SmoothedFrameRateRange").String()
	if smoothedFrameRateRangeStr != "" {
		frameRateRange, err := parseFrameRateRange(smoothedFrameRateRangeStr)
		if err != nil {
			return nil, fmt.Errorf("parsing SmoothedFrameRateRange failed: %w", err)
		}
		engine.EngineConfig.SmoothedFrameRateRange = frameRateRange
	}

	// 手动设置bSmoothFrameRate和bUseFixedFrameRate字段
	engine.EngineConfig.BSmoothFrameRate = cfg.Section("/script/engine.engine").Key("bSmoothFrameRate").MustBool(false)
	engine.EngineConfig.BUseFixedFrameRate = cfg.Section("/script/engine.engine").Key("bUseFixedFrameRate").MustBool(false)

	return engine, nil
}

// formatFrameRateRange 格式化FrameRateRange字段
func formatFrameRateRange(frr FrameRateRange) string {
	return fmt.Sprintf("(LowerBound=(Type=%s,Value=%.6f),UpperBound=(Type=%s,Value=%.6f))",
		frr.LowerBound.Type, frr.LowerBound.Value, frr.UpperBound.Type, frr.UpperBound.Value)
}

func parseFrameRateRange(s string) (FrameRateRange, error) {
	var frr FrameRateRange
	var lbType, ubType string
	var lbValue, ubValue float64

	// 使用正则表达式来解析字符串
	pattern := `\((LowerBound=\(Type=(\w+),Value=([0-9.]+)\)),(UpperBound=\(Type=(\w+),Value=([0-9.]+)\))\)`
	re := regexp.MustCompile(pattern)
	matches := re.FindStringSubmatch(s)

	if len(matches) != 7 {
		return frr, fmt.Errorf("invalid format")
	}

	// 提取数据
	lbType, lbValueStr, ubType, ubValueStr := matches[2], matches[3], matches[5], matches[6]

	// 将字符串值转换为float64
	var err error
	lbValue, err = strconv.ParseFloat(lbValueStr, 64)
	if err != nil {
		return frr, err
	}
	ubValue, err = strconv.ParseFloat(ubValueStr, 64)
	if err != nil {
		return frr, err
	}

	// 构造FrameRateRange结构体
	frr = FrameRateRange{
		LowerBound: Bound{Type: lbType, Value: lbValue},
		UpperBound: Bound{Type: ubType, Value: ubValue},
	}

	return frr, nil
}

// WriteEngineSettings 将Engine结构体的数据写入INI文件
func WriteEngineSettings(config *Config, engine *Engine) error {
	var iniPath string

	// 根据操作系统选择不同的路径
	switch runtime.GOOS {
	case "windows":
		iniPath = filepath.Join(config.GameSavePath, "Config", "WindowsServer", "Engine.ini")
	case "linux":
		iniPath = filepath.Join(config.GameSavePath, "Config", "LinuxServer", "Engine.ini")
	default:
		iniPath = filepath.Join(config.GameSavePath, "Config", "LinuxServer", "Engine.ini")
	}

	// 读取INI文件的所有内容
	fileContent, err := ioutil.ReadFile(iniPath)
	if err != nil {
		return err
	}

	// 将文件内容转换为字符串
	content := string(fileContent)

	// 更新content字符串中的特定section
	content = updateIniSection(content, "/script/engine.player", &engine.Player)
	content = updateIniSection(content, "/script/socketsubsystemepic.epicnetdriver", &engine.SocketSubsystemEpic)
	content = updateIniSection(content, "/script/engine.engine", &engine.EngineConfig)

	// 将更新后的内容写回文件
	return os.WriteFile(iniPath, []byte(content), 0644)
}

// updateIniSection 更新或添加INI文件中的特定section
func updateIniSection(content, sectionName string, data interface{}) string {
	// 将结构体转换为键值对映射
	kvMap := structToMap(data)

	var updatedLines []string
	inSection := false
	sectionExists := false
	lines := strings.Split(content, "\n")

	for _, line := range lines {
		// 检查是否进入或离开目标section
		if strings.HasPrefix(line, "["+sectionName+"]") {
			inSection = true
			sectionExists = true
			updatedLines = append(updatedLines, line)
			continue
		} else if strings.HasPrefix(line, "[") && inSection {
			// 已离开目标section
			inSection = false
		}

		if inSection {
			// 在目标section内部，处理键值对
			keyValue := strings.SplitN(line, "=", 2)
			if len(keyValue) == 2 {
				key := keyValue[0]
				if _, exists := kvMap[key]; exists {
					// 替换现有键值对
					updatedLines = append(updatedLines, key+"="+fmt.Sprintf("%v", kvMap[key]))
					delete(kvMap, key) // 从映射中移除该键
					continue
				}
			}
		}

		// 不是目标section或者键不存在，则保留原始行
		updatedLines = append(updatedLines, line)
	}

	// 如果section不存在，添加新section及其键值对
	if !sectionExists {
		updatedLines = append(updatedLines, "["+sectionName+"]")
		for key, value := range kvMap {
			updatedLines = append(updatedLines, key+"="+fmt.Sprintf("%v", value))
		}
	}

	return strings.Join(updatedLines, "\n")
}

// 解释一下，为什么要从头实现ini解析，因为游戏engine配置中的节对应了多个，重复名称的paths项，所以无法通过go的iniv1包解析，否则这些重复项会归一
func structToMap(data interface{}) map[string]string {
	kvMap := make(map[string]string)
	v := reflect.ValueOf(data)

	if v.Kind() == reflect.Ptr && v.Elem().Kind() == reflect.Struct {
		v = v.Elem()
	} else {
		log.Println("Data is not a pointer to a struct")
		return kvMap
	}

	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldType := field.Type()

		// 特殊处理FrameRateRange
		if t.Field(i).Name == "SmoothedFrameRateRange" {
			if fieldType.Kind() == reflect.Struct {
				fr, ok := field.Interface().(FrameRateRange)
				if ok {
					kvMap["SmoothedFrameRateRange"] = formatFrameRateRange(fr)
				} else {
					log.Printf("Field %s is not of type FrameRateRange\n", t.Field(i).Name)
				}
			}
			continue
		}

		key := t.Field(i).Tag.Get("json")
		if key == "" {
			key = t.Field(i).Name
		}

		var value string
		switch field.Kind() {
		case reflect.String:
			value = field.String()
		case reflect.Int, reflect.Int64:
			value = fmt.Sprintf("%d", field.Int())
		case reflect.Float64:
			value = fmt.Sprintf("%.6f", field.Float())
		case reflect.Bool:
			value = fmt.Sprintf("%t", field.Bool())
		default:
			log.Printf("Unsupported type for field %s\n", t.Field(i).Name)
			continue
		}

		kvMap[key] = value
		//log.Printf("Processed field: %s, Value: %s\n", key, value)
	}

	return kvMap
}
