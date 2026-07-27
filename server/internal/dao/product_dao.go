package dao

import (
	"github.com/KamyangTech/xt-pms/server/internal/model/entity"
	"github.com/KamyangTech/xt-pms/server/internal/utils"
	"gorm.io/gorm"
)

type ProductDAO struct {
	db *gorm.DB
}

func NewProductDAO() *ProductDAO {
	return &ProductDAO{db: DB}
}

// NewProductDAOWithDB 使用指定的 DB（通常用于事务中）
func NewProductDAOWithDB(db *gorm.DB) *ProductDAO {
	return &ProductDAO{db: db}
}

func (d *ProductDAO) getDB() *gorm.DB {
	if d.db != nil {
		return d.db
	}
	return DB
}

func (d *ProductDAO) Create(product *entity.Product) error {
	return d.getDB().Create(product).Error
}

func (d *ProductDAO) FindByID(id int64) (*entity.Product, error) {
	var product entity.Product
	err := d.getDB().First(&product, id).Error
	if err != nil {
		return nil, err
	}
	return &product, nil
}

func (d *ProductDAO) FindByCode(code string) (*entity.Product, error) {
	var product entity.Product
	err := d.getDB().Where("pro_code = ?", code).First(&product).Error
	if err != nil {
		return nil, err
	}
	return &product, nil
}

func (d *ProductDAO) FindByName(name string) (*entity.Product, error) {
	var product entity.Product
	err := d.getDB().Where("pro_name = ?", name).First(&product).Error
	if err != nil {
		return nil, err
	}
	return &product, nil
}

// FindByNames 批量根据产品名称查询（避免 N+1）
func (d *ProductDAO) FindByNames(names []string) ([]*entity.Product, error) {
	if len(names) == 0 {
		return nil, nil
	}
	var products []*entity.Product
	err := d.getDB().Where("pro_name IN ?", names).Find(&products).Error
	return products, err
}

func (d *ProductDAO) FindAll() ([]*entity.Product, error) {
	var products []*entity.Product
	err := d.getDB().Select("id", "pro_code", "pro_name", "status").Order("id DESC").Limit(1000).Find(&products).Error
	return products, err
}

func (d *ProductDAO) FindPage(page, pageSize int) (*utils.PageResult, error) {
	var products []*entity.Product
	return utils.Paginate(d.getDB(), page, pageSize, &products, func(db *gorm.DB) *gorm.DB {
		return db.Order("id DESC")
	})
}

func (d *ProductDAO) Update(product *entity.Product) error {
	return d.getDB().Save(product).Error
}

func (d *ProductDAO) Delete(id int64) error {
	return d.getDB().Delete(&entity.Product{}, id).Error
}

func (d *ProductDAO) UpdateStatus(id int64, status int8) error {
	return d.getDB().Model(&entity.Product{}).Where("id = ?", id).Update("status", status).Error
}
