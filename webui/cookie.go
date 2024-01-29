package webui

import (
	"encoding/binary"
	"errors"
	"log"
	"time"

	"github.com/boltdb/bolt"
	"github.com/google/uuid"
)

const (
	DBName          = "cookie.db"
	CookieBucket    = "cookies"
	ExpirationKey   = "expiration"
	ExpirationHours = 30 * 24 // Cookie 有效期改为一个月
)

var dbcookie *bolt.DB
var ErrCookieNotFound = errors.New("cookie not found")
var ErrCookieExpired = errors.New("cookie has expired")

func InitializeDB() {
	var err error
	dbcookie, err = bolt.Open(DBName, 0600, nil)
	if err != nil {
		log.Fatalf("Error opening DB: %v", err)
	}

	dbcookie.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(CookieBucket))
		return err
	})
}

func CloseDB() {
	dbcookie.Close()
}

func GenerateCookie() (string, error) {
	cookie := uuid.New().String()
	expiration := time.Now().Add(ExpirationHours * time.Hour).Unix()

	err := dbcookie.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(CookieBucket))
		if err := bucket.Put([]byte(cookie), intToBytes(expiration)); err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return "", err
	}

	return cookie, nil
}

func ValidateCookie(cookie string) (bool, error) {
	isValid := false
	err := dbcookie.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(CookieBucket))
		expBytes := bucket.Get([]byte(cookie))
		if expBytes == nil {
			return ErrCookieNotFound
		}

		expiration := bytesToInt(expBytes)
		if time.Now().Unix() > expiration {
			return ErrCookieExpired
		}

		isValid = true
		return nil
	})

	return isValid, err
}

func intToBytes(n int64) []byte {
	bytes := make([]byte, 8)
	binary.BigEndian.PutUint64(bytes, uint64(n))
	return bytes
}

func bytesToInt(b []byte) int64 {
	return int64(binary.BigEndian.Uint64(b))
}
