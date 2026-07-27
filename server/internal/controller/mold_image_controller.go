package controller

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/KamyangTech/xt-pms/server/internal/common/response"
	"github.com/KamyangTech/xt-pms/server/internal/service"
)

// MoldImageController 模具图片控制器（管理关联，不含上传）
// 图片上传统一走 UploadController，该控制器仅处理模具图片的查询/删除/主图设置
type MoldImageController struct {
	imageService *service.MoldImageService
}

func NewMoldImageController() *MoldImageController {
	return &MoldImageController{
		imageService: service.NewMoldImageService(),
	}
}

// GetMoldImages 获取模具的所有图片
func (ctrl *MoldImageController) GetMoldImages(c *gin.Context) {
	moldIDStr := c.Param("id")
	moldID, err := strconv.ParseInt(moldIDStr, 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "模具 ID 格式错误")
		return
	}

	images, err := ctrl.imageService.GetImagesByMoldID(moldID)
	if err != nil {
		response.Error(c, http.StatusBadRequest, fmt.Sprintf("获取图片列表失败: %v", err))
		return
	}

	response.Success(c, images)
}

// AttachFiles 将已上传的文件关联到模具（设置 business_id）
func (ctrl *MoldImageController) AttachFiles(c *gin.Context) {
	moldIDStr := c.Param("id")
	moldID, err := strconv.ParseInt(moldIDStr, 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "模具 ID 格式错误")
		return
	}

	var req struct {
		FileIDs []string `json:"file_ids"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "参数格式错误")
		return
	}

	if err := ctrl.imageService.AttachFiles(moldID, req.FileIDs); err != nil {
		response.Error(c, http.StatusBadRequest, fmt.Sprintf("关联文件失败: %v", err))
		return
	}

	response.Success(c, nil)
}

// DeleteMoldImage 删除模具图片
func (ctrl *MoldImageController) DeleteMoldImage(c *gin.Context) {
	imageID := c.Param("imageId")

	if err := ctrl.imageService.DeleteImage(imageID); err != nil {
		response.Error(c, http.StatusBadRequest, fmt.Sprintf("删除图片失败: %v", err))
		return
	}

	response.Success(c, nil)
}

// SetPrimaryImage 设置主图
func (ctrl *MoldImageController) SetPrimaryImage(c *gin.Context) {
	moldIDStr := c.Param("id")
	moldID, err := strconv.ParseInt(moldIDStr, 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "模具 ID 格式错误")
		return
	}

	fileID := c.Param("imageId")

	if err := ctrl.imageService.SetPrimaryImage(moldID, fileID); err != nil {
		response.Error(c, http.StatusBadRequest, fmt.Sprintf("设置主图失败: %v", err))
		return
	}

	response.Success(c, nil)
}
