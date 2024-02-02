package bot

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hoshinonyaruko/palworld-go/config"
)

// 群信息事件
type OnebotGroupMessage struct {
	RawMessage      string      `json:"raw_message"`
	MessageID       int         `json:"message_id"`
	GroupID         int64       `json:"group_id"` // Can be either string or int depending on p.Settings.CompleteFields
	MessageType     string      `json:"message_type"`
	PostType        string      `json:"post_type"`
	SelfID          int64       `json:"self_id"` // Can be either string or int
	Sender          Sender      `json:"sender"`
	SubType         string      `json:"sub_type"`
	Time            int64       `json:"time"`
	Avatar          string      `json:"avatar,omitempty"`
	Echo            string      `json:"echo,omitempty"`
	Message         interface{} `json:"message"` // For array format
	MessageSeq      int         `json:"message_seq"`
	Font            int         `json:"font"`
	UserID          int64       `json:"user_id"`
	RealMessageType string      `json:"real_message_type,omitempty"`  //当前信息的真实类型 group group_private guild guild_private
	IsBindedGroupId bool        `json:"is_binded_group_id,omitempty"` //当前群号是否是binded后的
	IsBindedUserId  bool        `json:"is_binded_user_id,omitempty"`  //当前用户号号是否是binded后的
}

type Sender struct {
	Nickname string `json:"nickname"`
	TinyID   string `json:"tiny_id"`
	UserID   int64  `json:"user_id"`
	Role     string `json:"role,omitempty"`
	Card     string `json:"card,omitempty"`
	Sex      string `json:"sex,omitempty"`
	Age      int32  `json:"age,omitempty"`
	Area     string `json:"area,omitempty"`
	Level    string `json:"level,omitempty"`
	Title    string `json:"title,omitempty"`
}

type KickOrBanRequest struct {
	PlayerUID string `json:"playeruid"`
	SteamID   string `json:"steamid"`
	Type      string `json:"type"`
}

// BroadcastRequest 用于封装广播请求的结构体
type BroadcastRequest struct {
	Message string `json:"message"`
}

// RestartLaterRequest 用于绑定JSON请求体
type RestartLaterRequest struct {
	Seconds string `json:"seconds"`
	Message string `json:"message"`
}

// GensokyoHandlerClosure 创建一个中间件闭包
func GensokyoHandlerClosure(c *gin.Context, config config.Config) {

	if c.Request.Method != http.MethodPost {
		c.String(http.StatusMethodNotAllowed, "Only POST method is allowed")
		return
	}

	// 读取请求体
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.String(http.StatusInternalServerError, "Error reading request body: %v", err)
		return
	}
	defer c.Request.Body.Close()

	// 解析请求体到OnebotGroupMessage结构体
	var message OnebotGroupMessage
	err = json.Unmarshal(body, &message)
	if err != nil {
		c.String(http.StatusInternalServerError, "Error parsing request body: %v", err)
		return
	}

	// 打印消息和其他相关信息
	fmt.Printf("Received message: %v\n", message.Message)
	fmt.Printf("Full message details: %+v\n", message)

	// 判断message.Message的类型
	switch msg := message.Message.(type) {
	case string:
		// 如果消息是字符串类型
		fmt.Printf("Received string message: %s\n", msg)
		// 去除字符串前后的空格
		msg = strings.TrimSpace(msg)

		// 处理以 "getbot" 开头的消息
		if strings.HasPrefix(msg, "getbot") {
			getBotHandler(msg, message, config)
			return
		}

		// 处理以 "player" 开头的消息
		if strings.HasPrefix(msg, "player") {
			getplayerHandler(msg, message, config, false)
			return
		}

		// 处理以 "update player" 开头的消息
		if strings.HasPrefix(msg, "update player") {
			getplayerHandler(msg, message, config, true)
			return
		}

		// 处理以 "玩家列表" 开头的消息
		if strings.HasPrefix(msg, "玩家列表") {
			getplayerHandler(msg, message, config, false)
			return
		}

		// 处理以 "刷新玩家列表" 开头的消息
		if strings.HasPrefix(msg, "刷新玩家列表") {
			getplayerHandler(msg, message, config, true)
			return
		}

		// 处理以 "kick" 开头的消息
		if strings.HasPrefix(msg, "kick") {
			kickorbanHandler(msg, message, config, "kick")
			return
		}

		// 处理以 "踢人" 开头的消息
		if strings.HasPrefix(msg, "踢人") {
			kickorbanHandler(msg, message, config, "kick")
			return
		}

		// 处理以 "ban" 开头的消息
		if strings.HasPrefix(msg, "ban") {
			kickorbanHandler(msg, message, config, "ban")
			return
		}

		// 处理以 "封禁" 开头的消息
		if strings.HasPrefix(msg, "封禁") {
			kickorbanHandler(msg, message, config, "ban")
			return
		}

		// 处理以 "Broadcast" 开头的消息
		if strings.HasPrefix(msg, "Broadcast") {
			broadcastMessageHandler(msg, message, config)
			return
		}

		// 处理以 "广播" 开头的消息
		if strings.HasPrefix(msg, "广播") {
			broadcastMessageHandler(msg, message, config)
			return
		}

		// 处理以 "重启服务器" 开头的消息
		if strings.HasPrefix(msg, "重启服务器") {
			restartHandler(msg, message, config)
			return
		}

		// 处理以 "restart" 开头的消息
		if strings.HasPrefix(msg, "restart") {
			restartHandler(msg, message, config)
			return
		}

		// 处理以 "指令列表" 开头的消息
		if strings.HasPrefix(msg, "指令列表") {
			listCommandsHandler(message, config)
			return
		}

		// 处理以 "commonlist" 开头的消息
		if strings.HasPrefix(msg, "commonlist") {
			listCommandsHandler(message, config)
			return
		}

		// 其他消息处理
		// ...

	default:
		// 其他类型的消息处理
		// ...
	}

	// ... 其他响应逻辑
	c.String(http.StatusOK, "ok")
}

