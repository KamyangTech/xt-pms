package service

import (
	"errors"
	"time"

	"github.com/KamyangTech/xt-pms/server/internal/common/response"
	"github.com/KamyangTech/xt-pms/server/internal/dao"
	"github.com/KamyangTech/xt-pms/server/internal/model/constant"
	"github.com/KamyangTech/xt-pms/server/internal/model/dto"
	"github.com/KamyangTech/xt-pms/server/internal/model/entity"
)

type MoldService struct {
	moldDAO *dao.MoldDAO
}

func NewMoldService() *MoldService {
	return &MoldService{
		moldDAO: dao.NewMoldDAO(),
	}
}

// CreateMold 创建模具档案
func (s *MoldService) CreateMold(moldDTO dto.MoldDTO) (*entity.Mold, error) {
	existing, _ := s.moldDAO.FindByCode(moldDTO.MoldCode)
	if existing != nil {
		return nil, &response.HttpError{Code: response.BAD_REQUEST, Message: "模具编号已存在"}
	}

	var productionDate *time.Time
	if moldDTO.ProductionDate != "" {
		parsedTime, err := time.ParseInLocation("2006-01-02", moldDTO.ProductionDate, time.Local)
		if err != nil {
			return nil, errors.New("生产日期格式错误")
		}
		productionDate = &parsedTime
	}

	mold := &entity.Mold{
		MoldCode:        moldDTO.MoldCode,
		MoldName:        moldDTO.MoldName,
		CustomerID:      moldDTO.CustomerID,
		ProductionDate:  productionDate,
		LifeCycleStatus: moldDTO.LifeCycleStatus,
		MoldType:        moldDTO.MoldType,
		MoldProcess:     moldDTO.MoldProcess,
		MaterialType:    moldDTO.MaterialType,
		CavityCount:     moldDTO.CavityCount,
		Remark:          moldDTO.Remark,
	}

	// 设置默认值
	if mold.LifeCycleStatus == "" {
		mold.LifeCycleStatus = constant.MoldStatusDesigning
	}
	if mold.MoldType == "" {
		mold.MoldType = "NEW"
	}
	if mold.MaterialType == "" {
		mold.MaterialType = "STEEL"
	}

	err := s.moldDAO.Create(mold)
	if err != nil {
		return nil, err
	}

	return mold, nil
}

// GetMoldByID 根据 ID 获取模具
func (s *MoldService) GetMoldByID(id int64) (*entity.Mold, error) {
	return s.moldDAO.FindByID(id)
}

// GetAllMolds 获取所有模具
func (s *MoldService) GetAllMolds() ([]*entity.Mold, error) {
	return s.moldDAO.FindAll()
}

// GetMoldsByPage 分页获取模具
func (s *MoldService) GetMoldsByPage(page, pageSize int) ([]*entity.Mold, int64, error) {
	result, err := s.moldDAO.FindPage(page, pageSize)
	if err != nil {
		return nil, 0, err
	}
	return result.List.([]*entity.Mold), result.Total, nil
}

// GetMoldsByPageWithFilter 带筛选条件的分页获取模具
func (s *MoldService) GetMoldsByPageWithFilter(page, pageSize int, moldCode, moldName string, customerID *int64, lifeCycleStatus, moldType, moldProcess, materialType, startDate, endDate string) ([]*entity.Mold, int64, error) {
	offset := (page - 1) * pageSize
	return s.moldDAO.FindPageWithFilter(offset, pageSize, moldCode, moldName, customerID, lifeCycleStatus, moldType, moldProcess, materialType, startDate, endDate)
}

// GetPrimaryImagesByMoldIDs 根据模具ID列表批量查询主图
func (s *MoldService) GetPrimaryImagesByMoldIDs(moldIDs []int64) (map[int64]string, error) {
	return s.moldDAO.GetPrimaryImagesByMoldIDs(moldIDs)
}

// UpdateMold 更新模具
func (s *MoldService) UpdateMold(moldDTO dto.MoldUpdateDTO) error {
	mold, err := s.moldDAO.FindByID(moldDTO.ID)
	if err != nil {
		return errors.New("模具不存在")
	}

	// 构造更新数据，只更新非空/非零值的字段
	updateData := make(map[string]interface{})

	if moldDTO.MoldCode != "" {
		updateData["mold_code"] = moldDTO.MoldCode
	}
	if moldDTO.MoldName != "" {
		updateData["mold_name"] = moldDTO.MoldName
	}
	if moldDTO.CustomerID != 0 {
		updateData["customer_id"] = moldDTO.CustomerID
	}

	if moldDTO.ProductionDate != "" {
		parsedTime, err := time.ParseInLocation("2006-01-02", moldDTO.ProductionDate, time.Local)
		if err == nil {
			updateData["production_date"] = parsedTime
		}
	}

	// 只更新非空值的状态字段
	if moldDTO.LifeCycleStatus != "" {
		updateData["life_cycle_status"] = moldDTO.LifeCycleStatus
	}
	if moldDTO.MoldType != "" {
		updateData["mold_type"] = moldDTO.MoldType
	}
	if moldDTO.MoldProcess != "" {
		updateData["mold_process"] = moldDTO.MoldProcess
	}
	if moldDTO.MaterialType != "" {
		updateData["material_type"] = moldDTO.MaterialType
	}
	if moldDTO.CavityCount > 0 {
		updateData["cavity_count"] = moldDTO.CavityCount
	}
	if moldDTO.Remark != "" {
		updateData["remark"] = moldDTO.Remark
	}

	// 如果没有需要更新的字段，直接返回
	if len(updateData) == 0 {
		return nil
	}

	// 使用 Updates 只更新指定字段，避免影响关联的图片记录
	return s.moldDAO.UpdateWithMap(mold.ID, updateData)
}

// DeleteMold 删除模具（已禁用物理删除：结档请使用生命周期「报废 SCRAPPED」）
func (s *MoldService) DeleteMold(id int64, _ bool) error {
	_, err := s.moldDAO.FindByID(id)
	if err != nil {
		return errors.New("模具不存在")
	}
	return errors.New("已取消物理删除模具：请将生命周期调整为报废（SCRAPPED）以结档，历史单据与追溯记录继续保留")
}

// UpdateMoldStatus 更新模具状态
func (s *MoldService) UpdateMoldStatus(id int64, status int8) error {
	return s.moldDAO.UpdateStatus(id, status)
}

// GetMoldStatistics 获取模具统计信息（生命周期分布由 DB 分组查询，替代数据库视图 v_mold_status_stat）
func (s *MoldService) GetMoldStatistics() (map[string]int64, error) {
	statistics := make(map[string]int64)

	total, err := s.moldDAO.CountAll()
	if err != nil {
		return nil, err
	}
	statistics["total"] = total

	statusMap, err := s.moldDAO.CountByStatusBatch()
	if err != nil {
		return nil, err
	}

	// 全量生命周期分布（键为 life_cycle_status 字典值）
	for k, v := range statusMap {
		statistics[k] = v
	}

	// 兼容旧前端字段名
	statistics["processing"] = statusMap[constant.MoldStatusProcessing]
	statistics["completed"] = statusMap[constant.MoldStatusInStock] +
		statusMap[constant.MoldStatusDelivered]
	statistics["repairing"] = statusMap[constant.MoldStatusRepairing]

	return statistics, nil
}
