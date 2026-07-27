package utils

import (
	"crypto/sha256"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"

	"github.com/KamyangTech/xt-pms/server/config"
)

// SanitizeFilename 清理文件名中的不安全字符，只保留字母数字下划线连字符中文
func SanitizeFilename(name string) string {
	var result []rune
	for _, r := range name {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' || r == '-' || r == '.' || (r >= 0x4e00 && r <= 0x9fff) {
			result = append(result, r)
		} else {
			result = append(result, '_')
		}
	}
	return string(result)
}

// TruncateName 截取文件名前 n 个字符（按 rune，避免截断多字节 UTF-8 字符）
func TruncateName(name string, n int) string {
	runes := []rune(name)
	if len(runes) > n {
		return string(runes[:n])
	}
	return name
}

// GenerateStoragePath 根据 fileID hash 生成两级分片目录 + 存储路径
// 路径结构: uploads/{hash[:2]}/{hash[2:4]}/{fileID}{ext}
// 磁盘存储使用纯 UUID 命名，避免文件名编码问题；下载时通过 Content-Disposition 返回原始文件名
// 最多 256*256=65536 个子目录，每子目录文件均匀分布，无热点问题
func GenerateStoragePath(fileID, ext string) (dir, filePath, urlPath string) {
	h := sha256.Sum256([]byte(fileID))
	dir1 := fmt.Sprintf("%02x", h[0])
	dir2 := fmt.Sprintf("%02x", h[1])

	fileName := fileID + ext
	dir = filepath.Join(config.GlobalConfig.Upload.BasePath, dir1, dir2)
	filePath = filepath.Join(dir, fileName)
	urlPath = fmt.Sprintf("/uploads/%s/%s/%s", dir1, dir2, fileName)
	return
}

// SaveUploadedFile 将上传的文件保存到磁盘（不依赖 gin.Context）
func SaveUploadedFile(file *multipart.FileHeader, storagePath string) error {
	if err := os.MkdirAll(filepath.Dir(storagePath), 0755); err != nil {
		return fmt.Errorf("创建目录失败: %w", err)
	}

	src, err := file.Open()
	if err != nil {
		return fmt.Errorf("打开上传文件失败: %w", err)
	}
	defer src.Close()

	dst, err := os.Create(storagePath)
	if err != nil {
		return fmt.Errorf("创建目标文件失败: %w", err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return fmt.Errorf("保存文件失败: %w", err)
	}

	return nil
}
