// https://github.com/zaigie/palworld-server-tool/tree/main
package tool

import (
	"encoding/json"
	"fmt"
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

type PlayerW struct {
	Name      string `json:"name"`
	SteamID   string `json:"steamid"`
	PlayerUID string `json:"playeruid"`
	Online    bool   `json:"online"`
}

func ScheduleTask(db *bbolt.DB, config config.Config) {
	ticker := time.NewTicker(3 * time.Minute)
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
	err := db.Batch(func(tx *bbolt.Tx) error { // Use Batch for better performance
		b := tx.Bucket([]byte("players"))
		for _, playerData := range playersData {
			if playerData["name"] == "<null/err>" {
				continue
			}

			// Check if player exists and update only if necessary
			needUpdate := false
			existingPlayerData := b.Get([]byte(playerData["name"]))
			var player Player
			if existingPlayerData != nil {
				if err := json.Unmarshal(existingPlayerData, &player); err != nil {
					return err
				}

				// Update fields only if they are different
				if player.SteamID != playerData["steamid"] && (player.SteamID == "<null/err>" || strings.Contains(player.SteamID, "000000")) {
					player.SteamID = playerData["steamid"]
					needUpdate = true
				}
				if player.PlayerUID != playerData["playeruid"] && (player.PlayerUID == "<null/err>" || strings.Contains(player.PlayerUID, "000000")) {
					player.PlayerUID = playerData["playeruid"]
					needUpdate = true
				}
				player.LastOnline = time.Now() // This might be always updated depending on your business logic
			} else {
				player = Player{
					Name:       playerData["name"],
					SteamID:    playerData["steamid"],
					PlayerUID:  playerData["playeruid"],
					LastOnline: time.Now(),
				}
				needUpdate = true
			}

			if needUpdate {
				serializedPlayer, err := json.Marshal(player)
				if err != nil {
					return err
				}
				if err := b.Put([]byte(player.Name), serializedPlayer); err != nil {
					return err
				}
			}
		}
		return nil
	})

	if err != nil {
		log.Println("Error updating player:", err)
	}
}

func UpdateLastOnlineForPlayer(db *bbolt.DB, steamID string) error {
	tenMinutesAgo := time.Now().Add(-10 * time.Minute)

	return db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("players"))

		// Iterate over all players to find the one with the given SteamID
		cursor := b.Cursor()
		for k, v := cursor.First(); k != nil; k, v = cursor.Next() {
			var player Player
			if err := json.Unmarshal(v, &player); err != nil {
				return err
			}

			if player.SteamID == steamID {
				// Update the LastOnline field
				player.LastOnline = tenMinutesAgo

				// Serialize the updated player data
				updatedPlayerData, err := json.Marshal(player)
				if err != nil {
					return err
				}

				// Save the updated player data back to the database
				return b.Put(k, updatedPlayerData)
			}
		}

		return fmt.Errorf("player with SteamID %s not found", steamID)
	})
}
