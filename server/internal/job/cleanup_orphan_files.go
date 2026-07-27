package job

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/KamyangTech/xt-pms/server/config"
	"github.com/KamyangTech/xt-pms/server/internal/dao"
	"github.com/KamyangTech/xt-pms/server/internal/model/entity"
	"github.com/KamyangTech/xt-pms/server/pkg/logger"
	"go.uber.org/zap"
)

// 自动清理保留天数配置（线程安全）
var (
	cleanupMu               sync.RWMutex
	autoCleanupRetentionDays = 7
	cleanupRunningMu        sync.Mutex // 防止并发执行清理任务
)

func GetRetentionDays() int {
	cleanupMu.RLock()
	defer cleanupMu.RUnlock()
	return autoCleanupRetentionDays
}

func SetRetentionDays(days int) {
	cleanupMu.Lock()
	defer cleanupMu.Unlock()
	autoCleanupRetentionDays = days
	logger.Info("[文件清理] 自动清理保留天数已更新", zap.Int("days", days))
}

// CleanupOrphanFiles 清理所有孤文件（无保留期，供"立即清理"按钮使用）
func CleanupOrphanFiles() {
	cleanupFiles(0)
}

// CleanupOrphanFilesWithRetention 清理超过保留期的孤文件（供定时任务使用）
func CleanupOrphanFilesWithRetention(days int) {
	cleanupFiles(days)
}

func cleanupFiles(days int) {
	// 互斥锁防止并发执行清理任务
	if !cleanupRunningMu.TryLock() {
		logger.Warn("[文件清理] 清理任务已在执行中，跳过本次请求")
		return
	}
	defer cleanupRunningMu.Unlock()

	if days > 0 {
		logger.Info("[文件清理] 开始扫描孤文件（保留期>=天）", zap.Int("retention_days", days))
	} else {
		logger.Info("[文件清理] 开始扫描孤文件（立即清理全部）")
	}

	fileDAO := dao.NewFileDAO()

	// 先标记所有 ref_count=0 但未软删除的文件
	if err := fileDAO.MarkDeleted(dao.GetDB()); err != nil {
		logger.Warn("[文件清理] 标记软删除失败", zap.Error(err))
	}

	var files []entity.File
	var err error

	if days > 0 {
		before := time.Now().AddDate(0, 0, -days)
		files, err = fileDAO.FindOrphanFiles(before)
	} else {
		files, err = fileDAO.FindAllOrphanFiles()
	}

	if err != nil {
		logger.Error("[文件清理] 查询孤文件失败", zap.Error(err))
		return
	}

	if len(files) > 0 {
		cleanedCount := 0
		for _, f := range files {
			imagePath := filepath.Join(".", strings.TrimPrefix(f.URL, "/"))
			if err := os.Remove(imagePath); err != nil && !os.IsNotExist(err) {
				logger.Warn("[文件清理] 删除物理文件失败",
					zap.String("url", f.URL),
					zap.String("file_id", f.ID),
					zap.Error(err))
				continue
			}

			if err := fileDAO.HardDeleteByID(f.ID); err != nil {
				logger.Warn("[文件清理] 删除DB记录失败",
					zap.String("file_id", f.ID),
					zap.Error(err))
				continue
			}

			cleanedCount++
			logger.Info("[文件清理] 已清理孤文件",
				zap.String("url", f.URL),
				zap.String("file_name", f.FileName),
				zap.Time("deleted_at", *f.DeletedAt))
		}

		logger.Info("[文件清理] 数据库孤文件清理完成",
			zap.Int("total", len(files)),
			zap.Int("cleaned", cleanedCount))
	} else {
		logger.Info("[文件清理] 未发现数据库孤文件，继续扫描磁盘")
	}

	// 第二阶段：扫描磁盘上无数据库记录的孤文件
	cleanupDiskOrphans(fileDAO)
}

