<p align="center">
  <a href="https://www.github.com/hoshinonyaruko/palworld-go">
    <img src="pic/1.gif" width="200" height="200" alt="palworld-go">
  </a>
</p>

<div align="center">

# palworld-go

_✨ 适用于palworld的跨平台服务端面板 ✨_  

_✨ 使用go+quasar实现的palworld webui ✨_  

## 特别鸣谢+推荐

本项目的直接参考（linux版的palworld服务端守护脚本）

https://gist.github.com/Bluefissure/b0fcb05c024ee60cad4e23eb55463062
本项目内置了该项目的编译后网页dist
https://github.com/Bluefissure/pal-conf

## 使用方法

本项目的webui特别对移动端设备进行优化，手机使用更轻松
（老版本iossafari 如果遇到按钮点不动刷新页面再点即可）

webui可友善的可视化的修改帕鲁服务器，守护配置，内存配置，目前webui端口固定8000

将可执行文件放置在

\steamcmd\steamapps\common\PalServer\PalServer.exe

同级目录

运行palworld-go.exe 会自动进入webui

webui默认地址:http://127.0.0.1:8000

端口可在config.json修改，放通至公网可在公网访问

控制台默认用户名 palgo 默认密码 useradmin

用户名即帕鲁服务器名（serverName），可中文 密码即rcon密码（adminPassword） 纯英文

图片介绍

![内存清理和定时广播等设定](pic/1.png)

![帕鲁服务器设定](pic/2.png)

![直接按钮开关](pic/3.png)

![可自动补全的rcon命令](pic/4.png)

![bluefissure制作的sav修改页面](pic/5.png)

![服务器监控](pic/6.png)

## 兼容性
windows通过了测试，linux有待测试

## 场景支持

在手机上痛快的操作和管理服务器，当管理不再手忙脚乱。

内存不足的时候，通过rcon通知服务器成员，然后重启服务器

通过调用微软的rammap释放无用内存，并将有用内存转移至虚拟内存，实现一次释放50%+内存
