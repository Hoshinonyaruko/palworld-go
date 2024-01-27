// https://github.com/zaigie/palworld-server-tool/tree/main
package tool

import (
	"encoding/json"
	"log"
	"strings"
	"time"

	"github.com/hoshinonyaruko/palworld-go/config"
	"go.etcd.io/bbolt"
)

type Player struct {
	Name       string    `json:"name"`
	SteamID    string    `json:"steamid"`
	PlayerUID  string    `json:"playeruid"`
	LastOnline time.Time `json:"last_online"`
}

func ScheduleTask(db *bbolt.DB, config config.Config) {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		players, err := ShowPlayers(config)
		if err != nil {
			log.Println("Error fetching players:", err)
			continue
		}
		UpdatePlayerData(db, players)
		log.Println("Schedule Updated Player Data")
	}
}

func UpdatePlayerData(db *bbolt.DB, playersData []map[string]string) {
	err := db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("players"))
		for _, playerData := range playersData {
			if playerData["name"] == "<null/err>" {
				continue
			}
			existingPlayerData := b.Get([]byte(playerData["name"]))
			var player Player
			if existingPlayerData != nil {
				if err := json.Unmarshal(existingPlayerData, &player); err != nil {
					return err
				}

				if player.SteamID == "<null/err>" || strings.Contains(player.SteamID, "000000") {
					player.SteamID = playerData["steamid"]
				}
				if player.PlayerUID == "<null/err>" || strings.Contains(player.PlayerUID, "000000") {
					player.PlayerUID = playerData["playeruid"]
				}
				player.LastOnline = time.Now()
			} else {
				player = Player{
					Name:       playerData["name"],
					SteamID:    playerData["steamid"],
					PlayerUID:  playerData["playeruid"],
					LastOnline: time.Now(),
				}
			}

			serializedPlayer, err := json.Marshal(player)
			if err != nil {
				return err
			}
			if err := b.Put([]byte(player.Name), serializedPlayer); err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		log.Println("Error updating player:", err)
	}
}
