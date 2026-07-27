package router

import (
	"github.com/gin-gonic/gin"
	"github.com/KamyangTech/xt-pms/server/internal/controller"
)

func SetupUploadRoutes(r *gin.RouterGroup) {
	uploadController := controller.NewUploadController()

	upload := r.Group("/upload")
	{
		// 上传接口不设独立权限，由前端页面权限控制
		// 例如：用户有创建订单权限，才能在订单页面使用上传功能
		upload.POST("/image", uploadController.UploadImage)
		upload.POST("/files", uploadController.UploadFiles)
		upload.POST("/generate-thumbnails", uploadController.GenerateMissingThumbnails)
	}

	// 文件下载接口（通过 ID 下载，Content-Disposition 返回原始文件名）
	r.GET("/files/:id/download", uploadController.DownloadFile)
}
