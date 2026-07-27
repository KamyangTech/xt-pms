package controller

import (
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/KamyangTech/xt-pms/server/config"
	"github.com/KamyangTech/xt-pms/server/internal/common/middleware"
	"github.com/KamyangTech/xt-pms/server/internal/common/response"
	"github.com/KamyangTech/xt-pms/server/internal/dao"
	"github.com/KamyangTech/xt-pms/server/internal/model/entity"
	"github.com/KamyangTech/xt-pms/server/internal/utils"
	"github.com/KamyangTech/xt-pms/server/pkg/logger"
	"go.uber.org/zap"
)

type UploadController struct {
	fileDAO *dao.FileDAO
}

func NewUploadController() *UploadController {
	return &UploadController{
		fileDAO: dao.NewFileDAO(),
	}
}

type UploadFileResponse struct {
	ID       string `json:"id"`
	URL      string `json:"url"`
	FileName string `json:"file_name"`
	FileSize int64  `json:"file_size"`
	FileType string `json:"file_type"`
}

func getValidExts() map[string]bool {
	exts := make(map[string]bool)
	types := config.GlobalConfig.Upload.AllowedFileTypes
	if len(types) == 0 {
		types = config.GlobalConfig.Upload.AllowedImageTypes
	}
	for _, ext := range types {
		exts[ext] = true
	}
	return exts
}

func getAllowedExts() string {
	types := config.GlobalConfig.Upload.AllowedFileTypes
	if len(types) == 0 {
		types = config.GlobalConfig.Upload.AllowedImageTypes
	}
	return strings.Join(types, ", ")
}

// validateFileMIME 验证文件内容的真实 MIME 类型是否与扩展名匹配
// 读取文件头 512 字节检测，防止恶意文件伪装扩展名上传
func validateFileMIME(file *multipart.FileHeader, ext string) error {
	f, err := file.Open()
	if err != nil {
		return fmt.Errorf("无法读取文件内容: %w", err)
	}
	defer f.Close()

	buf := make([]byte, 512)
	n, err := f.Read(buf)
	if err != nil && err != io.EOF {
		return fmt.Errorf("读取文件内容失败: %w", err)
	}

	mimeType := http.DetectContentType(buf[:n])

	// 扩展名 → 期望的 MIME 类型映射
	expectedMIME := map[string]string{
		".jpg":  "image/jpeg",
		".jpeg": "image/jpeg",
		".png":  "image/png",
		".gif":  "image/gif",
		".webp": "image/webp",
		".pdf":  "application/pdf",
		".doc":  "application/msword",
		".docx": "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
		".xls":  "application/vnd.ms-excel",
		".xlsx": "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
		".zip":  "application/zip",
		".rar":  "application/vnd.rar",
		".7z":   "application/x-7z-compressed",
		".tar":  "application/x-tar",
		".gz":   "application/gzip",
	}

	expect, ok := expectedMIME[ext]
	if !ok {
		return nil // 未知扩展名，由上层配置决定
	}

	// 部分 MIME 类型可能带附加参数（如 charset），做前缀匹配
	if !strings.HasPrefix(mimeType, expect) {
		logger.Warn("文件 MIME 类型不匹配，拒绝上传",
			zap.String("filename", file.Filename),
			zap.String("extension", ext),
			zap.String("expected_mime", expect),
			zap.String("actual_mime", mimeType),
		)
		return fmt.Errorf("文件内容类型不匹配，检测到 %s，期望 %s", mimeType, expect)
	}

	return nil
}

