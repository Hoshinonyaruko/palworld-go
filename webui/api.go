package webui

import (
	"embed"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorcon/rcon"
	"github.com/gorilla/websocket"
	"github.com/hoshinonyaruko/palworld-go/config"
	"github.com/hoshinonyaruko/palworld-go/sys"
	"github.com/hoshinonyaruko/palworld-go/tool"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/host"
	"github.com/shirou/gopsutil/mem"
	"github.com/shirou/gopsutil/process"
	"go.etcd.io/bbolt"
)

type Player struct {
	Name       string    `json:"name"`
	SteamID    string    `json:"steamid"`
	PlayerUID  string    `json:"playeruid"`
	LastOnline time.Time `json:"last_online"`
}

type Client struct {
	conn *websocket.Conn
	send chan string
}

// RconClient 结构体，用于存储RCON连接和配置信息
type RconClient struct {
	Conn *rcon.Conn
}

//go:embed dist/*
//go:embed dist/icons/*
//go:embed dist/assets/*
var content embed.FS

//go:embed dist2/*
//go:embed dist2/assets/*
var content2 embed.FS

func InitDB() *bbolt.DB {
	db, err := bbolt.Open("players.db", 0600, &bbolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		log.Fatal(err)
	}

	// 创建bucket
	err = db.Update(func(tx *bbolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("players"))
		return err
	})

	if err != nil {
		log.Fatal(err)
	}

	return db
}

// NewCombinedMiddleware 创建并返回一个带有依赖的中间件闭包
func CombinedMiddleware(config config.Config, db *bbolt.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		if strings.HasPrefix(c.Request.URL.Path, "/api") {

			if c.Param("filepath") == "/api/ws" {
				if c.GetHeader("Upgrade") == "websocket" {
					WsHandlerWithDependencies(c, config)
				}
				return
			}
			// 处理/api/login的POST请求
			if c.Param("filepath") == "/api/login" && c.Request.Method == http.MethodPost {
				HandleLoginRequest(c, config)
				return
			}
			// 处理/api/check-login-status的GET请求
			if c.Param("filepath") == "/api/check-login-status" && c.Request.Method == http.MethodGet {
				HandleCheckLoginStatusRequest(c)
				return
			}
			// 处理 /api/get-json 的GET请求
			if c.Request.URL.Path == "/api/getjson" && c.Request.Method == http.MethodGet {
				HandleGetJSON(c, config)
				return
			}
			// 处理 /api/save-json 的POST请求
			if c.Request.URL.Path == "/api/savejson" && c.Request.Method == http.MethodPost {
				HandleSaveJSON(c, config)
				return
			}
			// 处理 /api/save-json 的POST请求
			if c.Request.URL.Path == "/api/restart" && c.Request.Method == http.MethodPost {
				HandleRestart(c, config)
				return
			}
			// 进程监控
			if c.Param("filepath") == "/api/status" && c.Request.Method == http.MethodGet {
				// 检查操作系统是否既不是 Android 也不是 Darwin (macOS)
				if runtime.GOOS != "android" && runtime.GOOS != "darwin" {
					handleSysInfo(c)
				}
				return
			}
			// 处理 /player 的GET请求
			if c.Request.URL.Path == "/api/player" && c.Request.Method == http.MethodGet {
				listPlayer(c, config, db)
				return
			}
			// 处理 /kickorban 的GET请求
			if c.Request.URL.Path == "/api/kickorban" && c.Request.Method == http.MethodPost {
				handleKickOrBan(c, config)
				return
			}

		} else {
			// 否则，处理静态文件请求
			// 如果请求是 "/webui/" ，默认为 "index.html"
			filepathRequested := c.Param("filepath")
			if filepathRequested == "" || filepathRequested == "/" {
				filepathRequested = "index.html"
			}

			// 从适当的 embed.FS 读取文件内容
			var data []byte
			var err error

			if !strings.HasPrefix(c.Request.URL.Path, "/api") {
				filepathRequested := c.Param("filepath")
				if filepathRequested == "" || filepathRequested == "/" {
					filepathRequested = "index.html"
				} else {
					filepathRequested = strings.TrimPrefix(filepathRequested, "/")
				}

				// 首先尝试从 content 读取文件
				data, err = content.ReadFile("dist/" + filepathRequested)

				// 如果在 dist 中找不到文件，尝试从 dist2 中读取
				if err != nil {
					fmt.Println("Error reading file from dist:", err)

					if strings.HasPrefix(c.Request.URL.Path, "/sav") {
						// 处理 "/sav" 路径
						filepathRequested = strings.TrimPrefix(filepathRequested, "sav/")
					}

					// 尝试从 content2 读取文件
					data, err = content2.ReadFile("dist2/" + filepathRequested)
					if err != nil {
						fmt.Println("Error reading file from dist2:", err)
						c.Status(http.StatusNotFound)
						return
					}
				}
			}

			mimeType := getContentType(filepathRequested)

			c.Data(http.StatusOK, mimeType, data)
		}
		// 调用c.Next()以继续处理请求链
		c.Next()
	}
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// NewRconClient 创建一个新的RCON客户端
func NewRconClient(address, password string) *RconClient {
	conn, err := rcon.Dial(address, password)
	if err != nil {
		log.Printf("无法连接到RCON服务器: %v", err)
		return nil
	}
	return &RconClient{
		Conn: conn,
	}
}

