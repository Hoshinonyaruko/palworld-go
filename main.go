package main

import (
	"crypto/rand"
	"embed"
	"fmt"
	"io/fs"
	"log"
	"math/big"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hoshinonyaruko/palworld-go/config"
	"github.com/hoshinonyaruko/palworld-go/sys"
	"github.com/hoshinonyaruko/palworld-go/tool"
	"github.com/hoshinonyaruko/palworld-go/webui"
	"go.etcd.io/bbolt"
)

var version string

var db *bbolt.DB

//go:embed RAMMap64.exe
var rammapFS embed.FS

func main() {
	// 读取或创建配置
	jsonconfig := config.ReadConfig()

	// 打印配置以确认
	fmt.Printf("当前配置: %#v\n", jsonconfig)
	//给程序整个标题
	sys.SetTitle("Palworld管理")

	// 设置监控和自动重启
	supervisor := NewSupervisor(jsonconfig)
	go supervisor.Start()

	// 设置备份任务
	backupTask := NewBackupTask(jsonconfig)
	go backupTask.Schedule()

	if !supervisor.isServiceRunning() {
		supervisor.restartService()
	} else {
		fmt.Printf("当前服务端正常运行中,守护和内存助手已启动\n")
	}
	//cookie数据库
	webui.InitializeDB()
	//玩家数据库
	db = webui.InitDB()
	//启动周期任务
	go tool.ScheduleTask(db, jsonconfig)
	if db == nil {
		log.Fatal("Failed to initialize database")
	}
	defer db.Close()
	r := gin.Default()

	//webui和它的api
	webuiGroup := r.Group("/")
	{
		webuiGroup.GET("/*filepath", webui.CombinedMiddleware(jsonconfig, db))
		webuiGroup.POST("/*filepath", webui.CombinedMiddleware(jsonconfig, db))
		webuiGroup.PUT("/*filepath", webui.CombinedMiddleware(jsonconfig, db))
		webuiGroup.DELETE("/*filepath", webui.CombinedMiddleware(jsonconfig, db))
		webuiGroup.PATCH("/*filepath", webui.CombinedMiddleware(jsonconfig, db))
	}

	// 创建一个http.Server实例(主服务器)
	httpServer := &http.Server{
		Addr:    "0.0.0.0:" + jsonconfig.WebuiPort,
		Handler: r,
	}

	if jsonconfig.UseHttps {
		fmt.Printf("webui-api运行在 HTTPS 端口 %v\n", jsonconfig.WebuiPort)
		// 在一个新的goroutine中启动主服务器
		go func() {
			// 定义默认的证书和密钥文件名 自签名证书
			certFile := "cert.pem"
			keyFile := "key.pem"
			// 使用 HTTPS
			if err := httpServer.ListenAndServeTLS(certFile, keyFile); err != nil && err != http.ErrServerClosed {
				log.Fatalf("listen: %s\n", err)
			}

		}()
	} else {
		fmt.Printf("webui-api运行在 HTTP 端口 %v\n", jsonconfig.WebuiPort)
		// 在一个新的goroutine中启动主服务器
		go func() {
			// 使用HTTP
			if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				log.Fatalf("listen: %s\n", err)
			}
		}()
	}

	// 设置推送任务
	palworldBroadcast := NewpalworldBroadcast(jsonconfig)
	go palworldBroadcast.Schedule()

	// 设置内存检查任务
	memoryCheckTask := NewMemoryCheckTask(jsonconfig, backupTask)
	go memoryCheckTask.Schedule()
	fmt.Printf("webui-api运行在%v端口\n", jsonconfig.WebuiPort)
	fmt.Printf("webui地址:http://127.0.0.1:%v\n", jsonconfig.WebuiPort)
	fmt.Printf("开放8000端口后可外网访问,用户名,服务器名(可以中文),用户名【你的服务器名称】密码【你的管理员密码】\n")
	fmt.Printf("为了防止误修改,8000端口仅可在config.json修改\n")
	if jsonconfig.AutolaunchWebui {
		OpenWebUI(&jsonconfig)
	}

	if runtime.GOOS == "windows" {
		if jsonconfig.MemoryCleanupInterval != 0 {
			log.Printf("你决定使用rammap清理内存....这不会导致游戏卡顿\n")

			// 提取并保存RAMMap到临时文件
			rammapExecutable, err := extractRAMMapExecutable()
			if err != nil {
				log.Fatalf("无法提取RAMMap可执行文件: %v", err)
			}
			defer os.Remove(rammapExecutable) // 确保程序结束时删除文件

			// 创建定时器，根据配置间隔定期运行RAMMap
			ticker := time.NewTicker(time.Duration(jsonconfig.MemoryCleanupInterval) * time.Second)
			go func() {
				defer ticker.Stop()
				for range ticker.C {
					runRAMMap(rammapExecutable)
				}
			}()
		}
	}

	if runtime.GOOS == "windows" {
		// 创建一个定时器，每10秒触发一次，保存游戏设置，允许玩家修改json配置并同步到ini
		saveSettingsTicker := time.NewTicker(10 * time.Second)
		go func() {
			defer saveSettingsTicker.Stop()
			for range saveSettingsTicker.C {
				// 定时保存配置
				jsonconfig := config.ReadConfigv2()
				//保存帕鲁服务端配置
				err := config.WriteGameWorldSettings(&jsonconfig, jsonconfig.WorldSettings)
				if err != nil {
					fmt.Println("Error writing game world settings:", err)
				} else {
					fmt.Println("Game world settings saved successfully.")
				}
				//保存引擎配置
				err = config.WriteEngineSettings(&jsonconfig, jsonconfig.Engine)
				if err != nil {
					fmt.Println("Error writing Engine settings:", err)
				} else {
					fmt.Println("Engine settings saved successfully.")
				}
			}
		}()
	}

	if jsonconfig.RestartInterval != 0 {
		restartInterval := time.Duration(jsonconfig.RestartInterval) * time.Second
		restartTicker := time.NewTicker(restartInterval)
		go func() {
			defer restartTicker.Stop()
			for range restartTicker.C {
				// 定时推送并重启 120秒 发数组第一条信息
				tool.Shutdown(jsonconfig, "120", jsonconfig.RegularMessages[0])
			}
		}()
	}

	// 设置信号捕获
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// 等待信号
	<-sigChan
	if runtime.GOOS == "windows" {
		// 接收到退出信号，写回配置，守护退出会刷新游戏ini
		jsonconfig := config.ReadConfigv2()
		err := config.WriteGameWorldSettings(&jsonconfig, jsonconfig.WorldSettings)
		if err != nil {
			// 处理写回错误
			fmt.Println("Error writing game world settings:", err)
		} else {
			fmt.Println("Success writing game world settings")
		}
	}

	// 正常退出程序
	os.Exit(0)

}

