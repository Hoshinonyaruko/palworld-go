package main

import (
	"io"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/hoshinonyaruko/palworld-go/config"
)

type BackupTask struct {
	Config config.Config
	Ticker *time.Ticker
}

func NewBackupTask(config config.Config) *BackupTask {
	return &BackupTask{
		Config: config,
		Ticker: time.NewTicker(time.Duration(config.BackupInterval) * time.Second),
	}
}

func (task *BackupTask) Schedule() {
	for range task.Ticker.C {
		task.RunBackup()
	}
}

func (task *BackupTask) RunBackup() {
	// 获取当前日期和时间
	currentDate := time.Now().Format("2006-01-02-15-04-05")

	// 创建新的备份目录
	backupDir := filepath.Join(task.Config.BackupPath, currentDate)
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		log.Printf("Failed to create backup directory: %v", err)
		return
	}

	// 确定源文件的路径和目标路径
	sourcePath := filepath.Join(task.Config.GameSavePath, "SaveGames")
	destinationPath := filepath.Join(backupDir, "SaveGames")

	// 执行文件复制操作
	if err := copyDir(sourcePath, destinationPath); err != nil {
		log.Printf("Failed to copy files for backup SaveGames: %v", err)
	} else {
		log.Printf("Backup completed successfully: %s", destinationPath)
	}

	// 确定源文件的路径和目标路径
	sourcePath = filepath.Join(task.Config.GameSavePath, "Config")
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