func (c *Client) readPump(config config.Config) {
	defer func() {
		c.conn.Close()
	}()
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			log.Printf("error: %v", err)
			break
		}
		// 初始化RCON客户端
		address := config.Address + ":" + strconv.Itoa(config.WorldSettings.RconPort)
		// 初始化RCON客户端
		rconClient := NewRconClient(address, config.WorldSettings.AdminPassword)
		if rconClient == nil {
			log.Println("RCON客户端初始化失败,无法处理webui面板请求,请按教程正确开启rcon和设置服务端admin密码")
			return
		}
		response, err := rconClient.Conn.Execute(string(message))
		if err != nil {
			log.Printf("RCON execute error: %v", err)
			continue
		}
		c.send <- response
	}
}
func (c *Client) writePump() {
	defer func() {
		c.conn.Close()
	}()
	for {
		message, ok := <-c.send
		if !ok {
			// The channel is closed
			break
		}
		err := c.conn.WriteMessage(websocket.TextMessage, []byte(message))
		if err != nil {
			log.Printf("write error: %v", err)
			break
		}
	}
}

func WsHandlerWithDependencies(c *gin.Context, cfg config.Config) {
	ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println("Upgrade failed:", err)
		return
	}
	client := &Client{
		conn: ws,
		send: make(chan string),
	}
	go client.writePump()
	go client.readPump(cfg)
}

// HandleGetJSON 返回当前的config作为JSON
func HandleGetJSON(c *gin.Context, cfg config.Config) {
	c.JSON(http.StatusOK, cfg)
}

const configFile = "config.json"

// HandleSaveJSON 从请求体中读取JSON并更新config
func HandleSaveJSON(c *gin.Context, cfg config.Config) {

	var newConfig config.Config
	if err := c.ShouldBindJSON(&newConfig); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 调用saveFunc来保存config
	writeConfigToFile(newConfig)
	// 把网页修改的配置刷新到ini
	err := config.WriteGameWorldSettings(&newConfig, newConfig.WorldSettings)
	if err != nil {
		fmt.Println("Error writing game world settings:", err)
	} else {
		fmt.Println("Game world settings saved successfully.")
	}

	c.JSON(http.StatusOK, gin.H{"message": "Config updated successfully"})

	//重启自身 很快 唰的一下
	sys.RestartApplication()

}

func HandleRestart(c *gin.Context, cfg config.Config) {
	// 从请求中获取cookie
	cookieValue, err := c.Cookie("login_cookie")
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: Cookie not provided"})
		return
	}

	// 使用ValidateCookie函数验证cookie
	isValid, err := ValidateCookie(cookieValue)
	if err != nil || !isValid {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: Invalid cookie"})
		return
	}

	// Cookie验证通过后，执行重启操作
	go restartService(cfg)
	c.JSON(http.StatusOK, gin.H{"message": "Restart initiated"})

}

