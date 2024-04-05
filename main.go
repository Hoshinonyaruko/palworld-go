package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"embed"
	"encoding/pem"
	"fmt"
	"io"
	"io/fs"
	"log"
	"math/big"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"go.etcd.io/bbolt"

	"github.com/hoshinonyaruko/palworld-go/bot"
	"github.com/hoshinonyaruko/palworld-go/config"
	"github.com/hoshinonyaruko/palworld-go/status"
	"github.com/hoshinonyaruko/palworld-go/sys"
	"github.com/hoshinonyaruko/palworld-go/tool"
	"github.com/hoshinonyaruko/palworld-go/webui"
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
	fmt.Printf("作者 早苗狐 答疑群:587997911\n")
	//给程序整个标题
	sys.SetTitle(jsonconfig.Title + " 作者 早苗狐 答疑群:587997911")

	//为小白和萌新贴心打造的引导启动流程，解决小白不会寻找路径，不会复制文件的问题。点击即用，友好方便。
	if runtime.GOOS == "windows" {
		//检查程序是否运行在合适目录
		if _, err := os.Stat("PalServer.exe"); os.IsNotExist(err) {
			// PalServer.exe不存在，查找PalServer-Win64-Shipping-Cmd.exe的路径
			cmd := exec.Command("cmd", "/C", "wmic process where name='PalServer-Win64-Shipping-Cmd.exe' get ExecutablePath")
			output, err := cmd.Output()
			if err != nil {
				fmt.Println("Failed to execute command:", err)
				return
			}

			outputStr := string(output)
			if !strings.Contains(outputStr, "ExecutablePath") {
				fmt.Println("PalServer-Win64-Shipping-Cmd.exe not found")
				fmt.Println("Notice:Please restart this program after launching the game server, and the installation will be completed automatically, releasing a launch icon to the desktop.")
				fmt.Println("请打开游戏服务端后再次运行本程序，将会自动完成安装，释放启动图标到桌面。")
				showMessage("提示", "请打开游戏服务端保持运行后的同时运行本程序,\n将会自动寻找游戏路径,\n自动完成安装,并释放启动图标到桌面。\n请到桌面使用StartPalWorld.bat启动我。\nNotice:Please restart this program after launching the game server, and the installation will be completed automatically, releasing a launch icon to the desktop.")
				return
			}

			lines := strings.Split(outputStr, "\n")
			for _, line := range lines {
				if strings.TrimSpace(line) != "" && !strings.Contains(line, "ExecutablePath") {
					exePath := strings.TrimSpace(line)
					// 获取到的exe路径向上移动到PalServer这个级别
					targetDir := filepath.Dir(filepath.Dir(filepath.Dir(filepath.Dir(exePath)))) // 假设路径格式没有变化
					targetExePath := filepath.Join(targetDir, "palworld-go.exe")

					// 复制自身到目标目录
					self, _ := os.Executable()
					target, err := os.Create(targetExePath)
					if err != nil {
						fmt.Println("Failed to create target file:", err)
						return
					}
					defer target.Close()

					source, err := os.Open(self)
					if err != nil {
						fmt.Println("Failed to open source file:", err)
						return
					}
					defer source.Close()

					_, err = io.Copy(target, source)
					if err != nil {
						fmt.Println("Failed to copy file:", err)
						return
					}

					//创建启动脚本
					createBATScript(targetDir, "palworld-go.exe")

					fmt.Println("PalWorld-Go setup completed successfully.")
					showMessage("提示", "现在请关闭服务端窗口！,并在桌面找到脚本启动palworld-go(会自动打开服务端).\n安装成功,请从桌面名为 StartPalWorld.bat 的脚本图标启动pal-go面板。\n请在5秒内点击确认按钮~确保文件能够释放成功~\nPalWorld-Go setup completed successfully. please launch palworld-go form the desktop.")
					deleteSelfAndConfigs()
					return
				}
			}
		} else {
			fmt.Println("PalServer.exe exists in the current directory.")
		}
	}
	//还原状态
	status.SetManualServerShutdown(false)

	// 设置监控和自动重启
	supervisor := NewSupervisor(jsonconfig)
	go supervisor.Start()

	// 设置备份任务
	backupTask := NewBackupTask(jsonconfig)
	go backupTask.Schedule()

	if !supervisor.isServiceRunning() {
		sys.RestartService(jsonconfig)
	} else {
		fmt.Printf("当前服务端正常运行中,守护和内存助手已启动\n")
	}
	//cookie数据库
	webui.InitializeDB()
	//玩家数据库
	db = webui.InitDB()
	//机器人数据库
	if jsonconfig.Onebotv11HttpApiPath != "" {
		bot.InitializeDB()
	}
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

	if jsonconfig.UseHttps && jsonconfig.Cert == "" && jsonconfig.Key == "" {
		//创造自签名证书
		priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		if err != nil {
			panic(err)
		}
		randomOrg := generateRandomString(10)
		template := x509.Certificate{
			SerialNumber: big.NewInt(1),
			Subject: pkix.Name{
				Organization: []string{"Palworld-go-" + randomOrg},
			},
			NotBefore: time.Now(),
			NotAfter:  time.Now().Add(365 * 24 * time.Hour),

			KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
			ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
			BasicConstraintsValid: true,
		}
		publicip, err := sys.GetPublicIP()
		if err != nil {
			fmt.Println("获取当前地址生成https证书失败")
		}
		ipAddresses := []net.IP{net.ParseIP("127.0.0.1"), net.ParseIP(publicip)}
		template.IPAddresses = append(template.IPAddresses, ipAddresses...)

		derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
		if err != nil {
			panic(err)
		}

		certOut, err := os.Create("cert.pem")
		if err != nil {
			panic(err)
		}
		pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
		certOut.Close()

		// 编码 ECDSA 私钥
		encodedKey, err := x509.MarshalECPrivateKey(priv)
		if err != nil {
			panic(err) // 或适当的错误处理
		}

		// 创建 PEM 文件
		keyOut, err := os.Create("key.pem")
		if err != nil {
			panic(err) // 或适当的错误处理
		}
		defer keyOut.Close() // 确保文件被正确关闭

		// 将编码后的私钥写入 PEM 文件
		if err := pem.Encode(keyOut, &pem.Block{Type: "EC PRIVATE KEY", Bytes: encodedKey}); err != nil {
			panic(err) // 或适当的错误处理
		}
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
			if jsonconfig.Cert != "" && jsonconfig.Key != "" {
				certFile = jsonconfig.Cert
				keyFile = jsonconfig.Key
			}
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
	fmt.Printf("开放52000端口后可外网访问,用户名,服务器名(可以中文),初始用户名palgo初始密码useradmin\n")
	fmt.Printf("为了防止误修改,52000端口仅可在config.json修改\n")
	if jsonconfig.AutolaunchWebui {
		OpenWebUI(&jsonconfig)
	}

	latestTag, err := sys.GetLatestTag("sanaefox/palworld-go")
	if err != nil {
		fmt.Println("Error fetching latest tag:", err)
	}

	fmt.Printf("当前版本: %s 最新版本: %s \n", version, latestTag)

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
				if jsonconfig.EnableEngineSetting {
					//保存引擎配置
					err = config.WriteEngineSettings(&jsonconfig, jsonconfig.Engine)
					if err != nil {
						fmt.Println("Error writing Engine settings:", err)
					} else {
						fmt.Println("Engine settings saved successfully.")
					}
				}
			}
		}()
	}

	if jsonconfig.WhiteCheckTime != 0 {
		//白名单
		whiteInterval := time.Duration(jsonconfig.WhiteCheckTime) * time.Second
		whiteTicker := time.NewTicker(whiteInterval)
		go func() {
			defer whiteTicker.Stop()
			for range whiteTicker.C {
				fmt.Println("checking player whitelist")
				tool.CheckAndKickPlayers(jsonconfig)
			}
		}()
	}

	//定时重启
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

