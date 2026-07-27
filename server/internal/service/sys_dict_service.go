package service

import (
	"sync/atomic"
	"time"

	"github.com/KamyangTech/xt-pms/server/internal/dao"
	"github.com/KamyangTech/xt-pms/server/internal/model/entity"
	"github.com/KamyangTech/xt-pms/server/internal/utils"
)

// 字典版本号，每次字典数据或类型变更时递增
var dictVersion atomic.Int64

// GetDictVersion 获取当前字典版本号
func GetDictVersion() int64 {
	return dictVersion.Load()
}

func bumpDictVersion() {
	dictVersion.Add(1)
}

// PaymentTypeNeedsDeliveryUnlock 检查款项类型是否需要交付后解锁
// 从字典 payment-type 中读取 needs_delivery_unlock 标志，替代硬编码的 "final-payment" 判断
func PaymentTypeNeedsDeliveryUnlock(paymentType string) bool {
	if paymentType == "" {
		return false
	}

	cache := utils.GetCache()
	cacheKey := "payment_type_unlock:" + paymentType

	if cached, ok := cache.Get(cacheKey); ok {
		if val, ok2 := cached.(bool); ok2 {
			return val
		}
	}

	var data entity.SysDictData
	if err := dao.DB.Where("dict_value = ? AND status = 1", paymentType).
		Joins("JOIN sys_dict_type ON sys_dict_type.id = sys_dict_data.dict_type_id").
		Where("sys_dict_type.dict_code = ?", "payment-type").
		First(&data).Error; err != nil {
		return false
	}

	result := data.NeedsDeliveryUnlock == 1
	cache.Set(cacheKey, result, 10*time.Minute)
	return result
}

type SysDictTypeService struct {
	sysDictTypeDAO *dao.SysDictTypeDAO
}

func NewSysDictTypeService() *SysDictTypeService {
	return &SysDictTypeService{
		sysDictTypeDAO: dao.NewSysDictTypeDAO(),
	}
}

// GetList 获取字典类型列表（带缓存）
func (s *SysDictTypeService) GetList() ([]*entity.SysDictType, error) {
	// 尝试从缓存获取
	cache := utils.GetCache()
	cacheKey := "dict_types_all"

	if cached, ok := cache.Get(cacheKey); ok {
		if dictTypes, ok := cached.([]*entity.SysDictType); ok {
			return dictTypes, nil
		}
	}

	// 缓存未命中，查询数据库
	dictTypes, err := s.sysDictTypeDAO.FindAll()
	if err != nil {
		return nil, err
	}

	// 写入缓存，有效期30分钟（字典数据很少变更）
	cache.Set(cacheKey, dictTypes, 30*time.Minute)

	return dictTypes, nil
}

// GetPage 分页获取字典类型列表
func (s *SysDictTypeService) GetPage(page, pageSize int) ([]*entity.SysDictType, int64, error) {
	result, err := s.sysDictTypeDAO.FindPage(page, pageSize)
	if err != nil {
		return nil, 0, err
	}
	return result.List.([]*entity.SysDictType), result.Total, nil
}

// GetPageWithFilter 分页查询字典类型列表（带名称和编码筛选）
func (s *SysDictTypeService) GetPageWithFilter(page, pageSize int, dictName, dictCode string) ([]*entity.SysDictType, int64, error) {
	return s.sysDictTypeDAO.FindPageWithFilter(page, pageSize, dictName, dictCode)
}

// GetByID 根据 ID 获取字典类型
func (s *SysDictTypeService) GetByID(id int64) (*entity.SysDictType, error) {
	return s.sysDictTypeDAO.FindByID(id)
}

// GetByCode 根据编码获取字典类型
func (s *SysDictTypeService) GetByCode(code string) (*entity.SysDictType, error) {
	return s.sysDictTypeDAO.FindByCode(code)
}

// Create 创建字典类型
func (s *SysDictTypeService) Create(dictType *entity.SysDictType) error {
	err := s.sysDictTypeDAO.Create(dictType)
	if err == nil {
		// 清除缓存并递增版本号
		utils.GetCache().Delete("dict_types_all")
		bumpDictVersion()
	}
	return err
}

// Update 更新字典类型
func (s *SysDictTypeService) Update(dictType *entity.SysDictType) error {
	err := s.sysDictTypeDAO.Update(dictType)
	if err == nil {
		// 清除缓存并递增版本号
		utils.GetCache().Delete("dict_types_all")
		bumpDictVersion()
	}
	return err
}

// Delete 删除字典类型
func (s *SysDictTypeService) Delete(id int64) error {
	err := s.sysDictTypeDAO.Delete(id)
	if err == nil {
		// 清除缓存并递增版本号
		utils.GetCache().Delete("dict_types_all")
		bumpDictVersion()
	}
	return err
}

// 受保护字典类型列表 - 字典值被后端代码硬编码引用，禁止修改 dict_value 和删除
var ProtectedDictTypes = map[string]string{
	"mold_order_type":           "订单/工单类型",
	"production-type":           "生产类型",
	"mold-source":               "模具来源",
	"production-order-status":   "生产订单状态",
	"master-order-status":       "综合订单状态",
	"outsourcing-status":        "外发状态",
	"receivable-status":         "应收状态",
	"payable-status":            "应付状态",
	"aging-category":            "账龄分类",
	"payment-type":              "款项类型",
	"master-order-type":         "综合订单类型",
	"dependency-status":         "依赖状态",
}