// extractRAMMapExecutable 从嵌入的文件系统中提取RAMMap并写入临时文件
func extractRAMMapExecutable() (string, error) {
	rammapData, err := fs.ReadFile(rammapFS, "RAMMap64.exe")
	if err != nil {
		return "", err
	}

	tmpFile, err := os.CreateTemp("", "RAMMap64-*.exe")
	if err != nil {
		return "", err
	}
	defer tmpFile.Close()

	if _, err := tmpFile.Write(rammapData); err != nil {
		return "", err
	}

	return tmpFile.Name(), nil
}

func runRAMMap(rammapExecutable string) {
	log.Printf("正在使用rammap清理内存....")
	// 调用RAMMap的命令
	cmd := exec.Command(rammapExecutable, "-Ew")
	err := cmd.Run()
	if err != nil {
		log.Printf("运行RAMMap时发生错误: %v", err)
	}
}

// OpenWebUI 在默认浏览器中打开Web UI
func OpenWebUI(config *config.Config) error {
	url := fmt.Sprintf("http://127.0.0.1:%s", config.WebuiPort)
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", url)
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	default:
		return fmt.Errorf("不支持的操作系统: %s", runtime.GOOS)
	}

	return cmd.Start()
}

func generateRandomString(n int) string {
	var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	b := make([]rune, n)
	for i := range b {
		randInt, _ := rand.Int(rand.Reader, big.NewInt(int64(len(letters))))
		b[i] = letters[randInt.Int64()]
	}
	return string(b)
}
