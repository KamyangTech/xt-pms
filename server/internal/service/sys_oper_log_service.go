package service

import (
	"encoding/csv"
	"fmt"
	"io"
	"strconv"

	"github.com/KamyangTech/xt-pms/server/internal/dao"
	"github.com/KamyangTech/xt-pms/server/internal/model/entity"
)

type SysOperLogService struct {
	dao *dao.SysOperLogDAO
}

// NewSysOperLogService 创建操作日志Service实例
func NewSysOperLogService() *SysOperLogService {
	return &SysOperLogService{
		dao: dao.NewSysOperLogDAO(),
	}
}

// GetListByPage 分页查询操作日志
func (s *SysOperLogService) GetListByPage(page, pageSize int, title, operName string, operType int8, status int8, businessType, startTime, endTime string) ([]*entity.SysOperLog, int64, error) {
	offset := (page - 1) * pageSize
	return s.dao.FindPageWithFilter(offset, pageSize, title, operName, operType, status, businessType, startTime, endTime)
}

// GetByID 获取操作日志详情
func (s *SysOperLogService) GetByID(id int64) (*entity.SysOperLog, error) {
	return s.dao.FindByID(id)
}

const maxExport = 10000

// ExportCSV 将操作日志导出为 CSV 并写入 writer
func (s *SysOperLogService) ExportCSV(w io.Writer, title, operName string, operType int8, status int8, businessType, startTime, endTime string) error {
	list, err := s.dao.FindAllForExport(title, operName, operType, status, businessType, startTime, endTime, maxExport)
	if err != nil {
		return fmt.Errorf("查询导出数据失败: %w", err)
	}

	// BOM 头，解决 Excel 打开 CSV 中文乱码
	w.Write([]byte{0xEF, 0xBB, 0xBF})

	cw := csv.NewWriter(w)
	defer cw.Flush()

	// 表头
	cw.Write([]string{"ID", "操作时间", "操作模块", "操作类型", "操作人", "操作IP", "请求方式", "请求URL", "业务类型", "耗时(ms)", "状态"})

	// 数据行
	for _, log := range list {
		statusText := "正常"
		if log.Status == 1 {
			statusText = "异常"
		}
		cw.Write([]string{
			strconv.FormatInt(log.ID, 10),
			log.CreateTime.Format("2006-01-02 15:04:05"),
			log.Title,
			entity.OperTypeName(log.OperType),
			log.OperName,
			log.OperIP,
			log.RequestMethod,
			log.RequestURL,
			log.BusinessType,
			strconv.FormatInt(log.CostTime, 10),
			statusText,
		})
	}

	return nil
}
