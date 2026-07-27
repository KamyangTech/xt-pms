package dao

import (
	"github.com/KamyangTech/xt-pms/server/internal/model/entity"
	"github.com/KamyangTech/xt-pms/server/internal/utils"
	"gorm.io/gorm"
)

// SysDictTypeDAO 字典类型 DAO
type SysDictTypeDAO struct{}

// NewSysDictTypeDAO 创建字典类型 DAO 实例
func NewSysDictTypeDAO() *SysDictTypeDAO {
	return &SysDictTypeDAO{}
}

// Create 创建字典类型
func (d *SysDictTypeDAO) Create(dictType *entity.SysDictType) error {
	return DB.Create(dictType).Error
}

// FindByID 根据 ID 查询字典类型
func (d *SysDictTypeDAO) FindByID(id int64) (*entity.SysDictType, error) {
	var dictType entity.SysDictType
	err := DB.First(&dictType, id).Error
	if err != nil {
		return nil, err
	}
	return &dictType, nil
}

// FindByCode 根据编码查询字典类型
func (d *SysDictTypeDAO) FindByCode(code string) (*entity.SysDictType, error) {
	var dictType entity.SysDictType
	err := DB.Where("dict_code = ?", code).First(&dictType).Error
	if err != nil {
		return nil, err
	}
	return &dictType, nil
}

// FindAll 查询所有字典类型（带数据）
func (d *SysDictTypeDAO) FindAll() ([]*entity.SysDictType, error) {
	var dictTypes []*entity.SysDictType
	err := DB.Preload("DataList", func(db *gorm.DB) *gorm.DB {
		return db.Where("status = ?", 1).Order("sort ASC")
	}).Order("id ASC").Find(&dictTypes).Error
	return dictTypes, err
}

// FindPage 分页查询字典类型
func (d *SysDictTypeDAO) FindPage(page, pageSize int) (*utils.PageResult, error) {
	var dictTypes []*entity.SysDictType
	return utils.Paginate(DB, page, pageSize, &dictTypes, func(db *gorm.DB) *gorm.DB {
		return db.Order("id DESC")
	})
}

// FindPageWithFilter 分页查询字典类型（带名称和编码筛选）
func (d *SysDictTypeDAO) FindPageWithFilter(page, pageSize int, dictName, dictCode string) ([]*entity.SysDictType, int64, error) {
	var dictTypes []*entity.SysDictType
	var total int64

	query := DB.Model(&entity.SysDictType{})
	if dictName != "" {
		query = query.Where("dict_name LIKE ?", "%"+dictName+"%")
	}
	if dictCode != "" {
		query = query.Where("dict_code LIKE ?", dictCode+"%")
	}

	query.Count(&total)
	err := query.Offset((page - 1) * pageSize).Limit(pageSize).Order("id DESC").Find(&dictTypes).Error
	return dictTypes, total, err
}

// Update 更新字典类型
func (d *SysDictTypeDAO) Update(dictType *entity.SysDictType) error {
	return DB.Save(dictType).Error
}

// Delete 删除字典类型
func (d *SysDictTypeDAO) Delete(id int64) error {
	return DB.Delete(&entity.SysDictType{}, id).Error
}

// SysDictDataDAO 字典数据 DAO
type SysDictDataDAO struct{}

// NewSysDictDataDAO 创建字典数据 DAO 实例
func NewSysDictDataDAO() *SysDictDataDAO {
	return &SysDictDataDAO{}
}

// FindByTypeID 根据字典类型 ID 查询所有字典数据
func (d *SysDictDataDAO) FindByTypeID(dictTypeID int64) ([]*entity.SysDictData, error) {
	var dicts []*entity.SysDictData
	err := DB.Where("dict_type_id = ?", dictTypeID).Order("sort ASC").Find(&dicts).Error
	return dicts, err
}

