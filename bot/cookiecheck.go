package bot

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"log"

	"github.com/boltdb/bolt"
)

var (
	db *bolt.DB
)

const (
	DBName    = "bot.db"
	BotBucket = "botdata"
)

func InitializeDB() {
	var err error
	db, err = bolt.Open(DBName, 0600, nil)
	if err != nil {
		log.Fatalf("Error opening DB: %v", err)
	}

	db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(BotBucket))
		return err
	})
}

// CloseDatabase 提供关闭数据库的方法
func CloseDatabase() error {
	if db != nil {
		return db.Close()
	}
	return nil
}

// CheckAndWriteCookie 检查cookie是否存在，如果不存在则写入数据库
func CheckAndWriteCookie(value string) (bool, error) {
	if db == nil {
		return false, errors.New("database not initialized")
	}

	// 检查键是否存在
	exists := false
	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(BotBucket))
		if b != nil {
			v := b.Get([]byte(value))
			exists = v != nil
		}
		return nil
	})
	if err != nil {
		return false, err
	}
	if exists {
		return true, nil
	}

	// 键不存在，写入新键值对
	err = db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte(BotBucket))
		if err != nil {
			return err
		}
		return b.Put([]byte(value), []byte("true"))
	})
	return false, err
}

type UserIPData struct {
	IP    string `json:"ip"`
	UUID  string `json:"uuid"`
	Https bool   `json:"https"`
}

// StoreUserIDAndIP 用于存储用户ID（int64）、IP地址（string）和UUID（string）
func StoreUserIDAndIP(userID int64, ip string, uuid string, https bool) error {
	if db == nil {
		return errors.New("database not initialized")
	}

	userData := UserIPData{
		IP:    ip,
		UUID:  uuid,
		Https: https,
	}
	userDataBytes, err := json.Marshal(userData)
	if err != nil {
		return err
	}

	return db.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte("ipuserdata"))
		if err != nil {
			return err
		}
		userIDBytes := make([]byte, 8)
		binary.BigEndian.PutUint64(userIDBytes, uint64(userID))
		return bucket.Put(userIDBytes, userDataBytes)
	})
}

// RetrieveIPByUserID 用于通过用户ID（int64）检索IP地址（string）和UUID（string）
func RetrieveIPByUserID(userID int64) (UserIPData, error) {
	var userData UserIPData
	if db == nil {
		return userData, errors.New("database not initialized")
	}

	err := db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("ipuserdata"))
		if bucket == nil {
			return fmt.Errorf("bucket %s not found", "ipuserdata")
		}
		userIDBytes := make([]byte, 8)
		binary.BigEndian.PutUint64(userIDBytes, uint64(userID))
		value := bucket.Get(userIDBytes)
		if value == nil {
			return fmt.Errorf("no data found for user ID %d", userID)
		}
		return json.Unmarshal(value, &userData)
	})

	if err != nil {
		return userData, err
	}
	return userData, nil
}

type PlayerInfo struct {
	PlayerUID  string `json:"playerUID"`
	SteamID    string `json:"steamID"`
	Name       string `json:"name"`
	Online     bool   `json:"online"`
	LastOnline string `json:"last_online"`
}

// StorePlayerInfo 存储玩家信息并返回唯一ID
func StorePlayerInfo(playerUID, steamID, name string) (int64, error) {
	if db == nil {
		return 0, errors.New("database not initialized")
	}

	playerInfo := PlayerInfo{
		PlayerUID: playerUID,
		SteamID:   steamID,
		Name:      name,
	}

	var id int64
	err := db.Update(func(tx *bolt.Tx) error {
		// 创建或获取PlayerList bucket
		bucket, err := tx.CreateBucketIfNotExists([]byte("playerlist"))
		if err != nil {
			return err
		}

		// 检查是否已存在相同的玩家信息
		c := bucket.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			var existingPlayerInfo PlayerInfo
			if err := json.Unmarshal(v, &existingPlayerInfo); err != nil {
				continue
			}
			if existingPlayerInfo.PlayerUID == playerUID && existingPlayerInfo.SteamID == steamID && existingPlayerInfo.Name == name {
				// 解析ID
				id = int64(binary.BigEndian.Uint64(k))
				return nil
			}
		}

		// 分配新的唯一ID
		seq, _ := bucket.NextSequence()
		id = int64(seq) // 转换为 int64
		idBytes := make([]byte, 8)
		binary.BigEndian.PutUint64(idBytes, uint64(id))
		playerInfoBytes, err := json.Marshal(playerInfo)
		if err != nil {
			return err
		}
		return bucket.Put(idBytes, playerInfoBytes)
	})

	return id, err
}

// RetrievePlayerInfoByID 通过ID检索玩家信息
func RetrievePlayerInfoByID(id int64) (PlayerInfo, error) {
	var playerInfo PlayerInfo
	if db == nil {
		return playerInfo, errors.New("database not initialized")
	}

	err := db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("playerlist"))
		if bucket == nil {
			return fmt.Errorf("bucket %s not found", "playerlist")
		}
		idBytes := make([]byte, 8)
		binary.BigEndian.PutUint64(idBytes, uint64(id))
		value := bucket.Get(idBytes)
		if value == nil {
			return fmt.Errorf("no player found for ID %d", id)
		}
		return json.Unmarshal(value, &playerInfo)
	})

	return playerInfo, err
}
