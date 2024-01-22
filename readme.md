<p align="center">
  <a href="https://www.github.com/hoshinonyaruko/palworld-go">
    <img src="pic/1.gif" width="200" height="200" alt="palworld-go">
  </a>
</p>

<div align="center">

# palworld-go

_✨ 适用于palworld的进程守护+内存不足自动重启服务端 ✨_  

## 使用方法
启动后配置（会继续完善）

\steamcmd\steamapps\common\PalServer\Pal\Saved\Config\WindowsServer\PalWorldSettings.ini

将

\steamcmd\steamapps\common\PalServer\DefaultPalWorldSettings.ini

中内容复制放入，然后打开rcon，设置DefaultPalWorldSettings中的AdminPassword

{
    "gamePath": "C:\\Users\\Administrator\\Downloads\\steamcmd\\steamapps\\common\\PalServer",

    "gameSavePath": "C:\\Users\\Administrator\\Downloads\\steamcmd\\steamapps\\common\\PalServer\\Pal\\Saved",

    "backupPath": "C:\\Users\\Administrator\\Desktop\\save",

    "address": "127.0.0.1:25575",

    "rconPort": "25575",

    "adminPassword": "",

    "processName": "PalServer",

    "checkInterval": 30,

    "serviceLogFile": "/service.log",

    "serviceErrorFile": "/service.err",

    "backupInterval": 1800,

    "memoryCheckInterval": 30,

    "memoryUsageThreshold": 30,

    "totalMemoryGB": 16
}

配置说明

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


## 兼容性
windows通过了测试，linux有待测试

## 场景支持

内存不足的时候，通过rcon通知服务器成员，然后重启服务器