func IsProtectedDictType(dictCode string) bool {
	_, ok := ProtectedDictTypes[dictCode]
	return ok
}

func GetProtectedDictTypeName(dictCode string) string {
	return ProtectedDictTypes[dictCode]
}

// ProtectedDictDataValues 受保护字典数据值列表 - 特定值的 dict_value 不可修改
// key: dict_code, value: 受保护的 dict_value 列表
var ProtectedDictDataValues = map[string][]string{
	"order-priority": {"NORMAL"},
}

func IsProtectedDictDataValue(dictCode, dictValue string) bool {
	values, ok := ProtectedDictDataValues[dictCode]
	if !ok {
		return false
	}
	for _, v := range values {
		if dictValue == v {
			return true
		}
	}
	return false
}

func GetProtectedDictDataValues(dictCode string) []string {
	return ProtectedDictDataValues[dictCode]
}

type SysDictDataService struct {
	sysDictDataDAO *dao.SysDictDataDAO
}

func NewSysDictDataService() *SysDictDataService {
	return &SysDictDataService{
		sysDictDataDAO: dao.NewSysDictDataDAO(),
	}
}

// GetByTypeID 根据字典类型 ID 获取字典数据列表
func (s *SysDictDataService) GetByTypeID(dictTypeID int64) ([]*entity.SysDictData, error) {
	return s.sysDictDataDAO.FindByTypeID(dictTypeID)
}

// GetByID 根据 ID 获取字典数据
func (s *SysDictDataService) GetByID(id int64) (*entity.SysDictData, error) {
	return s.sysDictDataDAO.FindByID(id)
}

// Create 创建字典数据
func (s *SysDictDataService) Create(dict *entity.SysDictData) error {
	err := s.sysDictDataDAO.Create(dict)
	if err == nil {
		// 清除对应字典类型的缓存并递增版本号
		dictType, err := dao.NewSysDictTypeDAO().FindByID(dict.DictTypeID)
		if err == nil && dictType != nil {
			utils.GetCache().Delete("dict_data_" + dictType.DictCode)
		}
		bumpDictVersion()
	}
	return err
}

// Delete 删除字典数据
func (s *SysDictDataService) Delete(id int64) error {
	// 先查询获取 dict_type_id 用于清除缓存
	existing, err := s.sysDictDataDAO.FindByID(id)
	if err != nil {
		return err
	}
	err = s.sysDictDataDAO.Delete(id)
	if err == nil {
		dictType, err := dao.NewSysDictTypeDAO().FindByID(existing.DictTypeID)
		if err == nil && dictType != nil {
			utils.GetCache().Delete("dict_data_" + dictType.DictCode)
		}
		bumpDictVersion()
	}
	return err
}

// GetLabelByCodeAndValue 根据字典类型编码和字典值获取对应的中文标签
// 用于通知等场景中把字典值（如 BANK_TRANSFER）翻译为中文（如 银行转账）
// 未匹配到时返回原值
func (s *SysDictDataService) GetLabelByCodeAndValue(dictCode, dictValue string) (string, error) {
	if dictValue == "" {
		return "", nil
	}
	list, err := s.GetByTypeCode(dictCode)
	if err != nil {
		return dictValue, err
	}
	for _, item := range list {
		if item.DictValue == dictValue {
			return item.DictLabel, nil
		}
	}
	return dictValue, nil
}

// GetByTypeCode 根据字典类型编码获取字典数据列表（带缓存）
func (s *SysDictDataService) GetByTypeCode(dictCode string) ([]*entity.SysDictData, error) {
	// 尝试从缓存获取
	cache := utils.GetCache()
	cacheKey := "dict_data_" + dictCode

	if cached, ok := cache.Get(cacheKey); ok {
		if dictData, ok := cached.([]*entity.SysDictData); ok {
			return dictData, nil
		}
	}

	// 缓存未命中，查询数据库
	dictData, err := s.sysDictDataDAO.FindByTypeCode(dictCode)
	if err != nil {
		return nil, err
	}

	// 写入缓存，有效期30分钟（字典数据很少变更）
	cache.Set(cacheKey, dictData, 30*time.Minute)

	return dictData, nil
}

// GetPage 分页获取字典数据列表
func (s *SysDictDataService) GetPage(dictTypeID int64, page, pageSize int) ([]*entity.SysDictData, int64, error) {
	offset := (page - 1) * pageSize
	return s.sysDictDataDAO.FindPage(dictTypeID, offset, pageSize)
}

// Update 更新字典数据（系统内置状态仅允许修改颜色和备注）
func (s *SysDictDataService) Update(dict *entity.SysDictData) error {
	err := s.sysDictDataDAO.Update(dict)
	if err == nil {
		// 清除对应字典类型的缓存并递增版本号
		dictType, err := dao.NewSysDictTypeDAO().FindByID(dict.DictTypeID)
		if err == nil && dictType != nil {
			utils.GetCache().Delete("dict_data_" + dictType.DictCode)
		}
		bumpDictVersion()
	}
	return err
}