// FindByTypeCode 根据字典类型编码查询所有字典数据
func (d *SysDictDataDAO) FindByTypeCode(dictCode string) ([]*entity.SysDictData, error) {
	var dicts []*entity.SysDictData
	err := DB.Joins("JOIN sys_dict_type ON sys_dict_type.id = sys_dict_data.dict_type_id").
		Where("sys_dict_type.dict_code = ?", dictCode).
		Order("sort ASC").
		Find(&dicts).Error
	return dicts, err
}

// FindAll 查询所有字典数据
func (d *SysDictDataDAO) FindAll() ([]*entity.SysDictData, error) {
	var dicts []*entity.SysDictData
	err := DB.Preload("DictType").Order("dict_type_id, sort ASC").Find(&dicts).Error
	return dicts, err
}

// FindPage 分页查询字典数据
func (d *SysDictDataDAO) FindPage(dictTypeID int64, offset, limit int) ([]*entity.SysDictData, int64, error) {
	var dicts []*entity.SysDictData
	var total int64

	query := DB.Model(&entity.SysDictData{})
	if dictTypeID > 0 {
		query = query.Where("dict_type_id = ?", dictTypeID)
	}

	query.Count(&total)
	err := query.Offset(offset).Limit(limit).Order("sort ASC").Find(&dicts).Error
	return dicts, total, err
}

// Create 创建字典数据
func (d *SysDictDataDAO) Create(dict *entity.SysDictData) error {
	return DB.Create(dict).Error
}

// Update 更新字典数据
func (d *SysDictDataDAO) Update(dict *entity.SysDictData) error {
	// 只更新指定的字段，忽略 create_time
	return DB.Model(dict).Where("id = ?", dict.ID).Updates(map[string]interface{}{
		"dict_type_id":      dict.DictTypeID,
		"dict_label":        dict.DictLabel,
		"dict_value":        dict.DictValue,
		"color":             dict.Color,
		"sort":              dict.Sort,
		"status":            dict.Status,
		"is_locked":         dict.IsLocked,
		"allow_production":  dict.AllowProduction,
		"allow_work_report": dict.AllowWorkReport,
		"remark":            dict.Remark,
		"ext":               dict.Ext,
	}).Error
}

// Delete 删除字典数据
func (d *SysDictDataDAO) Delete(id int64) error {
	return DB.Delete(&entity.SysDictData{}, id).Error
}

// FindByID 根据 ID 查询字典数据
func (d *SysDictDataDAO) FindByID(id int64) (*entity.SysDictData, error) {
	var dict entity.SysDictData
	err := DB.First(&dict, id).Error
	if err != nil {
		return nil, err
	}
	return &dict, nil
}

// FindByDictTypeIDAndLabel 根据字典类型ID和标签查询字典数据（用于检查重复）
func (d *SysDictDataDAO) FindByDictTypeIDAndLabel(dictTypeID int64, dictLabel string) (*entity.SysDictData, error) {
	var dict entity.SysDictData
	err := DB.Where("dict_type_id = ? AND dict_label = ?", dictTypeID, dictLabel).First(&dict).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &dict, nil
}

// FindByDictTypeIDAndValue 根据字典类型ID和值查询字典数据（用于检查重复）
func (d *SysDictDataDAO) FindByDictTypeIDAndValue(dictTypeID int64, dictValue string) (*entity.SysDictData, error) {
	var dict entity.SysDictData
	err := DB.Where("dict_type_id = ? AND dict_value = ?", dictTypeID, dictValue).First(&dict).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &dict, nil
}

// FindByTypeCodeAndValue 根据字典类型编码和值查询字典数据
func (d *SysDictDataDAO) FindByTypeCodeAndValue(dictCode string, dictValue string) (*entity.SysDictData, error) {
	var dict entity.SysDictData
	err := DB.Joins("JOIN sys_dict_type ON sys_dict_type.id = sys_dict_data.dict_type_id").
		Where("sys_dict_type.dict_code = ? AND sys_dict_data.dict_value = ?", dictCode, dictValue).
		First(&dict).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &dict, nil
}
