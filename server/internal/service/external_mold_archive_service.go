package service

import (
	"errors"
	"time"

	"github.com/KamyangTech/xt-pms/server/internal/dao"
	"github.com/KamyangTech/xt-pms/server/internal/model/entity"
)

type ExternalMoldArchiveService struct {
	archiveDAO *dao.ExternalMoldArchiveDAO
}

func NewExternalMoldArchiveService() *ExternalMoldArchiveService {
	return &ExternalMoldArchiveService{
		archiveDAO: dao.NewExternalMoldArchiveDAO(),
	}
}

// GetByMoldCode 根据模具编号查询档案（用于自动填充）
func (s *ExternalMoldArchiveService) GetByMoldCode(moldCode string) (*entity.ExternalMoldArchive, error) {
	if moldCode == "" {
		return nil, errors.New("模具编号不能为空")
	}
	return s.archiveDAO.GetByMoldCode(moldCode)
}

// CreateOrUpdateArchive 创建或更新档案（首次使用时调用）
func (s *ExternalMoldArchiveService) CreateOrUpdateArchive(moldCode, productName string, customerID *int64, customerName *string, orderID int64) (*entity.ExternalMoldArchive, error) {
	if moldCode == "" || productName == "" {
		return nil, errors.New("模具编号和产品名称不能为空")
	}

	// 先查询是否已存在
	existing, err := s.archiveDAO.GetByMoldCode(moldCode)
	if err != nil {
		return nil, err
	}

	now := time.Now()

	if existing != nil {
		// 已存在，允许使用不同名称（用户已二次确认）
		if existing.ProductName != productName {
			existing.ProductName = productName
		}

		if customerID != nil {
			existing.CustomerID = customerID
		}
		if customerName != nil {
			existing.CustomerName = customerName
		}
		existing.LastOrderID = &orderID
		existing.LastUseDate = &now
		existing.UseCount++

		if err := s.archiveDAO.Update(existing); err != nil {
			return nil, err
		}
		return existing, nil
	} else {
		// 不存在，创建新档案
		archive := &entity.ExternalMoldArchive{
			MoldCode:     moldCode,
			ProductName:  productName,
			CustomerID:   customerID,
			CustomerName: customerName,
			FirstOrderID: &orderID,
			FirstUseDate: &now,
			LastOrderID:  &orderID,
			LastUseDate:  &now,
			UseCount:     1,
			Status:       1,
		}

		if err := s.archiveDAO.Create(archive); err != nil {
			return nil, err
		}
		return archive, nil
	}
}

// ListArchives 查询档案列表
func (s *ExternalMoldArchiveService) ListArchives(moldCode string, customerID *int64, page, pageSize int) ([]entity.ExternalMoldArchive, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	return s.archiveDAO.List(moldCode, customerID, page, pageSize)
}

// UpdateArchive 更新档案信息
func (s *ExternalMoldArchiveService) UpdateArchive(id int64, productName string, remark string) error {
	archive, err := s.archiveDAO.GetByID(id)
	if err != nil {
		return err
	}
	if archive == nil {
		return errors.New("档案不存在")
	}

	archive.ProductName = productName
	archive.Remark = remark

	return s.archiveDAO.Update(archive)
}

// DeleteArchive 删除档案（软删除）
func (s *ExternalMoldArchiveService) DeleteArchive(id int64) error {
	return s.archiveDAO.Delete(id)
}

// GetByID 根据ID查询档案
func (s *ExternalMoldArchiveService) GetByID(id int64) (*entity.ExternalMoldArchive, error) {
	var archive entity.ExternalMoldArchive
	err := dao.GetDB().Where("id = ?", id).First(&archive).Error
	if err != nil {
		return nil, err
	}
	return &archive, nil
}

// GetPendingOrdersByMoldCode 查询该模具编号的未完成订单（排除指定订单ID）
func (s *ExternalMoldArchiveService) GetPendingOrdersByMoldCode(moldCode string, excludeOrderID *int64) ([]map[string]interface{}, error) {
	return s.archiveDAO.GetPendingOrdersByMoldCode(moldCode, excludeOrderID)
}
