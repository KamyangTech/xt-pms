package dao

import "github.com/KamyangTech/xt-pms/server/internal/model/entity"

// ProductBOMDAO 产品 BOM 数据访问对象
type ProductBOMDAO struct{}

func NewProductBOMDAO() *ProductBOMDAO {
	return &ProductBOMDAO{}
}

func (d *ProductBOMDAO) Create(bom *entity.ProductBOM) error {
	return DB.Create(bom).Error
}

func (d *ProductBOMDAO) FindByID(id int64) (*entity.ProductBOM, error) {
	var bom entity.ProductBOM
	err := DB.First(&bom, id).Error
	if err != nil {
		return nil, err
	}
	return &bom, nil
}

func (d *ProductBOMDAO) FindByProductID(productID int64) ([]*entity.ProductBOM, error) {
	var boms []*entity.ProductBOM
	err := DB.Where("product_id = ?", productID).Find(&boms).Error
	return boms, err
}

func (d *ProductBOMDAO) Update(bom *entity.ProductBOM) error {
	return DB.Save(bom).Error
}

func (d *ProductBOMDAO) Delete(id int64) error {
	return DB.Delete(&entity.ProductBOM{}, id).Error
}

func (d *ProductBOMDAO) DeleteByProductID(productID int64) error {
	return DB.Where("product_id = ?", productID).Delete(&entity.ProductBOM{}).Error
}
