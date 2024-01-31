package mod

import (
	"crypto/md5"
	"embed"
	"encoding/hex"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

//go:embed embeds/*
var embeddedFiles embed.FS

// CheckAndWriteFiles 检查嵌入文件是否存在于指定路径下，对于exe和dll文件，如果不存在或者MD5不同，则写出；对于其他文件，如果不存在，则写出。
func CheckAndWriteFiles(path string) error {
	return fs.WalkDir(embeddedFiles, "embeds", func(embeddedPath string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// 跳过根目录
		if embeddedPath == "embeds" {
			return nil
		}

		// 从嵌入路径中移除“embeds”前缀
		relativePath := strings.TrimPrefix(embeddedPath, "embeds/")

		// 构建在外部文件系统中的路径
		externalPath := filepath.Join(path, relativePath)

		if d.IsDir() {
			// 如果是目录，则创建（如果不存在）
			return os.MkdirAll(externalPath, os.ModePerm)
		} else {
			// 处理文件
			return processFile(embeddedPath, externalPath)
		}
	})
}

// processFile 处理单个文件：根据文件类型和MD5决定是否写出
func processFile(embeddedPath, externalPath string) error {
	// 检查文件类型
	ext := strings.ToLower(filepath.Ext(externalPath))
	isExecutableOrDLL := ext == ".exe" || ext == ".dll"

	// 读取嵌入的文件内容
	embeddedData, err := fs.ReadFile(embeddedFiles, embeddedPath)
	if err != nil {
		return err
	}

	// 对于exe和dll文件，比较MD5，如果不同则写出；对于其他文件类型，如果文件不存在，则写出
	if isExecutableOrDLL {
		// 计算嵌入文件的MD5
		embeddedMD5 := calculateMD5(embeddedData)

		// 检查文件是否存在并比较MD5
		existingMD5, err := fileMD5(externalPath)
		if err == nil {
			// 如果MD5相同，则跳过写出
			if existingMD5 == embeddedMD5 {
				return nil
			}
		} // 如果文件不存在或无法读取MD5，则继续写出
	} else {
		// 对于非exe和dll文件，如果文件已存在，则跳过
		if _, err := os.Stat(externalPath); err == nil {
			return nil
		}
	}

	// 确保目标文件夹存在
	if err := os.MkdirAll(filepath.Dir(externalPath), os.ModePerm); err != nil {
		return err
	}

	// 写出文件
	return os.WriteFile(externalPath, embeddedData, os.ModePerm)
}

// calculateMD5 计算数据的MD5值
func calculateMD5(data []byte) string {
	hash := md5.Sum(data)
	return hex.EncodeToString(hash[:])
}

// fileMD5 计算文件的MD5值
func fileMD5(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			// 如果文件不存在，返回空字符串和nil错误
			return "", nil
		}
		return "", err
	}
	defer file.Close()

	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}
