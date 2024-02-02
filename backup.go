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
	var ticker *time.Ticker
	if config.BackupInterval > 0 {
		ticker = time.NewTicker(time.Duration(config.BackupInterval) * time.Second)
	}

	return &BackupTask{
		Config: config,
		Ticker: ticker,
	}
}

func (task *BackupTask) Schedule() {
	if task.Ticker == nil {
		// 如果 Ticker 为 nil，不需要进行定时备份
		return
	}

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

	// 删除旧备份（如果设置了天数）
	if task.Config.SaveDeleteDays > 0 {
		task.deleteOldBackups()
	}

}

func (task *BackupTask) deleteOldBackups() {
	// 读取备份目录
	files, err := os.ReadDir(task.Config.BackupPath)
	if err != nil {
		log.Printf("Failed to list backup directory: %v", err)
		return
	}

	// 删除超过SaveDeleteDays天数的备份
	for _, f := range files {
		if f.IsDir() {
			backupTime, err := time.Parse("2006-01-02-15-04-05", f.Name())
			if err != nil {
				log.Printf("Failed to parse backup directory name: %s, error: %v", f.Name(), err)
				continue
			}

			if time.Since(backupTime).Hours() > float64(task.Config.SaveDeleteDays*24) {
				err := os.RemoveAll(filepath.Join(task.Config.BackupPath, f.Name()))
				if err != nil {
					log.Printf("Failed to delete old backup: %s, error: %v", f.Name(), err)
				} else {
					log.Printf("Old backup deleted successfully: %s", f.Name())
				}
			}
		}
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
