package dao

import (
	"github.com/KamyangTech/xt-pms/server/internal/model/entity"
	"gorm.io/gorm"
)

// WorkOrderProcessDAO 工单工序关联DAO
type WorkOrderProcessDAO struct {
	db *gorm.DB
}

func NewWorkOrderProcessDAO() *WorkOrderProcessDAO {
	return &WorkOrderProcessDAO{db: GetDB()}
}

func (d *WorkOrderProcessDAO) Create(wop *entity.WorkOrderProcess) error {
	return d.db.Create(wop).Error
}

func (d *WorkOrderProcessDAO) FindByID(id int64) (*entity.WorkOrderProcess, error) {
	var wop entity.WorkOrderProcess
	err := d.db.First(&wop, id).Error
	return &wop, err
}

func (d *WorkOrderProcessDAO) FindByWorkOrderID(workOrderID int64) ([]*entity.WorkOrderProcess, error) {
	var wops []*entity.WorkOrderProcess
	err := d.db.Model(&entity.WorkOrderProcess{}).
		Joins("LEFT JOIN process_template pt ON work_order_process.process_template_id = pt.id").
		Where("work_order_process.work_order_id = ?", workOrderID).
		Order("pt.sort ASC").
		Find(&wops).Error

	if err != nil {
		return nil, err
	}

	// 批量加载 ProcessTemplate 关联数据（避免N+1查询）
	templateIDs := make([]int64, 0)
	templateIDSet := make(map[int64]bool)
	for _, wop := range wops {
		if wop.ProcessTemplateID > 0 && !templateIDSet[wop.ProcessTemplateID] {
			templateIDs = append(templateIDs, wop.ProcessTemplateID)
			templateIDSet[wop.ProcessTemplateID] = true
		}
	}

	// 批量查询 ProcessTemplate（一次查询）
	if len(templateIDs) > 0 {
		var templates []*entity.ProcessTemplate
		if err := d.db.Where("id IN ?", templateIDs).Find(&templates).Error; err == nil {
			templateMap := make(map[int64]*entity.ProcessTemplate)
			for _, t := range templates {
				templateMap[t.ID] = t
			}

			// 关联到 wop
			for _, wop := range wops {
				if wop.ProcessTemplateID > 0 {
					if t, ok := templateMap[wop.ProcessTemplateID]; ok {
						wop.ProcessTemplate = t
					}
				}
			}
		}
	}

	return wops, nil
}

func (d *WorkOrderProcessDAO) Update(wop *entity.WorkOrderProcess) error {
	// 使用 Select 明确指定要更新的字段，避免零值字段被忽略
	// updated_at 由 GORM autoUpdateTime 自动管理，无需手动指定
	return d.db.Model(&entity.WorkOrderProcess{}).
		Where("id = ?", wop.ID).
		Select("is_completed", "completed_at", "remark").
		Updates(wop).Error
}

func (d *WorkOrderProcessDAO) Delete(id int64) error {
	return d.db.Delete(&entity.WorkOrderProcess{}, id).Error
}

func (d *WorkOrderProcessDAO) DeleteByWorkOrderID(workOrderID int64) error {
	return d.db.Where("work_order_id = ?", workOrderID).Delete(&entity.WorkOrderProcess{}).Error
}

// FindByWorkOrderIDs 批量查询多个工单的工序数据（优化N+1查询）
func (d *WorkOrderProcessDAO) FindByWorkOrderIDs(workOrderIDs []int64) ([]*entity.WorkOrderProcess, error) {
	if len(workOrderIDs) == 0 {
		return []*entity.WorkOrderProcess{}, nil
	}

	var wops []*entity.WorkOrderProcess
	err := d.db.Model(&entity.WorkOrderProcess{}).
		Joins("LEFT JOIN process_template pt ON work_order_process.process_template_id = pt.id").
		Where("work_order_process.work_order_id IN ?", workOrderIDs).
		Order("pt.sort ASC").
		Find(&wops).Error

	if err != nil {
		return nil, err
	}

	templateIDs := make([]int64, 0)
	templateIDSet := make(map[int64]bool)
	for _, wop := range wops {
		if wop.ProcessTemplateID > 0 && !templateIDSet[wop.ProcessTemplateID] {
			templateIDs = append(templateIDs, wop.ProcessTemplateID)
			templateIDSet[wop.ProcessTemplateID] = true
		}
	}

	// 批量查询 ProcessTemplate（一次查询）
	if len(templateIDs) > 0 {
		var templates []*entity.ProcessTemplate
		if err := d.db.Where("id IN ?", templateIDs).Find(&templates).Error; err == nil {
			templateMap := make(map[int64]*entity.ProcessTemplate)
			for _, t := range templates {
				templateMap[t.ID] = t
			}

			// 关联到 wop
			for _, wop := range wops {
				if wop.ProcessTemplateID > 0 {
					if t, ok := templateMap[wop.ProcessTemplateID]; ok {
						wop.ProcessTemplate = t
					}
				}
			}
		}
	}

	return wops, nil
}

// FindOrCreate 查找或创建工单工序关联
func (d *WorkOrderProcessDAO) FindOrCreate(workOrderID, processTemplateID int64) (*entity.WorkOrderProcess, error) {
	var wop entity.WorkOrderProcess
	err := d.db.Where("work_order_id = ? AND process_template_id = ?", workOrderID, processTemplateID).
		FirstOrCreate(&wop, entity.WorkOrderProcess{
			WorkOrderID:       workOrderID,
			ProcessTemplateID: processTemplateID,
			IsCompleted:       0,
		}).Error
	return &wop, err
}

// FindByWorkOrderIDAndTemplateID 根据工单ID和工序模板ID查询工序状态
func (d *WorkOrderProcessDAO) FindByWorkOrderIDAndTemplateID(workOrderID, templateID int64) (*entity.WorkOrderProcess, error) {
	var wop entity.WorkOrderProcess
	err := d.db.Where("work_order_id = ? AND process_template_id = ?", workOrderID, templateID).First(&wop).Error
	if err != nil {
		return nil, err
	}
	return &wop, nil
}

// UpsertProcess 创建或更新工序状态
func (d *WorkOrderProcessDAO) UpsertProcess(workOrderID, templateID int64, status int8) error {
	var now interface{}
	if status == 1 {
		now = gorm.Expr("NOW()")
	} else {
		now = nil
	}
	return d.db.Where("work_order_id = ? AND process_template_id = ?", workOrderID, templateID).
		Assign(map[string]interface{}{
			"is_completed": status,
			"completed_at": now,
		}).
		FirstOrCreate(&entity.WorkOrderProcess{
			WorkOrderID:       workOrderID,
			ProcessTemplateID: templateID,
			IsCompleted:       status,
		}).Error
}
