// https://github.com/zaigie/palworld-server-tool/tree/main
package tool

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/gorcon/rcon"
	"github.com/hoshinonyaruko/palworld-go/config"
)

var (
	ErrPasswordEmpty = errors.New("未设置密码，检查config.yaml中的password配置")
)

type ExecuteCloser interface {
	Execute(command string) (string, error)
	Close() error
}

type Executor struct {
	skipErrors bool
	client     ExecuteCloser
}

var timeout int = 10

func NewExecutor(address, password string, skipErrors bool) (*Executor, error) {
	var client ExecuteCloser
	var err error

	if password == "" {
		return nil, ErrPasswordEmpty
	}

	timeoutDuration := time.Duration(timeout) * time.Second

	client, err = rcon.Dial(address, password, rcon.SetDialTimeout(timeoutDuration), rcon.SetDeadline(timeoutDuration))

	if err != nil {
		return nil, err
	}

	return &Executor{client: client, skipErrors: skipErrors}, nil
}

func (e *Executor) Execute(command string) (string, error) {

	response, err := e.client.Execute(command)

	if response != "" {
		response = strings.TrimSpace(response)
		if err != nil && e.skipErrors {
			return response, nil
		}
	}

	return response, err
}

func (e *Executor) Close() error {
	if e.client != nil {
		return e.client.Close()
	}
	return nil
}

// UpdateServer 使用SteamCMD更新服务端
func CreateAndRunPSScript(config config.Config) error {
	scriptContent := fmt.Sprintf("$SteamCmdPath = \"%s\\steamcmd.exe\"\n& $SteamCmdPath +login anonymous +app_update 2394010 validate +quit", config.SteamCmdPath)
	scriptPath := filepath.Join(os.TempDir(), "update-server.ps1")

	// 创建.ps1文件
	err := os.WriteFile(scriptPath, []byte(scriptContent), 0644)
	if err != nil {
		return fmt.Errorf("failed to create PowerShell script: %w", err)
	}

	// 使用cmd /C start powershell来在新窗口中执行.ps1脚本
	cmd := exec.Command("cmd", "/C", "start", "powershell", "-ExecutionPolicy", "Bypass", "-File", scriptPath)
	err = cmd.Start()
	if err != nil {
		return fmt.Errorf("failed to execute PowerShell script: %w", err)
	}

	log.Printf("PowerShell script started in a new window: %s", scriptPath)
	return nil
}
