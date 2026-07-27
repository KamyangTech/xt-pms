package service

import (
	"errors"
	"time"

	"github.com/KamyangTech/xt-pms/server/internal/dao"
	"github.com/KamyangTech/xt-pms/server/internal/model/entity"
	"gorm.io/gorm"
)

type MoldUsageService struct {
	usageDAO *dao.MoldUsageRecordDAO
	moldDAO  *dao.MoldDAO
}

func NewMoldUsageService() *MoldUsageService {
	return &MoldUsageService{
		usageDAO: dao.NewMoldUsageRecordDAO(),
		moldDAO:  dao.NewMoldDAO(),
	}
}

// CreateUsageRecord 创建模具使用记录
func (s *MoldUsageService) CreateUsageRecord(tx *gorm.DB, moldID, orderID int64, shots, okQty, badQty int) error {
	if tx == nil {
		return errors.New("数据库事务不能为空")
	}

	mold, err := s.moldDAO.FindByIDWithTx(tx, moldID)
	if err != nil {
		return errors.New("模具不存在")
	}

	// 检查是否已存在该订单的使用记录（兼容旧逻辑）
	existing, _ := s.usageDAO.FindByMoldAndOrder(moldID, orderID)
	if existing != nil {
		return errors.New("该生产订单的模具使用记录已存在")
	}

	now := time.Now()
	usageRecord := &entity.MoldUsageRecord{
		MoldID:            moldID,
		ProductionOrderID: orderID,
		UsageDate:         &now,
		TotalShots:        shots,
		OkQty:             okQty,
		BadQty:            badQty,
		StartDate:         &now,
		EndDate:           &now,
	}

	if err := tx.Create(usageRecord).Error; err != nil {
		return errors.New("创建模具使用记录失败: " + err.Error())
	}

	newTotalShots := mold.TotalShots + shots
	updateData := map[string]interface{}{
		"total_shots":     newTotalShots,
		"last_usage_date": now,
	}

	if err := s.moldDAO.UpdateWithMapWithTx(tx, moldID, updateData); err != nil {
		return errors.New("更新模具模次失败: " + err.Error())
	}

	return nil
}

// CreateUsageRecordForItem 为生产订单明细创建模具使用记录（支持按产品细分）
func (s *MoldUsageService) CreateUsageRecordForItem(tx *gorm.DB, moldID, orderID, itemID int64, productID *int64, productName *string, shots, okQty, badQty int) error {
	if tx == nil {
		return errors.New("数据库事务不能为空")
	}

	mold, err := s.moldDAO.FindByIDWithTx(tx, moldID)
	if err != nil {
		return errors.New("模具不存在")
	}

	existing, _ := s.usageDAO.FindByMoldAndItemWithTx(tx, moldID, itemID)
	if existing != nil {
		return errors.New("该生产订单明细的模具使用记录已存在")
	}

	now := time.Now()
	usageRecord := &entity.MoldUsageRecord{
		MoldID:                moldID,
		ProductionOrderID:     orderID,
		ProductionOrderItemID: &itemID,
		ProductID:             productID,
		ProductName:           productName,
		UsageDate:             &now,
		TotalShots:            shots,
		OkQty:                 okQty,
		BadQty:                badQty,
		StartDate:             &now,
		EndDate:               &now,
	}

	if err := tx.Create(usageRecord).Error; err != nil {
		return errors.New("创建模具使用记录失败: " + err.Error())
	}

	newTotalShots := mold.TotalShots + shots
	updateData := map[string]interface{}{
		"total_shots":     newTotalShots,
		"last_usage_date": now,
	}

	if err := s.moldDAO.UpdateWithMapWithTx(tx, moldID, updateData); err != nil {
		return errors.New("更新模具模次失败: " + err.Error())
	}

	return nil
}

// GetUsageHistory 获取模具使用历史
func (s *MoldUsageService) GetUsageHistory(moldID int64, page, pageSize int, moldCode string) ([]entity.MoldUsageRecord, int64, error) {
	return s.usageDAO.GetListByMoldID(moldID, page, pageSize, moldCode)
}

// UpdateMoldShots 更新模具累计模次
func (s *MoldUsageService) UpdateMoldShots(moldID int64, additionalShots int) error {
	mold, err := s.moldDAO.FindByID(moldID)
	if err != nil {
		return errors.New("模具不存在")
	}

	newTotalShots := mold.TotalShots + additionalShots
	now := time.Now()

	updateData := map[string]interface{}{
		"total_shots":     newTotalShots,
		"last_usage_date": now,
	}

	if err := s.moldDAO.UpdateWithMap(moldID, updateData); err != nil {
		return errors.New("更新模具模次失败: " + err.Error())
	}

	return nil
}

// UpdateMoldShotsWithTx 在事务内更新模具累计模次
func (s *MoldUsageService) UpdateMoldShotsWithTx(tx *gorm.DB, moldID int64, additionalShots int) error {
	mold, err := s.moldDAO.FindByIDWithTx(tx, moldID)
	if err != nil {
		return errors.New("模具不存在")
	}

	newTotalShots := mold.TotalShots + additionalShots
	now := time.Now()

	updateData := map[string]interface{}{
		"total_shots":     newTotalShots,
		"last_usage_date": now,
	}

	if err := s.moldDAO.UpdateWithMapWithTx(tx, moldID, updateData); err != nil {
		return errors.New("更新模具模次失败: " + err.Error())
	}

	return nil
}

// GetUsageStats 获取模具使用统计
func (s *MoldUsageService) GetUsageStats(moldID int64) (map[string]interface{}, error) {
	// 如果 moldID 为 0，查询所有模具的使用统计
	if moldID == 0 {
		records, _, err := s.usageDAO.GetListByMoldID(0, 1, 10000, "")
		if err != nil {
			return nil, errors.New("获取使用记录失败: " + err.Error())
		}

		totalOrders := len(records)
		totalShots := 0
		totalOkQty := 0
		totalBadQty := 0

		for _, record := range records {
			totalShots += record.TotalShots
			totalOkQty += record.OkQty
			totalBadQty += record.BadQty
		}

		yieldRate := 0.0
		if totalOkQty+totalBadQty > 0 {
			yieldRate = float64(totalOkQty) / float64(totalOkQty+totalBadQty) * 100
		}

		stats := map[string]interface{}{
			"total_orders":  totalOrders,
			"total_shots":   totalShots,
			"total_ok_qty":  totalOkQty,
			"total_bad_qty": totalBadQty,
			"yield_rate":    yieldRate,
		}

		return stats, nil
	}

	records, _, err := s.usageDAO.GetListByMoldID(moldID, 1, 10000, "")
	if err != nil {
		return nil, errors.New("获取使用记录失败: " + err.Error())
	}

	totalOrders := len(records)
	totalShots := 0
	totalOkQty := 0
	totalBadQty := 0

	for _, record := range records {
		totalShots += record.TotalShots
		totalOkQty += record.OkQty
		totalBadQty += record.BadQty
	}

	yieldRate := 0.0
	if totalOkQty+totalBadQty > 0 {
		yieldRate = float64(totalOkQty) / float64(totalOkQty+totalBadQty) * 100
	}

	stats := map[string]interface{}{
		"total_orders":  totalOrders,
		"total_shots":   totalShots,
		"total_ok_qty":  totalOkQty,
		"total_bad_qty": totalBadQty,
		"yield_rate":    yieldRate,
	}

	return stats, nil
}
