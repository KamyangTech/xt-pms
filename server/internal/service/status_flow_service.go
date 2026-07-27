package service

import (
	"sync"

	"github.com/KamyangTech/xt-pms/server/internal/dao"
	"github.com/KamyangTech/xt-pms/server/internal/model/constant"
	"github.com/KamyangTech/xt-pms/server/internal/model/entity"
)

type StatusFlowService struct {
	configDAO *dao.StatusFlowConfigDAO
	cache     map[string][]*entity.StatusFlowConfig // key: "work_order:{status}" / "mold_order:{status}" / "mold:{status}"
	cacheMu   sync.RWMutex
}

func NewStatusFlowService() *StatusFlowService {
	s := &StatusFlowService{
		configDAO: dao.NewStatusFlowConfigDAO(),
		cache:     make(map[string][]*entity.StatusFlowConfig),
	}
	// 启动时加载缓存
	s.refreshCache()
	return s
}

// GetConfigsByWorkOrderStatus 根据工单状态获取配置
func (s *StatusFlowService) GetConfigsByWorkOrderStatus(workOrderStatus string) []*entity.StatusFlowConfig {
	key := "work_order:" + workOrderStatus
	s.cacheMu.RLock()
	defer s.cacheMu.RUnlock()
	return s.cache[key]
}

// GetConfigsByMoldOrderStatus 根据订单状态获取配置
func (s *StatusFlowService) GetConfigsByMoldOrderStatus(moldOrderStatus string) []*entity.StatusFlowConfig {
	key := "mold_order:" + moldOrderStatus
	s.cacheMu.RLock()
	defer s.cacheMu.RUnlock()
	return s.cache[key]
}

// GetConfigsByMoldStatus 根据模具状态获取配置
func (s *StatusFlowService) GetConfigsByMoldStatus(moldStatus string) []*entity.StatusFlowConfig {
	key := "mold:" + moldStatus
	s.cacheMu.RLock()
	defer s.cacheMu.RUnlock()
	return s.cache[key]
}

// GetMoldOrderStatusByWorkOrder 根据工单状态获取对应的订单状态（取优先级最高的）
func (s *StatusFlowService) GetMoldOrderStatusByWorkOrder(workOrderStatus string) *string {
	configs := s.GetConfigsByWorkOrderStatus(workOrderStatus)
	for _, config := range configs {
		if config.MoldOrderStatus != nil {
			return config.MoldOrderStatus
		}
	}
	return nil
}

// GetMoldStatusByWorkOrder 根据工单状态获取对应的模具状态
func (s *StatusFlowService) GetMoldStatusByWorkOrder(workOrderStatus string) *string {
	configs := s.GetConfigsByWorkOrderStatus(workOrderStatus)
	for _, config := range configs {
		if config.MoldStatus != nil {
			return config.MoldStatus
		}
	}
	return nil
}

// GetWorkOrderStatusByMoldOrder 根据订单状态获取对应的工单状态
func (s *StatusFlowService) GetWorkOrderStatusByMoldOrder(moldOrderStatus string) *string {
	configs := s.GetConfigsByMoldOrderStatus(moldOrderStatus)
	for _, config := range configs {
		if config.WorkOrderStatus != nil {
			return config.WorkOrderStatus
		}
	}
	return nil
}

// GetMoldStatusByMoldOrder 根据订单状态获取对应的模具状态
func (s *StatusFlowService) GetMoldStatusByMoldOrder(moldOrderStatus string) *string {
	configs := s.GetConfigsByMoldOrderStatus(moldOrderStatus)
	for _, config := range configs {
		if config.MoldStatus != nil {
			return config.MoldStatus
		}
	}
	return nil
}

// GetWorkOrderStatusByMold 根据模具状态获取对应的工单状态
func (s *StatusFlowService) GetWorkOrderStatusByMold(moldStatus string) *string {
	configs := s.GetConfigsByMoldStatus(moldStatus)
	for _, config := range configs {
		if config.WorkOrderStatus != nil {
			return config.WorkOrderStatus
		}
	}
	return nil
}

// GetMoldOrderStatusByMold 根据模具状态获取对应的订单状态
func (s *StatusFlowService) GetMoldOrderStatusByMold(moldStatus string) *string {
	configs := s.GetConfigsByMoldStatus(moldStatus)
	for _, config := range configs {
		if config.MoldOrderStatus != nil {
			return config.MoldOrderStatus
		}
	}
	return nil
}

// RefreshCache 刷新缓存
func (s *StatusFlowService) RefreshCache() {
	s.refreshCache()
}

