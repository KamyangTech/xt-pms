package service

import (
	"fmt"
	"os"

	"github.com/KamyangTech/xt-pms/server/internal/dao"
	"github.com/KamyangTech/xt-pms/server/internal/model/entity"
)

// MoldImageService 模具图片服务（基于统一 files 表）
// 注意：图片上传走统一的 UploadController，该服务仅处理模具图片关联管理
type MoldImageService struct {
	fileDAO *dao.FileDAO
}

func NewMoldImageService() *MoldImageService {
	return &MoldImageService{
		fileDAO: dao.NewFileDAO(),
	}
}

// GetImagesByMoldID 获取模具的所有关联图片
func (s *MoldImageService) GetImagesByMoldID(moldID int64) ([]entity.File, error) {
	return s.fileDAO.FindByBusiness("mold", moldID)
}

// GetPrimaryImage 获取模具的主图
func (s *MoldImageService) GetPrimaryImage(moldID int64) (*entity.File, error) {
	return s.fileDAO.FindPrimary("mold", moldID)
}

// SetPrimaryImage 设置模具主图
func (s *MoldImageService) SetPrimaryImage(moldID int64, fileID string) error {
	return s.fileDAO.SetPrimary("mold", moldID, fileID)
}

// DeleteImage 删除模具图片（同时清理物理文件）
func (s *MoldImageService) DeleteImage(fileID string) error {
	file, err := s.fileDAO.FindByID(fileID)
	if err != nil {
		return fmt.Errorf("文件不存在: %v", err)
	}

	if err := s.fileDAO.DeleteByID(fileID); err != nil {
		return fmt.Errorf("删除文件记录失败: %v", err)
	}

	// 清理物理文件
	if err := os.Remove(file.URL); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("删除物理文件失败: %v", err)
	}

	return nil
}

// AttachFiles 将已上传的文件关联到模具（设置 business_type + business_id + sort 序号）
func (s *MoldImageService) AttachFiles(moldID int64, fileIDs []string) error {
	if len(fileIDs) == 0 {
		return nil
	}

	// 先查找该模具已有的图片数量，用于计算排序
	existing, err := s.fileDAO.FindByBusiness("mold", moldID)
	if err != nil {
		return fmt.Errorf("查询已有图片失败: %v", err)
	}
	baseSort := len(existing)

	// 使用 map 更新，只更新指定字段，避免 GORM Save 将零值字段写入
	for i, fileID := range fileIDs {
		if err := s.fileDAO.UpdateByMap(fileID, map[string]interface{}{
			"business_type": "mold",
			"business_id":   moldID,
			"sort":          baseSort + i,
		}); err != nil {
			return fmt.Errorf("关联文件 %s 失败: %v", fileID, err)
		}
	}

	return nil
}

// DeleteImagesByMoldID 删除模具的所有图片
func (s *MoldImageService) DeleteImagesByMoldID(moldID int64) error {
	files, err := s.fileDAO.FindByBusiness("mold", moldID)
	if err != nil {
		return err
	}

	for _, file := range files {
		os.Remove(file.URL)
	}

	return s.fileDAO.DeleteByBusiness("mold", moldID)
}