func restartService(cfg config.Config) {
	// 首先，尝试终止同名进程
	if err := sys.KillProcess(); err != nil {
		log.Printf("Failed to kill existing process: %v", err)
		// 可以选择在此处返回，也可以继续尝试启动新进程
	}

	// 构建启动命令
	var exePath string
	var args []string

	if runtime.GOOS == "windows" {
		exePath = filepath.Join(cfg.GamePath, cfg.ProcessName+".exe")
		args = []string{
			"-useperfthreads",
			"-NoAsyncLoadingThread",
			"-UseMultithreadForDS",
			"RconEnabled=True",
			fmt.Sprintf("-AdminPassword=%s", cfg.WorldSettings.AdminPassword),
			fmt.Sprintf("-port=%d", cfg.WorldSettings.PublicPort),
			fmt.Sprintf("-players=%d", cfg.WorldSettings.ServerPlayerMaxNum),
		}
	} else {
		exePath = filepath.Join(cfg.GamePath, cfg.ProcessName)
		args = []string{
			fmt.Sprintf("--port=%d", cfg.WorldSettings.PublicPort),
			fmt.Sprintf("--players=%d", cfg.WorldSettings.ServerPlayerMaxNum),
		}
	}

	log.Printf("webui重启服务端,启动命令: %s %s", exePath, strings.Join(args, " "))
	cmd := exec.Command(exePath, args...)
	cmd.Dir = cfg.GamePath

	if err := cmd.Start(); err != nil {
		log.Printf("Failed to restart game server: %v", err)
	} else {
		log.Printf("Game server restarted successfully")
	}
}

// writeConfigToFile 将配置写回文件
func writeConfigToFile(config config.Config) {
	configJSON, err := json.MarshalIndent(config, "", "    ")
	if err != nil {
		log.Fatalf("无法序列化配置: %v", err)
	}

	err = os.WriteFile(configFile, configJSON, 0644)
	if err != nil {
		log.Fatalf("无法写入配置文件: %v", err)
	}
}

// HandleLoginRequest处理登录请求
func HandleLoginRequest(c *gin.Context, config config.Config) {
	var json struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&json); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if checkCredentials(json.Username, json.Password, config) {
		// 如果验证成功，设置cookie
		cookieValue, err := GenerateCookie()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate cookie"})
			return
		}

		c.SetCookie("login_cookie", cookieValue, 3600*24, "/", "", false, true)

		c.JSON(http.StatusOK, gin.H{
			"isLoggedIn": true,
			"cookie":     cookieValue,
		})
	} else {
		c.JSON(http.StatusUnauthorized, gin.H{
			"isLoggedIn": false,
		})
	}
}

func checkCredentials(username, password string, jsonconfig config.Config) bool {
	serverUsername := jsonconfig.WorldSettings.ServerName
	serverPassword := jsonconfig.WorldSettings.AdminPassword
	fmt.Printf("有用户使用serverUsername:%v serverPassword:%v 进行登入\n", username, password)
	fmt.Printf("登入密码serverUsername:%v serverPassword:%v 进行登入\n", serverUsername, serverPassword)
	return username == serverUsername && password == serverPassword
}

// HandleCheckLoginStatusRequest 检查登录状态的处理函数
func HandleCheckLoginStatusRequest(c *gin.Context) {
	// 从请求中获取cookie
	cookieValue, err := c.Cookie("login_cookie")
	if err != nil {
		// 如果cookie不存在，而不是返回BadRequest(400)，我们返回一个OK(200)的响应
		c.JSON(http.StatusOK, gin.H{"isLoggedIn": false, "error": "Cookie not provided"})
		return
	}

	// 使用ValidateCookie函数验证cookie
	isValid, err := ValidateCookie(cookieValue)
	if err != nil {
		switch err {
		case ErrCookieNotFound:
			c.JSON(http.StatusOK, gin.H{"isLoggedIn": false, "error": "Cookie not found"})
		case ErrCookieExpired:
			c.JSON(http.StatusOK, gin.H{"isLoggedIn": false, "error": "Cookie has expired"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"isLoggedIn": false, "error": "Internal server error"})
		}
		return
	}

	if isValid {
		c.JSON(http.StatusOK, gin.H{"isLoggedIn": true})
	} else {
		c.JSON(http.StatusOK, gin.H{"isLoggedIn": false, "error": "Invalid cookie"})
	}
}