func getBotHandler(msg string, message OnebotGroupMessage, config config.Config) {
	// 检查消息是否至少以一个 "getbot" 开头
	if !strings.HasPrefix(msg, "getbot") {
		sendGroupMessage(message.GroupID, message.UserID, "错误,指令需要以getbot开头", config)
		return
	}

	// 将消息分割成多个部分
	parts := strings.Fields(msg)

	// 查找第一个不是 "getbot" 的部分
	var startIndex int
	for i, part := range parts {
		if part != "getbot" {
			startIndex = i
			break
		}
	}

	// 确保在第一个非 "getbot" 部分之后还有足够的参数
	if len(parts) < startIndex+3 {
		sendGroupMessage(message.GroupID, message.UserID, "指令错误,请在palworld-go项目的机器人管理面板生成指令", config)
		return
	}

	// 解析number
	number, err := strconv.ParseInt(parts[startIndex], 10, 64)
	if err != nil {
		sendGroupMessage(message.GroupID, message.UserID, "错误,请在palworld-go项目的机器人管理面板生成指令,number参数错误", config)
		return
	}

	// 获取uuid和httpsFlag
	uuid := parts[startIndex+1]
	httpsFlag := parts[startIndex+2] // 这应该是 "0" 或 "1"

	// 将httpsFlag从 "0"/"1" 转换回 bool
	useHttps := httpsFlag == "1"

	// 处理cookie和相关逻辑
	exists, _ := CheckAndWriteCookie(uuid)
	if !exists {
		ipWithPort := numberToIPWithPort(number)
		err := StoreUserIDAndIP(message.UserID, ipWithPort, uuid, useHttps)
		if err != nil {
			fmt.Printf("储存pal-go面板端user配置出错 userid: %v 地址:%v\n", message.UserID, ipWithPort)
		} else {
			sendGroupMessage(message.GroupID, message.UserID, "绑定成功,现在你可以在帕鲁帕鲁机器人管理你的palworld-go面板", config)
		}
	} else {
		sendGroupMessage(message.GroupID, message.UserID, "指令无效,请重新生成,为了面板安全,palworld-go指令不可重复使用,如需多人使用,可多次生成.", config)
	}
}

