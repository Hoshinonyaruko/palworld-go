package mod

import (
	"embed"
	"io/fs"
	"os"
	"path/filepath"
)

//go:embed embeds/*
var embeddedFiles embed.FS

// CheckAndWriteFiles 检查嵌入文件是否存在于指定路径下，如果不存在，则写出这些文件。
func CheckAndWriteFiles(path string) error {
	return fs.WalkDir(embeddedFiles, "embeds", func(embeddedPath string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// 跳过根目录
		if embeddedPath == "embeds" {
			return nil
		}

		// 构建在外部文件系统中的路径
		externalPath := filepath.Join(path, filepath.Base(embeddedPath))

		// 检查文件是否存在
		if _, err := os.Stat(externalPath); os.IsNotExist(err) {
			// 文件不存在，需要写出
			if d.IsDir() {
				// 创建目录
				if err := os.MkdirAll(externalPath, os.ModePerm); err != nil {
					return err
				}
			} else {
				// 读取嵌入的文件内容
				data, err := fs.ReadFile(embeddedFiles, embeddedPath)
				if err != nil {
					return err
				}

				// 写出文件
				if err := os.WriteFile(externalPath, data, os.ModePerm); err != nil {
					return err
				}
			}
		}

		return nil
	})
}