func CountDiskOrphanFiles() int {
	basePath := getBasePath()

	dbURLs := make(map[string]struct{})
	fileURLs, err := dao.NewFileDAO().FindAllFileURLs()
	if err != nil {
		logger.Error("[文件清理] 查询数据库文件URL失败", zap.Error(err))
		return 0
	}
	for _, url := range fileURLs {
		dbURLs[url] = struct{}{}
	}

	count := 0
	filepath.Walk(basePath, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		relPath, _ := filepath.Rel(filepath.Dir(basePath), path)
		urlPath := "/" + filepath.ToSlash(relPath)
		if _, exists := dbURLs[urlPath]; !exists {
			count++
		}
		return nil
	})
	return count
}

func getBasePath() string {
	basePath := config.GlobalConfig.Upload.BasePath
	if !strings.HasPrefix(basePath, "./") && !strings.HasPrefix(basePath, "/") {
		basePath = "." + string(os.PathSeparator) + basePath
	}
	return basePath
}

// cleanupDiskOrphans 扫描 uploads 目录，删除数据库中无对应记录的物理文件
func cleanupDiskOrphans(fileDAO *dao.FileDAO) {
	basePath := getBasePath()

	logger.Info("[文件清理] 开始扫描磁盘孤文件", zap.String("base_path", basePath))

	dbURLs := make(map[string]struct{})
	fileURLs, err := fileDAO.FindAllFileURLs()
	if err != nil {
		logger.Error("[文件清理] 查询数据库文件URL失败", zap.Error(err))
		return
	}
	for _, url := range fileURLs {
		dbURLs[url] = struct{}{}
	}

	const thumbnailSuffix = "_thumb"

	deletedCount := 0
	err = filepath.Walk(basePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // 跳过无法访问的目录
		}
		if info.IsDir() {
			return nil // 只处理文件
		}

		relPath, _ := filepath.Rel(filepath.Dir(basePath), path)
		urlPath := "/" + filepath.ToSlash(relPath)

		// 缩略图文件没有独立数据库记录，以原图是否存在为判断依据
		if strings.Contains(info.Name(), thumbnailSuffix) {
			// 推导对应的原图 URL：xxx_thumb.jpg -> xxx.jpg
			ext := filepath.Ext(urlPath)
			baseName := strings.TrimSuffix(urlPath, thumbnailSuffix+ext) + ext
			if _, exists := dbURLs[baseName]; exists {
				return nil // 原图在数据库中，保留缩略图
			}
			// 原图已删除，缩略图也要删除
			if removeErr := os.Remove(path); removeErr != nil {
				logger.Warn("[文件清理] 删除缩略图孤文件失败",
					zap.String("path", path),
					zap.Error(removeErr))
				return nil
			}
			deletedCount++
			logger.Info("[文件清理] 已删除缩略图孤文件", zap.String("path", path))
			return nil
		}

		// 如果数据库中不存在该URL，说明是磁盘孤文件
		if _, exists := dbURLs[urlPath]; !exists {
			if removeErr := os.Remove(path); removeErr != nil {
				logger.Warn("[文件清理] 删除磁盘孤文件失败",
					zap.String("path", path),
					zap.Error(removeErr))
				return nil
			}
			deletedCount++
			logger.Info("[文件清理] 已删除磁盘孤文件", zap.String("path", path))
		}
		return nil
	})

	if err != nil {
		logger.Error("[文件清理] 扫描磁盘目录失败", zap.Error(err))
		return
	}

	// 清理空目录
	cleanupEmptyDirs(basePath)

	logger.Info("[文件清理] 磁盘孤文件清理完成",
		zap.Int("deleted", deletedCount))
}

func cleanupEmptyDirs(root string) {
	filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() {
			return nil
		}
		// 尝试删除空目录
		if entries, _ := os.ReadDir(path); len(entries) == 0 {
			if rmErr := os.Remove(path); rmErr == nil {
				logger.Debug("[文件清理] 已删除空目录", zap.String("path", path))
			}
		}
		return nil
	})
}