func showMessage(title, text string) error {
	// PowerShell脚本来显示消息框
	psScript := `
Add-Type -AssemblyName PresentationFramework
[System.Windows.MessageBox]::Show('` + text + `', '` + title + `')
`
	cmd := exec.Command("powershell", "-Command", psScript)
	err := cmd.Start() // 使用Start而不是Run，以非阻塞方式执行
	if err != nil {
		return err
	}
	return nil
}

func createBATScript(targetDir, targetExeName string) error {
	desktopPathCmd := exec.Command("powershell", "-Command", "[System.Environment]::GetFolderPath('Desktop')")
	desktopPathBytes, err := desktopPathCmd.Output()
	if err != nil {
		return fmt.Errorf("failed to get desktop path: %v", err)
	}
	desktopPath := string(desktopPathBytes)
	desktopPath = desktopPath[:len(desktopPath)-2] // 移除末尾的\r\n

	batContent := fmt.Sprintf(`
@echo off
cd /d "%s"
start "" "%s"
`, targetDir, filepath.Join(targetDir, targetExeName))

	batFilePath := filepath.Join(desktopPath, "StartPalWorld.bat")
	err = os.WriteFile(batFilePath, []byte(batContent), 0644)
	if err != nil {
		return fmt.Errorf("failed to create BAT file: %v", err)
	}

	return nil
}

func deleteSelfAndConfigs() error {
	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %v", err)
	}
	dir := filepath.Dir(exePath)

	// 构造PowerShell脚本命令
	psScript := fmt.Sprintf(`
Start-Sleep -Seconds 5
Remove-Item '%s' -Force
Remove-Item '%s\\config.json' -Force
Remove-Item '%s\\config.ini' -Force
`, exePath, dir, dir)

	cmd := exec.Command("powershell", "-Command", psScript)

	// 开始执行PowerShell命令，不等待其完成
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start PowerShell script: %v", err)
	}

	// 退出当前程序，让PowerShell脚本有机会删除文件
	os.Exit(0)

	return nil
}
