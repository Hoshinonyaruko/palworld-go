package webui

import (
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorcon/rcon"
	"github.com/gorilla/websocket"
	"github.com/hoshinonyaruko/palworld-go/config"
	"github.com/hoshinonyaruko/palworld-go/mod"
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

type KickOrBanRequest struct {
	PlayerUID string `json:"playeruid"`
	SteamID   string `json:"steamid"`
	Type      string `json:"type"`
}

// ChangeSaveRequest 用于解析请求体
type ChangeSaveRequest struct {
	Path string `json:"path"`
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
			// 处理 /api/save-json 的POST请求
			if c.Request.URL.Path == "/api/start" && c.Request.Method == http.MethodPost {
				HandleStart(c, config)
				return
			}
			// 处理 /api/save-json 的POST请求
			if c.Request.URL.Path == "/api/stop" && c.Request.Method == http.MethodPost {
				HandleStop(c, config)
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
			// 处理 /kickorban 的POST请求
			if c.Request.URL.Path == "/api/kickorban" && c.Request.Method == http.MethodPost {
				handleKickOrBan(c, config, db)
				return
			}
			// 处理 /getsavelist 的POST请求
			if c.Request.URL.Path == "/api/getsavelist" && c.Request.Method == http.MethodGet {
				handleGetSavelist(c, config)
				return
			}
			// 处理 /changesave 的POST请求
			if c.Request.URL.Path == "/api/changesave" && c.Request.Method == http.MethodPost {
				handleChangeSave(c, config)
				return
			}
			// 处理 /savenow 的POST请求
			if c.Request.URL.Path == "/api/savenow" && c.Request.Method == http.MethodPost {
				handleSaveNow(c, config)
				return
			}
			// 处理 /delsave 的POST请求
			if c.Request.URL.Path == "/api/delsave" && c.Request.Method == http.MethodPost {
				handleDelSave(c, config)
				return
			}
			// 处理 /restartlater 的POST请求 过一段时间重启
			if c.Request.URL.Path == "/api/restartlater" && c.Request.Method == http.MethodPost {
				handleRestartLater(c, config)
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

	err = config.WriteEngineSettings(&newConfig, newConfig.Engine)
	if err != nil {
		fmt.Println("Error writing Engine settings:", err)
	} else {
		fmt.Println("Engine settings saved successfully.")
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
	go restartService(cfg, true)
	c.JSON(http.StatusOK, gin.H{"message": "Restart initiated"})

}

func HandleStart(c *gin.Context, cfg config.Config) {
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
	go restartService(cfg, false)
	c.JSON(http.StatusOK, gin.H{"message": "start initiated"})

}

func HandleStop(c *gin.Context, cfg config.Config) {
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

	// 终止进程
	if err := sys.KillProcess(); err != nil {
		log.Printf("Failed to kill existing process: %v", err)
		// 可以选择在此处返回，也可以继续尝试启动新进程
	}
	c.JSON(http.StatusOK, gin.H{"message": "Stop initiated"})

}

func restartService(cfg config.Config, kill bool) {
	//结束以前的服务端
	if kill {
		// 首先，尝试终止同名进程
		if err := sys.KillProcess(); err != nil {
			log.Printf("Failed to kill existing process: %v", err)
			// 可以选择在此处返回，也可以继续尝试启动新进程
		}
	}

	var exePath string
	var args []string

	if runtime.GOOS == "windows" {
		if cfg.CommunityServer {
			exePath = filepath.Join(cfg.SteamPath, "Steam.exe")
			args = []string{"-applaunch", "2394010"}
		} else if cfg.UseDll {
			err := mod.CheckAndWriteFiles(filepath.Join(cfg.GamePath, "Pal", "Binaries", "Win64"))
			if err != nil {
				log.Printf("Failed to write files: %v", err)
				return
			}
			exePath = filepath.Join(cfg.GamePath, "Pal", "Binaries", "Win64", "PalServerInject.exe")
			args = []string{
				"-RconEnabled=True",
				fmt.Sprintf("-AdminPassword=%s", cfg.WorldSettings.AdminPassword),
				fmt.Sprintf("-port=%d", cfg.WorldSettings.PublicPort),
				fmt.Sprintf("-players=%d", cfg.WorldSettings.ServerPlayerMaxNum),
			}
		} else {
			exePath = filepath.Join(cfg.GamePath, cfg.ProcessName+".exe")
			args = []string{
				"-RconEnabled=True",
				fmt.Sprintf("-AdminPassword=%s", cfg.WorldSettings.AdminPassword),
				fmt.Sprintf("-port=%d", cfg.WorldSettings.PublicPort),
				fmt.Sprintf("-players=%d", cfg.WorldSettings.ServerPlayerMaxNum),
			}
		}
	} else {
		exePath = filepath.Join(cfg.GamePath, cfg.ProcessName+".sh")
		args = []string{
			"--RconEnabled=True",
			fmt.Sprintf("--AdminPassword=%s", cfg.WorldSettings.AdminPassword),
			fmt.Sprintf("--port=%d", cfg.WorldSettings.PublicPort),
			fmt.Sprintf("--players=%d", cfg.WorldSettings.ServerPlayerMaxNum),
		}
	}

	args = append(args, cfg.ServerOptions...) // 添加GameWorldSettings参数

	// 执行启动命令
	log.Printf("启动命令: %s %s", exePath, strings.Join(args, " "))
	if cfg.UseDll && runtime.GOOS == "windows" {
		log.Printf("use bat")
		sys.RunViaBatch(cfg, exePath, args)
	} else {
		cmd := exec.Command(exePath, args...)
		cmd.Dir = cfg.GamePath // 设置工作目录为游戏路径

		// 启动进程
		if err := cmd.Start(); err != nil {
			log.Printf("Failed to restart game server: %v", err)
		} else {
			log.Printf("Game server restarted successfully")
		}
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

	var currentPlayersMap map[string]bool
	if update == "true" {
		getCurrentPlayers, err := tool.ShowPlayers(config)
		if err != nil {
			// Log the error instead of returning it
			log.Println("Error fetching current players:", err)

			// Initialize currentPlayersMap as empty if fetching online players fails
			currentPlayersMap = make(map[string]bool)
		} else {
			tool.UpdatePlayerData(db, getCurrentPlayers)

			// Create a map for quick lookup
			currentPlayersMap = make(map[string]bool)
			for _, player := range getCurrentPlayers {
				currentPlayersMap[player["name"]] = true
			}
		}
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

	// Sort players
	sort.Slice(players, func(i, j int) bool {
		return players[i].LastOnline.After(players[j].LastOnline)
	})

	currentLocalTime := time.Now()
	allPlayers := make([]map[string]interface{}, len(players))

	for i, player := range players {
		diff := currentLocalTime.Sub(player.LastOnline)
		online := diff < 3*time.Minute

		if update == "true" {
			// Check if the player is in the currentPlayersMap for online status
			if _, exists := currentPlayersMap[player.Name]; exists {
				online = true
			}
		}

		allPlayers[i] = map[string]interface{}{
			"name":        player.Name,
			"steamid":     player.SteamID,
			"playeruid":   player.PlayerUID,
			"last_online": player.LastOnline.Format("2006-01-02 15:04:05"),
			"online":      online,
		}
	}

	c.JSON(http.StatusOK, allPlayers)
}

// Handler for /api/kickorban
func handleKickOrBan(c *gin.Context, config config.Config, db *bbolt.DB) {
	var req KickOrBanRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	var err error
	if req.Type == "kick" {
		err = tool.KickPlayer(config, req.SteamID)
	} else if req.Type == "ban" {
		err = tool.BanPlayer(config, req.SteamID)
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid type"})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	tool.UpdateLastOnlineForPlayer(db, req.SteamID)
	c.JSON(http.StatusOK, gin.H{"message": "Success"})
}

// handleGetSavelist 处理 /api/getsavelist 请求
func handleGetSavelist(c *gin.Context, config config.Config) {
	// 获取保存路径
	savePath := config.BackupPath

	if savePath == "" && runtime.GOOS != "windows" {
		savePath = "."
	}

	// 正则表达式匹配特定的日期时间格式
	regex, err := regexp.Compile(`^\d{4}-\d{2}-\d{2}-\d{2}-\d{2}-\d{2}$`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Regex compile error"})
		return
	}

	// 枚举文件夹并匹配
	var folders []string
	err = filepath.Walk(savePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() && regex.MatchString(info.Name()) {
			folders = append(folders, info.Name())
		}
		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 反转文件夹列表
	for i, j := 0, len(folders)-1; i < j; i, j = i+1, j-1 {
		folders[i], folders[j] = folders[j], folders[i]
	}

	// 返回符合条件的文件夹名称
	c.JSON(http.StatusOK, folders)
}

// handleChangeSave 处理 /api/changesave 请求
func handleChangeSave(c *gin.Context, config config.Config) {
	var req ChangeSaveRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	// 首先，尝试终止同名进程
	if err := sys.KillProcess(); err != nil {
		log.Printf("Failed to kill existing process: %v", err)
		// 可以选择在此处返回，也可以继续尝试启动新进程
	}

	// 检查源路径是否存在
	sourcePath := filepath.Join(config.BackupPath, req.Path, "SaveGames", "0")
	if _, err := os.Stat(sourcePath); os.IsNotExist(err) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Source save path does not exist"})
		return
	}

	// 获取源路径中的哈希文件夹名称
	sourceHashFolder, err := getHashFolderName(sourcePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 获取目标路径中的哈希文件夹名称
	destPath := filepath.Join(config.GamePath, "Pal", "Saved", "SaveGames", "0")
	destHashFolder, err := getHashFolderName(destPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 复制文件
	err = copyDir(filepath.Join(sourcePath, sourceHashFolder), filepath.Join(destPath, destHashFolder))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 存档更换成功后
	go restartService(config, true)

	c.JSON(http.StatusOK, gin.H{"message": "Save changed successfully"})
}

// getHashFolderName 获取哈希命名的文件夹名称
func getHashFolderName(path string) (string, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return "", err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			// 假设每个子目录都是哈希命名的
			return entry.Name(), nil
		}
	}

	return "", errors.New("no hash folder found")
}

// runBackup 执行备份操作
func runBackup(config config.Config) {
	// 获取当前日期和时间
	currentDate := time.Now().Format("2006-01-02-15-04-05")

	// 创建新的备份目录
	backupDir := filepath.Join(config.BackupPath, currentDate)
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		log.Printf("Failed to create backup directory: %v", err)
		return
	}

	// 确定源文件的路径和目标路径
	sourcePath := filepath.Join(config.GameSavePath, "SaveGames")
	destinationPath := filepath.Join(backupDir, "SaveGames")

	// 执行文件复制操作
	if err := copyDir(sourcePath, destinationPath); err != nil {
		log.Printf("Failed to copy files for backup SaveGames: %v", err)
	} else {
		log.Printf("Backup completed successfully: %s", destinationPath)
	}

	// 确定源文件的路径和目标路径
	sourcePath = filepath.Join(config.GameSavePath, "Config")
	destinationPath = filepath.Join(backupDir, "Config")

	// 执行文件复制操作
	if err := copyDir(sourcePath, destinationPath); err != nil {
		log.Printf("Failed to copy files for backup Config: %v", err)
	} else {
		log.Printf("Backup completed successfully: %s", destinationPath)
	}
}

// copyDir 递归复制目录及其内容
func copyDir(src string, dst string) error {
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(dst, srcInfo.Mode()); err != nil {
		return err
	}

	dir, _ := os.Open(src)
	defer dir.Close()
	entries, _ := dir.Readdir(-1)

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			if err := copyDir(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			if err := copyFile(srcPath, dstPath); err != nil {
				return err
			}
		}
	}
	return nil
}

// copyFile 复制单个文件
func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	if _, err = io.Copy(dstFile, srcFile); err != nil {
		return err
	}

	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	return os.Chmod(dst, srcInfo.Mode())
}

// SaveNowRequest 用于解析请求体
type SaveNowRequest struct {
	Timestamp int64 `json:"timestamp"`
}

// handleSaveNow 处理 /api/savenow 请求
func handleSaveNow(c *gin.Context, config config.Config) {
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

	// 解析请求体
	var req SaveNowRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	// 校验时间戳
	currentTime := time.Now().Unix()
	if abs(currentTime-req.Timestamp) > 10 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid timestamp"})
		return
	}

	// 执行备份操作
	go runBackup(config)
	c.JSON(http.StatusOK, gin.H{"message": "Backup initiated"})
}

// abs 返回绝对值
func abs(n int64) int64 {
	if n < 0 {
		return -n
	}
	return n
}

// handleDelSave 处理 /api/delsave 请求
func handleDelSave(c *gin.Context, config config.Config) {
	// 从请求体直接读取文件名数组
	var files []string
	if err := c.BindJSON(&files); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	savePath := config.BackupPath

	for _, file := range files {
		// 构建完整的文件路径
		filePath := filepath.Join(savePath, file)

		// 检查文件是否存在
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "File does not exist: " + file})
			return
		}

		// 删除文件
		if err := os.Remove(filePath); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete file: " + file})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "Files deleted successfully"})
}

// BroadcastRequest 用于解析传入的JSON请求体
type BroadcastRequest struct {
	Message string `json:"message"`
}

// handleBroadcast 处理 /api/broadcast 的POST请求
func handleBroadcast(c *gin.Context, config config.Config) {
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

	var req BroadcastRequest

	// 绑定JSON请求体到req
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	// 调用 tool.Broadcast 发送广播
	err = tool.Broadcast(config, req.Message)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Broadcast sent successfully"})
}

// RestartLaterRequest 用于绑定JSON请求体
type RestartLaterRequest struct {
	Seconds string `json:"seconds"`
	Message string `json:"message"`
}

// handleRestartLater 处理 /api/restartlater 的POST请求
func handleRestartLater(c *gin.Context, config config.Config) {
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

	var req RestartLaterRequest

	// 绑定JSON请求体到req
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	// 调用tool.Shutdown来安排重启
	err = tool.Shutdown(config, req.Seconds, req.Message)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Restart scheduled successfully"})
}
