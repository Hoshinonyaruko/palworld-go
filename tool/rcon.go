// https://github.com/zaigie/palworld-server-tool/tree/main
package tool

import (
	"errors"
	"fmt"
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

	lines := strings.Split(response, "\n")
	titles := strings.Split(lines[0], ",")
	result := make([]map[string]string, 0)
	for _, line := range lines[1:] {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		fields := strings.Split(line, ",")
		dataMap := make(map[string]string)
		for i, title := range titles {
			value := "<null/err>"
			if i < len(fields) {
				value = fields[i]
				if strings.Contains(value, "\u0000") {
					// Usually \u0000 is an error
					value = "<null/err>"
				}
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
	address := config.Address + ":" + strconv.Itoa(config.WorldSettings.RconPort)
	exec, err := NewExecutor(address, config.WorldSettings.AdminPassword, true)
	if err != nil {
		return err
	}
	defer exec.Close()

	message = strings.ReplaceAll(message, " ", "_")

	response, err := exec.Execute("Broadcast " + message)
	if err != nil {
		return err
	}
	if response != fmt.Sprintf("Broadcasted: %s", message) {
		return errors.New(response)
	}
	return nil
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