func getContentType(path string) string {
	// todo 根据需要增加更多的 MIME 类型
	switch filepath.Ext(path) {
	case ".html":
		return "text/html"
	case ".js":
		return "application/javascript"
	case ".css":
		return "text/css"
	case ".png":
		return "image/png"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	default:
		return "text/plain"
	}
}

func handleSysInfo(c *gin.Context) {
	// 获取CPU使用率
	cpuPercent, _ := cpu.Percent(time.Second, false)

	// 获取内存信息
	vmStat, _ := mem.VirtualMemory()

	// 获取磁盘使用情况
	diskStat, _ := disk.Usage("/")

	// 获取系统启动时间
	bootTime, _ := host.BootTime()

	// 获取当前进程信息
	proc, _ := process.NewProcess(int32(os.Getpid()))
	procPercent, _ := proc.CPUPercent()
	memInfo, _ := proc.MemoryInfo()
	procStartTime, _ := proc.CreateTime()

	// 构造返回的JSON数据
	sysInfo := gin.H{
		"cpu_percent": cpuPercent[0], // CPU使用率
		"memory": gin.H{
			"total":     vmStat.Total,       // 总内存
			"available": vmStat.Available,   // 可用内存
			"percent":   vmStat.UsedPercent, // 内存使用率
		},
		"disk": gin.H{
			"total":   diskStat.Total,       // 磁盘总容量
			"free":    diskStat.Free,        // 磁盘剩余空间
			"percent": diskStat.UsedPercent, // 磁盘使用率
		},
		"boot_time": bootTime, // 系统启动时间
		"process": gin.H{
			"pid":         proc.Pid,      // 当前进程ID
			"status":      "running",     // 进程状态，这里假设为运行中
			"memory_used": memInfo.RSS,   // 进程使用的内存
			"cpu_percent": procPercent,   // 进程CPU使用率
			"start_time":  procStartTime, // 进程启动时间
		},
	}
	// 返回JSON数据
	c.JSON(http.StatusOK, sysInfo)
}

func listPlayer(c *gin.Context, config config.Config, db *bbolt.DB) {
	update, _ := c.GetQuery("update")
	var currentPlayers []map[string]string
	if update == "true" {
		getCurrentPlayers, err := tool.ShowPlayers(config)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{"error": err.Error()})
			return
		}
		tool.UpdatePlayerData(db, getCurrentPlayers)
		currentPlayers = getCurrentPlayers
	}
	var players []Player
	err := db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("players"))
		return b.ForEach(func(k, v []byte) error {
			var player Player
			if err := json.Unmarshal(v, &player); err != nil {
				return err
			}
			players = append(players, player)
			return nil
		})
	})
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"error": err.Error()})
		return
	}

	// 按 LastOnline 倒序排序
	sort.Slice(players, func(i, j int) bool {
		return players[i].LastOnline.After(players[j].LastOnline)
	})

	// 构建包含所有玩家信息的列表
	allPlayers := make([]map[string]interface{}, 0)
	currentLocalTime := time.Now()
	for _, player := range players {
		diff := currentLocalTime.Sub(player.LastOnline)
		online := false
		if diff < 5*time.Minute {
			online = true
		}
		lastOnlineTimeStr := player.LastOnline.Format("2006-01-02 15:04:05")
		allPlayers = append(allPlayers, map[string]interface{}{
			"name":        player.Name,
			"steamid":     player.SteamID,
			"playeruid":   player.PlayerUID,
			"last_online": lastOnlineTimeStr,
			"online":      online,
		})
	}

	// 标记当前在线的玩家
	if update == "true" {
		for idx, player := range allPlayers {
			for _, currentPlayer := range currentPlayers {
				if player["name"] == currentPlayer["name"] {
					allPlayers[idx]["online"] = true
					break
				}
			}
		}
	}

	c.JSON(http.StatusOK, allPlayers)
}

// Handler for /api/kickorban
func handleKickOrBan(c *gin.Context, config config.Config) {
	steamID := c.Query("steamid")
	actionType := c.Query("type")

	var err error
	if actionType == "kick" {
		err = tool.KickPlayer(config, steamID)
	} else if actionType == "ban" {
		err = tool.BanPlayer(config, steamID)
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid type"})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Success"})
}
