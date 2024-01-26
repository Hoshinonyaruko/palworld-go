package main

import (
	"log"
	"math/rand"
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
	log.Println("准备进行全服推送...由于帕鲁暂未支持中文，仅支持英文")
	// 初始化RCON客户端
	rconClient := NewRconClient(task.Config.Address, task.Config.AdminPassword, task.BackupTask)
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

		// 如果只有一个消息，则使用它
		if len(task.Config.RegularMessages) == 1 {
			Broadcast(randomMessage, rconClient)
		} else {
			// 使用随机选择的消息作为参数调用Broadcast
			Broadcast(randomMessage, rconClient)
		}
	}
}