func getplayerHandler(msg string, message OnebotGroupMessage, config config.Config, update bool) {
	// 尝试获取用户的IP和UUID
	userIPData, err := RetrieveIPByUserID(message.UserID)
	if err != nil {
		// 发送错误消息
		sendGroupMessage(message.GroupID, message.UserID, "没有初始化,请使用palworld-go面板,在机器人管理或服务器主人处获取指令,然后发给机器人", config)
		return
	}

	// 检查IP是否为空
	if userIPData.IP != "" {
		// 根据https值确定使用HTTP还是HTTPS
		baseURL := "http://" + userIPData.IP
		if userIPData.Https {
			baseURL = "https://" + userIPData.IP
		}

		// 创建HTTP客户端并设置cookie
		client := &http.Client{}
		// 创建请求URL
		var reqURL string
		if update {
			// 如果update为true，则在请求URL中添加update=true
			reqURL = baseURL + "/api/player?update=true"
		} else {
			// 如果update为false，则不添加
			reqURL = baseURL + "/api/player"
		}

		// 创建HTTP请求
		req, err := http.NewRequest("GET", reqURL, nil)
		if err != nil {
			// 处理错误
			return
		}
		req.AddCookie(&http.Cookie{Name: "login_cookie", Value: userIPData.UUID})

		// 发送请求
		resp, err := client.Do(req)
		if err != nil {
			// 处理错误
			return
		}
		defer resp.Body.Close()
		sendGroupMessage(message.GroupID, message.UserID, "正在刷新玩家,可能需要3-5秒返回,请勿重复操作", config)
		// 读取响应
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			// 处理错误
			return
		}

		// 解析响应
		var players []PlayerInfo
		err = json.Unmarshal(body, &players)
		if err != nil {
			// 处理错误
			return
		}

		// 处理并发送玩家信息
		var responseMessage string
		for _, player := range players {
			// 解析玩家的最后在线时间
			lastOnlineTime, err := time.Parse("2006-01-02 15:04:05", player.LastOnline)
			if err != nil {
				// 处理解析错误
				continue
			}
			formattedLastOnline := formatTimeDifference(lastOnlineTime)

			uniqueID, _ := StorePlayerInfo(player.PlayerUID, player.SteamID, player.Name)
			responseMessage += fmt.Sprintf("[%d] %s 上次在线:%s 在线:%t\n", uniqueID, player.Name, formattedLastOnline, player.Online)
		}
		sendGroupMessage(message.GroupID, message.UserID, responseMessage, config)
	}
}
func formatTimeDifference(t time.Time) string {
	now := time.Now()
	diff := now.Sub(t)

	days := int(diff.Hours() / 24)
	hours := int(diff.Hours()) % 24
	minutes := int(diff.Minutes()) % 60

	if days > 0 {
		return fmt.Sprintf("%d天前", days)
	}
	if hours > 0 {
		return fmt.Sprintf("%d小时前", hours)
	}
	return fmt.Sprintf("%d分钟前", minutes)
}

func sendGroupMessage(groupID int64, userID int64, message string, config config.Config) error {
	// 获取基础URL
	baseURL := config.Onebotv11HttpApiPath

	// 构建完整的URL
	url := baseURL + "/send_group_msg"

	// 构造请求体
	requestBody, err := json.Marshal(map[string]interface{}{
		"group_id": groupID,
		"message":  message,
		"user_id":  userID,
	})
	if err != nil {
		return fmt.Errorf("failed to marshal request body: %w", err)
	}

	// 发送POST请求
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		return fmt.Errorf("failed to send POST request: %w", err)
	}
	defer resp.Body.Close()

	// 检查响应状态
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("received non-OK response status: %s", resp.Status)
	}

	return nil
}

