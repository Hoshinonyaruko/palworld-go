package main

import (
	"log"
	"math/rand"
	"strconv"
	"time"

	"github.com/hoshinonyaruko/palworld-go/config"
)

type palworldBroadcast struct {
	Config     config.Config
	Ticker     *time.Ticker
	BackupTask *BackupTask
}

func NewpalworldBroadcast(config config.Config) *palworldBroadcast {
	return &palworldBroadcast{
		Config: config,
		Ticker: time.NewTicker(time.Duration(config.MessageBroadcastInterval) * time.Second),
	}
}

func (task *palworldBroadcast) Schedule() {
	for range task.Ticker.C {
		task.RunpalworldBroadcast()
	}
}

func (task *palworldBroadcast) RunpalworldBroadcast() {
	log.Println("准备进行全服推送...现已支持所有语言broadcast!")
	// 初始化RCON客户端
	address := task.Config.Address + ":" + strconv.Itoa(task.Config.WorldSettings.RconPort)
	rconClient := NewRconClient(address, task.Config.WorldSettings.AdminPassword, task.BackupTask, &task.Config)
	if rconClient == nil {
		log.Println("RCON客户端初始化失败,无法进行定期推送,请按教程正确开启rcon和设置服务端admin密码")
		return
	}
	// RegularMessages是RegularMessages切片
	if len(task.Config.RegularMessages) > 0 {
		// 随机生成一个索引来选择消息
		randomIndex := rand.Intn(len(task.Config.RegularMessages))

		// 获取随机选择的消息
		randomMessage := task.Config.RegularMessages[randomIndex]
		Broadcast(randomMessage, rconClient, task.Config.UseDll)
	}
}