// UploadImage 通用单图片上传（向后兼容，hash 分片存储 + DB 记录）
func (ctrl *UploadController) UploadImage(c *gin.Context) {
	file, err := c.FormFile("image")
	if err != nil {
		response.Error(c, http.StatusBadRequest, "获取上传文件失败")
		return
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	validExts := getValidExts()
	if !validExts[ext] {
		response.Error(c, http.StatusBadRequest, fmt.Sprintf("不支持的文件格式，仅支持 %v", getAllowedExts()))
		return
	}
	if file.Size > config.GlobalConfig.Upload.MaxSize {
		maxSizeMB := config.GlobalConfig.Upload.MaxSize / 1024 / 1024
		response.Error(c, http.StatusBadRequest, fmt.Sprintf("文件大小不能超过 %dMB", maxSizeMB))
		return
	}

	// 验证文件内容真实类型（防止伪装扩展名上传恶意文件）
	if err := validateFileMIME(file, ext); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	fileID := uuid.New().String()

	// hash 分片存储
	_, filePath, urlPath := utils.GenerateStoragePath(fileID, ext)
	if err := utils.SaveUploadedFile(file, filePath); err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	// 为图片文件生成缩略图（失败不阻断上传流程，仅记录日志）
	utils.GenerateThumbnail(filePath, urlPath)

	// 记录 DB
	fileRecord := &entity.File{
		ID:       fileID,
		URL:      urlPath,
		FileName: file.Filename,
		FileSize: file.Size,
		FileType: ext,
	}
	if userID := middleware.GetUserID(c); userID > 0 {
		fileRecord.CreatedBy = &userID
	}
	if err := ctrl.fileDAO.Create(fileRecord); err != nil {
		response.Error(c, http.StatusBadRequest, "保存文件记录失败")
		return
	}

	response.Success(c, &UploadFileResponse{
		ID:       fileID,
		URL:      urlPath,
		FileName: file.Filename,
		FileSize: file.Size,
		FileType: ext,
	})
}

// UploadFiles 通用多文件上传接口（hash 分片 + DB 记录）
// 请求参数：
//   - files[]: 文件列表（必填）
//   - category: 业务类型（可选，如 expense/receivable/mold/product，会写入 DB）
// 存储路径：uploads/{hash[:2]}/{hash[2:4]}/{uuid}_{safe_filename}
func (ctrl *UploadController) UploadFiles(c *gin.Context) {
	form, err := c.MultipartForm()
	if err != nil {
		response.Error(c, http.StatusBadRequest, "获取上传文件失败: "+err.Error())
		return
	}

	files := form.File["files[]"]
	if len(files) == 0 {
		singleFile, err := c.FormFile("file")
		if err != nil {
			response.Error(c, http.StatusBadRequest, "请选择要上传的文件")
			return
		}
		files = []*multipart.FileHeader{singleFile}
	}

	category := c.DefaultPostForm("category", "")
	validExts := getValidExts()
	maxSize := config.GlobalConfig.Upload.MaxSize
	maxSizeMB := maxSize / 1024 / 1024
	userID := middleware.GetUserID(c)

	var uploadedFiles []UploadFileResponse

	for _, file := range files {
		ext := strings.ToLower(filepath.Ext(file.Filename))
		if !validExts[ext] {
			response.Error(c, http.StatusBadRequest, fmt.Sprintf("文件 %s 格式不支持，仅支持 %v", file.Filename, getAllowedExts()))
			return
		}
		if file.Size > maxSize {
			response.Error(c, http.StatusBadRequest, fmt.Sprintf("文件 %s 大小不能超过 %dMB", file.Filename, maxSizeMB))
			return
		}

		if err := validateFileMIME(file, ext); err != nil {
			response.Error(c, http.StatusBadRequest, fmt.Sprintf("文件 %s: %s", file.Filename, err.Error()))
			return
		}

		fileID := uuid.New().String()

		_, filePath, urlPath := utils.GenerateStoragePath(fileID, ext)
		if err := utils.SaveUploadedFile(file, filePath); err != nil {
			response.Error(c, http.StatusInternalServerError, err.Error())
			return
		}

		// 为图片文件生成缩略图（失败不阻断上传流程）
		utils.GenerateThumbnail(filePath, urlPath)

		// 记录 DB
		fileRecord := &entity.File{
			ID:           fileID,
			URL:          urlPath,
			FileName:     file.Filename,
			FileSize:     file.Size,
			FileType:     ext,
			BusinessType: category,
		}
		if userID > 0 {
			fileRecord.CreatedBy = &userID
		}
		if err := ctrl.fileDAO.Create(fileRecord); err != nil {
			response.Error(c, http.StatusBadRequest, "保存文件记录失败")
			return
		}

		uploadedFiles = append(uploadedFiles, UploadFileResponse{
			ID:       fileID,
			URL:      urlPath,
			FileName: file.Filename,
			FileSize: file.Size,
			FileType: ext,
		})
	}

	response.Success(c, uploadedFiles)
}

// DownloadFile 文件下载接口
// 通过文件 ID 查找记录，以 Content-Disposition: attachment 返回文件，
// 下载时使用 DB 中保存的原始文件名，而非磁盘上的 UUID 文件名
// GET /api/files/:id/download
func (ctrl *UploadController) DownloadFile(c *gin.Context) {
	fileID := c.Param("id")
	if fileID == "" {
		response.Error(c, http.StatusBadRequest, "缺少文件ID")
		return
	}

	fileRecord, err := ctrl.fileDAO.FindByID(fileID)
	if err != nil {
		response.Error(c, http.StatusNotFound, "文件不存在")
		return
	}

	// 从 url 推导磁盘路径
	diskPath := filepath.Join(config.GlobalConfig.Upload.BasePath, strings.TrimPrefix(fileRecord.URL, "/uploads/"))

	// 防御路径遍历：确保最终路径仍在 uploads 基准目录内
	absDiskPath, err := filepath.Abs(diskPath)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "路径解析失败")
		return
	}
	absBasePath, _ := filepath.Abs(config.GlobalConfig.Upload.BasePath)
	if !strings.HasPrefix(absDiskPath, absBasePath+string(filepath.Separator)) {
		logger.Error("文件下载路径遍历检测", zap.String("url", fileRecord.URL), zap.String("resolved", absDiskPath))
		response.Error(c, http.StatusForbidden, "文件路径非法")
		return
	}

	if _, err := os.Stat(absDiskPath); os.IsNotExist(err) {
		response.Error(c, http.StatusNotFound, "文件已被删除")
		return
	}

	// FileAttachment 自动处理 Content-Disposition 和 Content-Type
	c.FileAttachment(absDiskPath, fileRecord.FileName)
}

// GenerateMissingThumbnails 批量扫描现有图片，为缺失缩略图的文件生成缩略图
func (ctrl *UploadController) GenerateMissingThumbnails(c *gin.Context) {
	var urls []string
	if err := dao.GetDB().Model(&entity.ProductionOrderItem{}).
		Where("product_image_url IS NOT NULL AND product_image_url != ''").
		Distinct("product_image_url").
		Pluck("product_image_url", &urls).Error; err != nil {
		logger.Error("查询产品图片URL失败", zap.Error(err))
		response.Error(c, http.StatusInternalServerError, "查询失败")
		return
	}

	basePath := config.GlobalConfig.Upload.BasePath
	result := utils.GenerateMissingThumbnails(basePath, urls, nil)

	response.Success(c, result)
}