func kickorbanHandler(msg string, message OnebotGroupMessage, config config.Config, operation string) {
	// 使用strings.Fields按空格分割字符串
	parts := strings.Fields(msg)

	// 检查是否至少有两部分（例如："踢人 123"）
	if len(parts) < 2 {
		sendGroupMessage(message.GroupID, message.UserID, "指令格式错误 应为 踢人 1 封禁 1 kick 1 ban 1", config)
		return
	}

	// 解析数字部分
	var uniqueID int64
	_, err := fmt.Sscanf(parts[1], "%d", &uniqueID)
	if err != nil {
		sendGroupMessage(message.GroupID, message.UserID, "指令格式错误 后方应为数字 空格为分割", config)
		return
	}

	//测试提审核代码 不要删除
	if uniqueID == 666 {
		sendGroupMessage(message.GroupID, message.UserID, operation+"测试玩家 成功", config)
		return
	}

	// 通过uniqueID获取玩家信息
	playerInfo, err := RetrievePlayerInfoByID(uniqueID)
	if err != nil {
		sendGroupMessage(message.GroupID, message.UserID, "获取玩家信息失败: "+err.Error(), config)
		return
	}

	// 检查SteamID是否有效
	_, err = strconv.ParseInt(playerInfo.SteamID, 10, 64)
	if err != nil {
		sendGroupMessage(message.GroupID, message.UserID, playerInfo.Name+"无效的SteamID,帕鲁服务端通病,玩家增加后再次使用 玩家列表 获取可解决", config)
		return
	}

	// 构建请求体
	reqBody, err := json.Marshal(KickOrBanRequest{
		PlayerUID: playerInfo.PlayerUID,
		SteamID:   playerInfo.SteamID,
		Type:      operation,
	})
	if err != nil {
		sendGroupMessage(message.GroupID, message.UserID, "构建请求失败: "+err.Error(), config)
		return
	}

	// 尝试获取用户的IP和UUID
	userIPData, err := RetrieveIPByUserID(message.UserID)
	if err != nil {
		// 发送错误消息
		sendGroupMessage(message.GroupID, message.UserID, "没有正确设置,请使用palworld-go面板,在机器人管理或服务器主人处获取指令,然后发给我", config)
		return
	}

	// 检查IP是否为空
	if userIPData.IP != "" {
		// 根据https值确定使用HTTP还是HTTPS
		baseURL := "http://" + userIPData.IP
		if userIPData.Https {
			baseURL = "https://" + userIPData.IP
		}

		// 创建HTTP客户端并设置cookie
		client := &http.Client{}
		apiURL := baseURL + "/api/kickorban"
		req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(reqBody))
		if err != nil {
			sendGroupMessage(message.GroupID, message.UserID, "创建请求失败: "+err.Error(), config)
			return
		}
		req.Header.Set("Content-Type", "application/json")
		req.AddCookie(&http.Cookie{Name: "login_cookie", Value: userIPData.UUID})

		resp, err := client.Do(req)
		if err != nil {
			sendGroupMessage(message.GroupID, message.UserID, "发送请求失败: "+err.Error(), config)
			return
		}
		defer resp.Body.Close()

		// 检查响应状态
		if resp.StatusCode != http.StatusOK {
			sendGroupMessage(message.GroupID, message.UserID, fmt.Sprintf("%s %s 失败", operation, playerInfo.Name), config)
			return
		}

		// 发送成功消息
		sendGroupMessage(message.GroupID, message.UserID, fmt.Sprintf("%s %s 成功", operation, playerInfo.Name), config)
	} else {
		// 发送错误消息
		sendGroupMessage(message.GroupID, message.UserID, "没有获取到面板信息,请使用palworld-go面板,在机器人管理或服务器主人处获取指令,然后发给我", config)
		return
	}
}

func broadcastMessageHandler(msg string, message OnebotGroupMessage, config config.Config) {
	// 从msg中提取广播内容
	parts := strings.SplitN(msg, " ", 2)
	if len(parts) != 2 {
		sendGroupMessage(message.GroupID, message.UserID, "广播指令格式错误", config)
		return
	}

	// 组装BroadcastRequest
	broadcastReq := BroadcastRequest{
		Message: parts[1],
	}
	reqBody, err := json.Marshal(broadcastReq)
	if err != nil {
		sendGroupMessage(message.GroupID, message.UserID, "创建广播请求失败", config)
		return
	}

	// 尝试获取用户的IP和UUID
	userIPData, err := RetrieveIPByUserID(message.UserID)
	if err != nil {
		// 发送错误消息
		sendGroupMessage(message.GroupID, message.UserID, "没有正确设置,请使用palworld-go面板,在机器人管理或服务器主人处获取指令,然后发给我", config)
		return
	}

	// 检查IP是否为空
	if userIPData.IP != "" {
		// 根据https值确定使用HTTP还是HTTPS
		baseURL := "http://" + userIPData.IP
		if userIPData.Https {
			baseURL = "https://" + userIPData.IP
		}

		// 创建HTTP客户端并设置cookie
		apiURL := baseURL + "/api/broadcast"
		req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(reqBody))
		if err != nil {
			sendGroupMessage(message.GroupID, message.UserID, "创建请求失败", config)
			return
		}
		req.Header.Set("Content-Type", "application/json")
		req.AddCookie(&http.Cookie{Name: "login_cookie", Value: userIPData.UUID})

		// 执行请求
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			sendGroupMessage(message.GroupID, message.UserID, "发送广播请求失败", config)
			return
		}
		defer resp.Body.Close()

		// 检查响应状态
		if resp.StatusCode != http.StatusOK {
			sendGroupMessage(message.GroupID, message.UserID, fmt.Sprintf("广播失败，响应状态码: %d", resp.StatusCode), config)
			return
		}

		sendGroupMessage(message.GroupID, message.UserID, "广播消息已成功发送", config)
	} else {
		// 发送错误消息
		sendGroupMessage(message.GroupID, message.UserID, "没有获取到面板信息,请使用palworld-go面板,在机器人管理或服务器主人处获取指令,然后发给我", config)
		return
	}
}

