package utils

import (
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/KamyangTech/xt-pms/server/config"
	"golang.org/x/image/draw"
)

const ThumbnailSuffix = "_thumb"

// imageExtensions 需要生成缩略图的图片扩展名
// webp 和 gif 本身压缩率较高，跳过缩略图生成以简化实现
var imageExtensions = map[string]bool{
	".jpg":  true,
	".jpeg": true,
	".png":  true,
}

// GenerateThumbnail 为指定图片生成缩略图，返回缩略图的 URL 路径
// 缩略图命名规则：原文件名 + _thumb 后缀，如 abc_photo.jpg -> abc_photo_thumb.jpg
// 非图片文件或生成失败时返回空字符串（不阻断上传流程）
func GenerateThumbnail(storagePath, urlPath string) (thumbnailURL string, err error) {
	ext := strings.ToLower(filepath.Ext(storagePath))
	if !imageExtensions[ext] {
		return "", nil // 非图片文件，跳过缩略图生成
	}

	maxWidth := config.GlobalConfig.Upload.ThumbnailMaxWidth
	maxHeight := config.GlobalConfig.Upload.ThumbnailMaxHeight
	if maxWidth <= 0 {
		maxWidth = 200
	}
	if maxHeight <= 0 {
		maxHeight = 200
	}

	// 打开原图
	srcFile, err := os.Open(storagePath)
	if err != nil {
		return "", fmt.Errorf("打开原图失败: %w", err)
	}
	defer srcFile.Close()

	// 解码图片
	srcImg, _, err := decodeImage(srcFile, ext)
	if err != nil {
		return "", fmt.Errorf("解码图片失败: %w", err)
	}

	srcBounds := srcImg.Bounds()
	srcW := srcBounds.Dx()
	srcH := srcBounds.Dy()

	// 如果原图已经很小，直接返回原图 URL 作为缩略图
	if srcW <= maxWidth && srcH <= maxHeight {
		return urlPath, nil
	}

	thumbW, thumbH := calcThumbnailSize(srcW, srcH, maxWidth, maxHeight)

	thumbImg := image.NewRGBA(image.Rect(0, 0, thumbW, thumbH))
	draw.CatmullRom.Scale(thumbImg, thumbImg.Bounds(), srcImg, srcBounds, draw.Over, nil)

	thumbName := getThumbnailName(filepath.Base(storagePath))
	thumbDir := filepath.Dir(storagePath)
	thumbPath := filepath.Join(thumbDir, thumbName)

	outFile, err := os.Create(thumbPath)
	if err != nil {
		return "", fmt.Errorf("创建缩略图文件失败: %w", err)
	}
	defer outFile.Close()

	switch ext {
	case ".png":
		if err := png.Encode(outFile, thumbImg); err != nil {
			return "", fmt.Errorf("编码 PNG 缩略图失败: %w", err)
		}
	default:
		if err := jpeg.Encode(outFile, thumbImg, &jpeg.Options{Quality: 80}); err != nil {
			return "", fmt.Errorf("编码 JPEG 缩略图失败: %w", err)
		}
	}

	// 使用 path 而非 filepath 以确保 URL 始终使用 / 分隔符
	thumbURLName := getThumbnailName(path.Base(urlPath))
	thumbURL := path.Join(path.Dir(urlPath), thumbURLName)

	return thumbURL, nil
}

func decodeImage(r io.Reader, ext string) (image.Image, string, error) {
	switch ext {
	case ".png":
		img, err := png.Decode(r)
		return img, "png", err
	case ".jpg", ".jpeg":
		img, err := jpeg.Decode(r)
		return img, "jpeg", err
	default:
		return nil, "", fmt.Errorf("不支持的图片格式: %s", ext)
	}
}

func calcThumbnailSize(srcW, srcH, maxW, maxH int) (w, h int) {
	w = srcW
	h = srcH

	if w > maxW {
		h = h * maxW / w
		w = maxW
	}
	if h > maxH {
		w = w * maxH / h
		h = maxH
	}
	return w, h
}

// getThumbnailName 获取缩略图文件名
// 例如: abc_photo.jpg -> abc_photo_thumb.jpg
func getThumbnailName(originalName string) string {
	ext := path.Ext(originalName)
	base := strings.TrimSuffix(originalName, ext)
	return base + ThumbnailSuffix + ext
}

// GetThumbnailURL 根据原图 URL 推导缩略图 URL
// 例如: /uploads/ab/cd/uuid_name.jpg -> /uploads/ab/cd/uuid_name_thumb.jpg
func GetThumbnailURL(originalURL string) string {
	if originalURL == "" {
		return ""
	}
	return getThumbnailName(originalURL)
}

type ThumbnailResult struct {
	Total     int `json:"total"`
	Generated int `json:"generated"`
	Skipped   int `json:"skipped"`
	Failed    int `json:"failed"`
}

// GenerateMissingThumbnails 批量扫描图片 URL 列表，为缺失缩略图的文件生成缩略图
// basePath 为上传根目录（如 ./uploads），urls 为需要检查的图片 URL 列表
func GenerateMissingThumbnails(basePath string, urls []string, logFn func(msg string, fields ...interface{})) ThumbnailResult {
	result := ThumbnailResult{Total: len(urls)}

	for _, urlPath := range urls {
		thumbnailURL := GetThumbnailURL(urlPath)
		if thumbnailURL == "" {
			result.Skipped++
			continue
		}

		thumbDiskPath := filepath.Join(basePath, thumbnailURL)
		if _, err := os.Stat(thumbDiskPath); err == nil {
			result.Skipped++
			continue
		}

		// 原图路径
		originalDiskPath := filepath.Join(basePath, urlPath)
		if _, err := os.Stat(originalDiskPath); os.IsNotExist(err) {
			result.Failed++
			if logFn != nil {
				logFn("原图文件不存在，无法生成缩略图", "path", originalDiskPath)
			}
			continue
		}

		// 确保目录存在
		thumbDir := filepath.Dir(thumbDiskPath)
		if err := os.MkdirAll(thumbDir, 0755); err != nil {
			result.Failed++
			if logFn != nil {
				logFn("创建缩略图目录失败", "dir", thumbDir, "error", err)
			}
			continue
		}

		if _, err := GenerateThumbnail(originalDiskPath, urlPath); err != nil {
			result.Failed++
			if logFn != nil {
				logFn("生成缩略图失败", "url", urlPath, "error", err)
			}
		} else {
			result.Generated++
		}
	}

	if logFn != nil {
		logFn("缩略图批量生成完成",
			"total", result.Total,
			"generated", result.Generated,
			"skipped", result.Skipped,
			"failed", result.Failed,
		)
	}
	return result
}
