<p align="center">
  <a href="https://www.github.com/hoshinonyaruko/palworld-go">
    <img src="pic/1.gif" width="200" height="200" alt="palworld-go">
  </a>
</p>

<div align="center">

# palworld-go

_✨ 适用于palworld的进程守护+强力内存释放+内存不足自动重启服务端 ✨_  

_✨ 使用go+quasar实现的palworld webui ✨_  

## 特别鸣谢+推荐

本项目的直接参考（linux版的palworld服务端守护脚本）

https://gist.github.com/Bluefissure/b0fcb05c024ee60cad4e23eb55463062
本项目内置了该项目的编译后网页dist
https://github.com/Bluefissure/pal-conf

## 使用方法

webui可友善的可视化的修改帕鲁服务器，守护配置，内存配置，目前webui端口固定52000

本项目可以接管游戏服务端配置，可以以json的格式配置游戏服务端

在本程序运行时，直接修改config.json会将改动自动同步到PalWorldSettings.ini

启动后配置（会继续完善）

打开\steamcmd\steamapps\common\PalServer\DefaultPalWorldSettings.ini配置文件

修改RCONEnabled=False，把False改为True启用Rcon，修改AdminPassword=""在""中设置你的管理员密码

修改完成后保存配置文件，复制文档全部内容到

\steamcmd\steamapps\common\PalServer\Pal\Saved\Config\WindowsServer\PalWorldSettings.ini

保存配置文件

第一次启动palworld-go-windows-amd64.exe后会生成Config.JSON配置文件

{

	在""中填入你的服务端安装路径将\改为\\如下方举例

    "gamePath": "C:\\steamcmd\\steamapps\\common\\PalServer",

	在""中填入你的服务端存档路径将\改为\\如下方举例

    "gameSavePath": "C:\\steamcmd\\steamapps\\common\\PalServer\\Pal\\Saved",

	在""中填入你的备份存档保存路径将\改为\\如下方举例

    "backupPath": "C:\\Users\\Administrator\\Desktop\\BackUp",

	在""中填入127.0.0.1:25575固定值

    "address": "127.0.0.1:25575",

	在""中填入25575固定值

    "rconPort": "25575",

	在""中填入开始在DefaultPalWorldSettings.ini中设置的管理员密码

    "adminPassword": "",

	进程名称 PalServer，默认不要修改

    "processName": "PalServer",

	 进程存活检查时间，单位为秒，下方数值为30秒

    "checkInterval": 30,

	 日志文件路径

    "serviceLogFile": "/service.log",

	错误日志文件路径
    "serviceErrorFile": "/service.err",

	存档备份时间，以秒为单位，下方数值为3600秒

    "backupInterval": 3600,

	内存占用检测间隔时间，单位为秒，下方数值为30秒

    "memoryCheckInterval": 30,

	内存重启阈值，单位为百分比，下方数值为80%

    "memoryUsageThreshold": 80,

	你服务器最大内存，单位为G，下方数值为32G

    "totalMemoryGB": 32,

	服务器广播,单位为秒，下方数值为1800秒

    "memoryCleanupInterval": 1800,

	服务器广播内容，在""中填入，你要广播的内容

    "regularMessages": [

        "",

        ""

    ],

	内存自动清理间隔，单位为秒，下方数值为3600秒，第一次使用要看效果可设置10秒，然后再加大间隔。

    "messageBroadcastInterval": 3600,

	服务器重启广播在""中填入，你要广播的内容

    "maintenanceWarningMessage": "The server is about to be maintained. Your archive has been saved. Please log in again 
    
    in 1 minute."

}


## 兼容性
windows通过了测试，linux有待测试

## 场景支持

内存不足的时候，通过rcon通知服务器成员，然后重启服务器
通过调用微软的rammap释放无用内存，并将有用内存转移至虚拟内存，实现一次释放50%+内存