// restartHandler 处理重启服务器的消息
func restartHandler(msg string, message OnebotGroupMessage, config config.Config) {
	// 从msg中提取参数
	parts := strings.Fields(msg)
	if len(parts) < 3 {
		sendGroupMessage(message.GroupID, message.UserID, "重启指令格式错误,应为 重启服务器 多少秒数后重启(整数) 重启公告内容", config)
		return
	}

	// 检查时间参数是否为数字
	seconds, err := strconv.Atoi(parts[1])
	if err != nil {
		sendGroupMessage(message.GroupID, message.UserID, "重启时间应为数字", config)
		return
	}

	// 组装剩余部分为维护公告
	announcement := strings.Join(parts[2:], " ")
	// 尝试获取用户的IP和UUID
	userIPData, err := RetrieveIPByUserID(message.UserID)
	if err != nil {
		// 发送错误消息
		sendGroupMessage(message.GroupID, message.UserID, "没有正确设置,请使用palworld-go面板,在机器人管理或服务器主人处获取指令,然后发给我", config)
		return
	}

	// 检查IP是否为空
	if userIPData.IP != "" {
		// 根据https值确定使用HTTP还是HTTPS
		baseURL := "http://" + userIPData.IP
		if userIPData.Https {
			baseURL = "https://" + userIPData.IP
		}

		// 创建HTTP客户端并设置cookie
		apiURL := baseURL + "/api/restartlater"

		// 构造请求体
		restartReq := RestartLaterRequest{
			Seconds: strconv.Itoa(seconds),
			Message: announcement,
		}

		reqBody, err := json.Marshal(restartReq)
		if err != nil {
			sendGroupMessage(message.GroupID, message.UserID, "创建重启请求失败", config)
			return
		}

		req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(reqBody))
		if err != nil {
			sendGroupMessage(message.GroupID, message.UserID, "创建请求失败", config)
			return
		}

		req.Header.Set("Content-Type", "application/json")
		req.AddCookie(&http.Cookie{Name: "login_cookie", Value: userIPData.UUID})

		// 执行请求
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			sendGroupMessage(message.GroupID, message.UserID, "发送服务器延时重启请求失败", config)
			return
		}
		defer resp.Body.Close()

		// 检查响应状态
		if resp.StatusCode != http.StatusOK {
			sendGroupMessage(message.GroupID, message.UserID, fmt.Sprintf("服务器重启失败，响应状态码: %d", resp.StatusCode), config)
			return
		}

		// 发送成功消息
		sendGroupMessage(message.GroupID, message.UserID, "服务器将在"+strconv.Itoa(seconds)+"秒后重启，维护公告: "+announcement, config)
	} else {
		// 发送错误消息
		sendGroupMessage(message.GroupID, message.UserID, "没有获取到面板信息,请使用palworld-go面板,在机器人管理或服务器主人处获取指令,然后发给我", config)
		return
	}

}

func listCommandsHandler(message OnebotGroupMessage, config config.Config) {
	// 构建指令列表
	commands := []string{
		"getbot - 获取机器人信息",
		"player - 获取玩家信息",
		"update player - 更新玩家信息",
		"玩家列表 - 显示玩家列表",
		"刷新玩家列表 - 刷新并显示玩家列表",
		"kick - 踢出玩家",
		"踢人 - 踢出玩家",
		"ban - 封禁玩家",
		"封禁 - 封禁玩家",
		"Broadcast - 发送广播消息",
		"广播 - 发送广播消息",
		"重启服务器 - 重启游戏服务器",
		"restart - 重启游戏服务器",
	}

	// 将指令列表转换为字符串，每个指令后换行
	commandsStr := strings.Join(commands, "\n")

	// 发送指令列表
	sendGroupMessage(message.GroupID, message.UserID, "可用指令列表:\n"+commandsStr, config)
}
