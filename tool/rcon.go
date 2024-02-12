// https://github.com/zaigie/palworld-server-tool/tree/main
package tool

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/hoshinonyaruko/palworld-go/config"
)

func Info(config config.Config) (map[string]string, error) {
	address := config.Address + ":" + strconv.Itoa(config.WorldSettings.RconPort)
	exec, err := NewExecutor(address, config.WorldSettings.AdminPassword, true)
	if err != nil {
		return nil, err
	}
	defer exec.Close()

	response, err := exec.Execute("Info")
	if err != nil {
		return nil, err
	}
	re := regexp.MustCompile(`\[(v[\d\.]+)\]\s*(.+)`)
	matches := re.FindStringSubmatch(response)
	if matches == nil || len(matches) < 3 {
		return map[string]string{
			"version": "unknown",
			"name":    "unknown",
		}, nil
	}
	result := map[string]string{
		"version": matches[1],
		"name":    matches[2],
	}
	return result, nil
}

func ShowPlayers(config config.Config) ([]map[string]string, error) {
	address := config.Address + ":" + strconv.Itoa(config.WorldSettings.RconPort)
	exec, err := NewExecutor(address, config.WorldSettings.AdminPassword, true)
	if err != nil {
		return nil, err
	}
	defer exec.Close()

	response, err := exec.Execute("ShowPlayers")
	if err != nil {
		return nil, err
	}

	//第一行指令标题 然后才是内容
	lines := strings.Split(response, "\n")
	if len(lines) < 2 {
		return nil, fmt.Errorf("invalid response format")
	}

	titles := strings.Split(lines[0], ",")
	result := make([]map[string]string, 0, len(lines)-1)
	for _, line := range lines[1:] {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		fields := strings.Split(line, ",")
		dataMap := make(map[string]string, len(titles))
		for i, title := range titles {
			value := "<null/err>"
			if i < len(fields) && !strings.Contains(fields[i], "\u0000") {
				value = fields[i]
			}
			dataMap[title] = value
		}
		result = append(result, dataMap)
	}

	return result, nil
}

func KickPlayer(config config.Config, steamID string) error {
	address := config.Address + ":" + strconv.Itoa(config.WorldSettings.RconPort)
	exec, err := NewExecutor(address, config.WorldSettings.AdminPassword, true)
	if err != nil {
		return err
	}
	defer exec.Close()

	response, err := exec.Execute("KickPlayer " + steamID)
	if err != nil {
		return err
	}
	if response != fmt.Sprintf("Kicked: %s", steamID) {
		return errors.New(response)
	}
	return nil
}

func BanPlayer(config config.Config, steamID string) error {
	address := config.Address + ":" + strconv.Itoa(config.WorldSettings.RconPort)
	exec, err := NewExecutor(address, config.WorldSettings.AdminPassword, true)
	if err != nil {
		return err
	}
	defer exec.Close()

	response, err := exec.Execute("BanPlayer " + steamID)
	if err != nil {
		return err
	}
	if response != fmt.Sprintf("Banned: %s", steamID) {
		return errors.New(response)
	}
	return nil
}

func Broadcast(config config.Config, message string) error {
	// 如果config.Usedll为true，则使用HTTP方式发送消息
	if config.UseDll {
		base := "http://127.0.0.0.1:" + config.DllPort + "/rcon?text="
		messageText := url.QueryEscape("broadcast " + message)
		fullURL := base + messageText

		// 发送HTTP请求
		resp, err := http.Get(fullURL)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		// 这里可以添加对resp的处理逻辑
		log.Println("Broadcast message sent successfully via HTTP")
		return nil
	} else {
		// 原始方法发送
		address := config.Address + ":" + strconv.Itoa(config.WorldSettings.RconPort)
		exec, err := NewExecutor(address, config.WorldSettings.AdminPassword, true)
		if err != nil {
			return err
		}
		defer exec.Close()

		response, err := exec.Execute("Broadcast " + strings.ReplaceAll(message, " ", "_"))
		if err != nil {
			return err
		}
		if response != fmt.Sprintf("Broadcasted: %s", message) {
			return errors.New(response)
		}
		return nil
	}
}

func Shutdown(config config.Config, seconds string, message string) error {
	address := config.Address + ":" + strconv.Itoa(config.WorldSettings.RconPort)
	exec, err := NewExecutor(address, config.WorldSettings.AdminPassword, true)
	if err != nil {
		return err
	}
	defer exec.Close()

	message = strings.ReplaceAll(message, " ", "_")

	response, err := exec.Execute(fmt.Sprintf("Shutdown %s %s", seconds, message))
	if err != nil {
		return err
	}
	if response != fmt.Sprintf("Shutdown: %s", message) {
		// return errors.New(response)
		return nil // HACK: Not Tested
	}
	return nil
}

func DoExit(config config.Config) error {
	address := config.Address + ":" + strconv.Itoa(config.WorldSettings.RconPort)
	exec, err := NewExecutor(address, config.WorldSettings.AdminPassword, true)
	if err != nil {
		return err
	}
	defer exec.Close()

	response, err := exec.Execute("DoExit")
	if err != nil {
		return err
	}
	if response != "Exited" {
		// return errors.New(response)
		return nil // Hack: Not Tested
	}
	return nil
}

func CheckAndKickPlayers(config config.Config) {
	if len(config.Players) == 0 {
		return // 白名单为空时不执行操作
	}

	apiURL := fmt.Sprintf("http://127.0.0.1:%s/api/player?update=true", config.WebuiPort)
	resp, err := http.Get(apiURL)
	if err != nil {
		log.Printf("获取玩家信息失败: %v", err)
		return
	}
	defer resp.Body.Close()

	var players []PlayerW
	if err := json.NewDecoder(resp.Body).Decode(&players); err != nil {
		log.Printf("解析玩家信息失败: %v", err)
		return
	}

	for _, player := range players {
		if player.Online && !IsPlayerInWhitelist(player, config.Players) {
			// 玩家在线但不在白名单，执行踢出操作
			if err := KickPlayer(config, player.SteamID); err != nil {
				log.Printf("踢出玩家失败: %v", err)
			} else {
				log.Printf("踢出玩家%v成功: %v", player.Name, err)
			}
		}
	}
}

func IsPlayerInWhitelist(player PlayerW, whitelist []*config.PlayerW) bool {
	for _, wp := range whitelist {
		if (wp.Name == "" || wp.Name == player.Name) &&
			(wp.SteamID == "" || wp.SteamID == player.SteamID) &&
			(wp.PlayerUID == "" || wp.PlayerUID == player.PlayerUID) {
			return true
		}
	}
	return false
}