func (s *StatusFlowService) refreshCache() {
	configs, err := s.configDAO.FindAllActive()
	if err != nil {
		return
	}

	s.cacheMu.Lock()
	defer s.cacheMu.Unlock()

	s.cache = make(map[string][]*entity.StatusFlowConfig)

	for _, config := range configs {
		// 按工单状态索引
		if config.WorkOrderStatus != nil {
			key := "work_order:" + *config.WorkOrderStatus
			s.cache[key] = append(s.cache[key], config)
		}

		// 按订单状态索引
		if config.MoldOrderStatus != nil {
			key := "mold_order:" + *config.MoldOrderStatus
			s.cache[key] = append(s.cache[key], config)
		}

		// 按模具状态索引
		if config.MoldStatus != nil {
			key := "mold:" + *config.MoldStatus
			s.cache[key] = append(s.cache[key], config)
		}
	}
}

// GetByID 根据ID获取配置
func (s *StatusFlowService) GetByID(id int64) (*entity.StatusFlowConfig, error) {
	return s.configDAO.FindByID(id)
}

// GetAll 获取所有配置
func (s *StatusFlowService) GetAll() ([]*entity.StatusFlowConfig, error) {
	return s.configDAO.FindAllActive()
}

// GetDeliveredWorkOrderStatus 获取"已交付"状态的工单状态值
func (s *StatusFlowService) GetDeliveredWorkOrderStatus() *string {
	return s.getStatusByDictValue(constant.WorkOrderStatusDelivered, "work_order")
}

// GetDeliveredOrderStatus 获取"已交付"状态的订单状态值
func (s *StatusFlowService) GetDeliveredOrderStatus() *string {
	return s.getStatusByDictValue(constant.MoldOrderStatusDelivered, "mold_order")
}

// GetCancelledOrderStatus 获取"已取消"状态的订单状态值
func (s *StatusFlowService) GetCancelledOrderStatus() *string {
	return s.getStatusByDictValue(constant.MoldOrderStatusCancelled, "mold_order")
}

// GetTestedWorkOrderStatus 获取"已试模"状态的工单状态值
func (s *StatusFlowService) GetTestedWorkOrderStatus() *string {
	return s.getStatusByDictValue(constant.WorkOrderStatusTested, "work_order")
}

// GetNotStartedWorkOrderStatus 获取"未开始"状态的工单状态值
func (s *StatusFlowService) GetNotStartedWorkOrderStatus() *string {
	return s.getStatusByDictValue(constant.WorkOrderStatusNotStarted, "work_order")
}

// GetProcessingWorkOrderStatus 获取"加工中"状态的工单状态值
func (s *StatusFlowService) GetProcessingWorkOrderStatus() *string {
	return s.getStatusByDictValue(constant.WorkOrderStatusProcessing, "work_order")
}

// GetCompletedWorkOrderStatus 获取"已完成"状态的工单状态值
func (s *StatusFlowService) GetCompletedWorkOrderStatus() *string {
	return s.getStatusByDictValue(constant.WorkOrderStatusCompleted, "work_order")
}

// getStatusByDictValue 根据字典值和类型获取状态值（使用缓存优化）
func (s *StatusFlowService) getStatusByDictValue(dictValue string, statusType string) *string {
	// 遍历缓存中的所有配置，查找匹配的字典值
	s.cacheMu.RLock()
	defer s.cacheMu.RUnlock()

	for _, configs := range s.cache {
		for _, config := range configs {
			var configValue *string
			switch statusType {
			case "work_order":
				configValue = config.WorkOrderStatus
			case "mold_order":
				configValue = config.MoldOrderStatus
			case "mold":
				configValue = config.MoldStatus
			}

			if configValue != nil && *configValue == dictValue {
				return configValue
			}
		}
	}
	return nil
}

// GetWorkOrderStatusByProgress 根据工序进度获取对应的工单状态
// 通过语义标签映射：progress=0 -> not_started, progress<100 -> processing, progress=100 -> completed
func (s *StatusFlowService) GetWorkOrderStatusByProgress(progress int) *string {
	if progress == 0 {
		return s.GetNotStartedWorkOrderStatus()
	} else if progress < 100 {
		return s.GetProcessingWorkOrderStatus()
	} else {
		return s.GetCompletedWorkOrderStatus()
	}
}

// GetStatusOrder 获取状态的顺序号（用于状态比较）
// 如果找不到配置，返回一个很大的数字（999）表示未知状态
func (s *StatusFlowService) GetStatusOrder(statusType string, statusValue string) int {
	var configValue *string
	var configs []*entity.StatusFlowConfig

	switch statusType {
	case "work_order":
		configs = s.GetConfigsByWorkOrderStatus(statusValue)
	case "mold_order":
		configs = s.GetConfigsByMoldOrderStatus(statusValue)
	case "mold":
		configs = s.GetConfigsByMoldStatus(statusValue)
	default:
		return 999
	}

	for _, config := range configs {
		switch statusType {
		case "work_order":
			configValue = config.WorkOrderStatus
		case "mold_order":
			configValue = config.MoldOrderStatus
		case "mold":
			configValue = config.MoldStatus
		}

		if configValue != nil && *configValue == statusValue {
			return config.StatusOrder
		}
	}

	return 999 // 未知状态返回最大值
}
