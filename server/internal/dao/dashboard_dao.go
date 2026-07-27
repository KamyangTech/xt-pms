package dao

import (
	"time"

	"github.com/KamyangTech/xt-pms/server/internal/model/constant"
	"github.com/KamyangTech/xt-pms/server/internal/model/dto"
	"github.com/KamyangTech/xt-pms/server/internal/model/entity"
	"gorm.io/gorm"
)

type DashboardDAO struct {
	db *gorm.DB
}

// NewDashboardDAO 创建 DashboardDAO 实例
func NewDashboardDAO() *DashboardDAO {
	return &DashboardDAO{
		db: GetDB(),
	}
}

// GetTotalMolds 获取模具总数
func (d *DashboardDAO) GetTotalMolds() (int64, error) {
	var count int64
	err := d.db.Model(&entity.Mold{}).Count(&count).Error
	return count, err
}

// GetMoldsByStatus 根据状态获取模具数量（已废弃，使用 GetMoldsByLifeCycleStatus 代替）
func (d *DashboardDAO) GetMoldsByStatus(status int8) (int64, error) {
	var count int64
	err := d.db.Model(&entity.Mold{}).Where("life_cycle_status = ?", status).Count(&count).Error
	return count, err
}

// GetMoldsByLifeCycleStatus 根据生命周期状态获取模具数量（使用字符串状态值）
func (d *DashboardDAO) GetMoldsByLifeCycleStatus(status string) (int64, error) {
	var count int64
	err := d.db.Model(&entity.Mold{}).Where("life_cycle_status = ?", status).Count(&count).Error
	return count, err
}

// GetTotalCustomers 获取客户总数
func (d *DashboardDAO) GetTotalCustomers() (int64, error) {
	var count int64
	err := d.db.Model(&entity.Customer{}).Count(&count).Error
	return count, err
}

// GetTotalSuppliers 获取供应商总数
func (d *DashboardDAO) GetTotalSuppliers() (int64, error) {
	var count int64
	err := d.db.Model(&entity.Supplier{}).Count(&count).Error
	return count, err
}

// GetStockWarningCount 获取库存预警数量
func (d *DashboardDAO) GetStockWarningCount() (int64, error) {
	var count int64
	err := d.db.Model(&entity.Stock{}).
		Joins("JOIN material ON stock.material_id = material.id").
		Where("stock.stock_qty < material.stock_warn").
		Count(&count).Error
	return count, err
}

// GetPendingWorkOrders 获取待处理工单数量
func (d *DashboardDAO) GetPendingWorkOrders() (int64, error) {
	var count int64
	// 统计模具工单中状态为 1(未开始)的数量
	err := d.db.Model(&entity.MoldWorkOrder{}).Where("status = ?", 1).Count(&count).Error
	return count, err
}

// GetTodayOutput 获取今日产量
func (d *DashboardDAO) GetTodayOutput() (int64, error) {
	var total int64
	today := time.Now().Format("2006-01-02")
	err := d.db.Model(&entity.ProductionReport{}).
		Where("DATE(report_date) = ?", today).
		Select("COALESCE(SUM(ok_qty), 0)").
		Scan(&total).Error
	return total, err
}

// GetYesterdayOutput 获取昨日产量
func (d *DashboardDAO) GetYesterdayOutput() (int64, error) {
	var total int64
	yesterday := time.Now().AddDate(0, 0, -1).Format("2006-01-02")
	err := d.db.Model(&entity.ProductionReport{}).
		Where("DATE(report_date) = ?", yesterday).
		Select("COALESCE(SUM(ok_qty), 0)").
		Scan(&total).Error
	return total, err
}

// GetRawMaterialStock 获取物料库存种类数
func (d *DashboardDAO) GetRawMaterialStock() (int64, error) {
	var count int64
	err := d.db.Model(&entity.Stock{}).Count(&count).Error
	return count, err
}

// GetQualificationRate 获取合格率
func (d *DashboardDAO) GetQualificationRate() (float64, error) {
	type Result struct {
		TotalQty int64
		OkQty    int64
	}
	var result Result
	err := d.db.Model(&entity.ProductionReport{}).
		Select("SUM(ok_qty + bad_qty) as total_qty, SUM(ok_qty) as ok_qty").
		Scan(&result).Error
	if err != nil || result.TotalQty == 0 {
		return 0, err
	}
	return float64(result.OkQty) / float64(result.TotalQty) * 100, nil
}

// GetMoldsCountYesterday 获取昨日模具总数（用于计算增长率）
func (d *DashboardDAO) GetMoldsCountYesterday() (int64, error) {
	var count int64
	yesterday := time.Now().AddDate(0, 0, -1)
	err := d.db.Model(&entity.Mold{}).
		Where("create_time <= ?", yesterday).
		Count(&count).Error
	return count, err
}

// GetStockWarningCountYesterday 获取昨日库存预警数量
func (d *DashboardDAO) GetStockWarningCountYesterday() (int64, error) {
	var count int64
	yesterday := time.Now().AddDate(0, 0, -1)
	err := d.db.Model(&entity.Stock{}).
		Joins("JOIN material ON stock.material_id = material.id").
		Where("stock.stock_qty < material.stock_warn").
		Where("stock.update_time <= ?", yesterday).
		Count(&count).Error
	return count, err
}

// GetPendingWorkOrdersYesterday 获取昨日待处理工单数量
func (d *DashboardDAO) GetPendingWorkOrdersYesterday() (int64, error) {
	var count int64
	yesterday := time.Now().AddDate(0, 0, -1)
	err := d.db.Model(&entity.MoldWorkOrder{}).
		Where("status = ?", 1).
		Where("create_time <= ?", yesterday).
		Count(&count).Error
	return count, err
}

// GetCustomersCountYesterday 获取昨日客户总数
func (d *DashboardDAO) GetCustomersCountYesterday() (int64, error) {
	var count int64
	yesterday := time.Now().AddDate(0, 0, -1)
	err := d.db.Model(&entity.Customer{}).
		Where("create_time <= ?", yesterday).
		Count(&count).Error
	return count, err
}

// GetSuppliersCountYesterday 获取昨日供应商总数
func (d *DashboardDAO) GetSuppliersCountYesterday() (int64, error) {
	var count int64
	yesterday := time.Now().AddDate(0, 0, -1)
	err := d.db.Model(&entity.Supplier{}).
		Where("create_time <= ?", yesterday).
		Count(&count).Error
	return count, err
}

// GetRawMaterialStockYesterday 获取昨日物料库存种类数
func (d *DashboardDAO) GetRawMaterialStockYesterday() (int64, error) {
	var count int64
	yesterday := time.Now().AddDate(0, 0, -1)
	err := d.db.Model(&entity.Stock{}).
		Where("create_time <= ?", yesterday).
		Count(&count).Error
	return count, err
}

// GetQualificationRateYesterday 获取昨日合格率
func (d *DashboardDAO) GetQualificationRateYesterday() (float64, error) {
	type Result struct {
		TotalQty int64
		OkQty    int64
	}
	var result Result
	yesterday := time.Now().AddDate(0, 0, -1)
	err := d.db.Model(&entity.ProductionReport{}).
		Where("DATE(report_date) <= ?", yesterday.Format("2006-01-02")).
		Select("SUM(ok_qty + bad_qty) as total_qty, SUM(ok_qty) as ok_qty").
		Scan(&result).Error
	if err != nil || result.TotalQty == 0 {
		return 0, err
	}
	return float64(result.OkQty) / float64(result.TotalQty) * 100, nil
}

// === 合并查询方法（一次 SQL 返回当前值和昨日值，减少 DB 查询次数）===

type totalYesterdayCount struct {
	Total     int64
	Yesterday int64
}

type totalYesterdayRate struct {
	Total     float64
	Yesterday float64
}

// GetMoldCounts 一次查询返回模具总数和昨日数量
func (d *DashboardDAO) GetMoldCounts() (total int64, yesterday int64, err error) {
	var result totalYesterdayCount
	yesterdayTime := time.Now().AddDate(0, 0, -1)
	err = d.db.Model(&entity.Mold{}).
		Select("COUNT(*) as total, SUM(CASE WHEN create_time <= ? THEN 1 ELSE 0 END) as yesterday", yesterdayTime).
		Scan(&result).Error
	return result.Total, result.Yesterday, err
}

// GetCustomerCounts 一次查询返回客户总数和昨日数量
func (d *DashboardDAO) GetCustomerCounts() (total int64, yesterday int64, err error) {
	var result totalYesterdayCount
	yesterdayTime := time.Now().AddDate(0, 0, -1)
	err = d.db.Model(&entity.Customer{}).
		Select("COUNT(*) as total, SUM(CASE WHEN create_time <= ? THEN 1 ELSE 0 END) as yesterday", yesterdayTime).
		Scan(&result).Error
	return result.Total, result.Yesterday, err
}

// GetSupplierCounts 一次查询返回供应商总数和昨日数量
func (d *DashboardDAO) GetSupplierCounts() (total int64, yesterday int64, err error) {
	var result totalYesterdayCount
	yesterdayTime := time.Now().AddDate(0, 0, -1)
	err = d.db.Model(&entity.Supplier{}).
		Select("COUNT(*) as total, SUM(CASE WHEN create_time <= ? THEN 1 ELSE 0 END) as yesterday", yesterdayTime).
		Scan(&result).Error
	return result.Total, result.Yesterday, err
}

// GetStockWarningCounts 一次查询返回库存预警数量和昨日数量
func (d *DashboardDAO) GetStockWarningCounts() (total int64, yesterday int64, err error) {
	var result totalYesterdayCount
	yesterdayTime := time.Now().AddDate(0, 0, -1)
	err = d.db.Model(&entity.Stock{}).
		Joins("JOIN material ON stock.material_id = material.id").
		Where("stock.stock_qty < material.stock_warn").
		Select("COUNT(*) as total, SUM(CASE WHEN stock.update_time <= ? THEN 1 ELSE 0 END) as yesterday", yesterdayTime).
		Scan(&result).Error
	return result.Total, result.Yesterday, err
}

// GetPendingWorkOrderCounts 一次查询返回待处理工单数量和昨日数量
func (d *DashboardDAO) GetPendingWorkOrderCounts() (total int64, yesterday int64, err error) {
	var result totalYesterdayCount
	yesterdayTime := time.Now().AddDate(0, 0, -1)
	err = d.db.Model(&entity.MoldWorkOrder{}).
		Where("status = ?", 1).
		Select("COUNT(*) as total, SUM(CASE WHEN create_time <= ? THEN 1 ELSE 0 END) as yesterday", yesterdayTime).
		Scan(&result).Error
	return result.Total, result.Yesterday, err
}

// GetRawMaterialStockCounts 一次查询返回物料库存种类数和昨日数量
func (d *DashboardDAO) GetRawMaterialStockCounts() (total int64, yesterday int64, err error) {
	var result totalYesterdayCount
	yesterdayTime := time.Now().AddDate(0, 0, -1)
	err = d.db.Model(&entity.Stock{}).
		Select("COUNT(*) as total, SUM(CASE WHEN create_time <= ? THEN 1 ELSE 0 END) as yesterday", yesterdayTime).
		Scan(&result).Error
	return result.Total, result.Yesterday, err
}

// GetQualificationRates 一次查询返回合格率和昨日合格率
func (d *DashboardDAO) GetQualificationRates() (total float64, yesterday float64, err error) {
	type rateResult struct {
		TotalQty     int64
		OkQty        int64
		YesterdayQty int64
		YesterdayOk  int64
	}
	var result rateResult
	yesterdayTime := time.Now().AddDate(0, 0, -1).Format("2006-01-02")
	err = d.db.Model(&entity.ProductionReport{}).
		Select(`
			SUM(ok_qty + bad_qty) as total_qty,
			SUM(ok_qty) as ok_qty,
			SUM(CASE WHEN DATE(report_date) <= ? THEN ok_qty + bad_qty ELSE 0 END) as yesterday_qty,
			SUM(CASE WHEN DATE(report_date) <= ? THEN ok_qty ELSE 0 END) as yesterday_ok
		`, yesterdayTime, yesterdayTime).
		Scan(&result).Error
	if err != nil {
		return 0, 0, err
	}
	if result.TotalQty > 0 {
		total = float64(result.OkQty) / float64(result.TotalQty) * 100
	}
	if result.YesterdayQty > 0 {
		yesterday = float64(result.YesterdayOk) / float64(result.YesterdayQty) * 100
	}
	return total, yesterday, nil
}

// GetInProductionMoldCounts 一次查询返回生产中模具数量和昨日数量
func (d *DashboardDAO) GetInProductionMoldCounts() (total int64, yesterday int64, err error) {
	var result totalYesterdayCount
	yesterdayTime := time.Now().AddDate(0, 0, -1)
	err = d.db.Model(&entity.Mold{}).
		Where("life_cycle_status = ?", constant.MoldStatusProcessing).
		Select("COUNT(*) as total, SUM(CASE WHEN update_time <= ? THEN 1 ELSE 0 END) as yesterday", yesterdayTime).
		Scan(&result).Error
	return result.Total, result.Yesterday, err
}

// GetProductionOutputs 一次查询返回今日产量和昨日产量
func (d *DashboardDAO) GetProductionOutputs() (today int64, yesterday int64, err error) {
	type outputResult struct {
		Today     int64
		Yesterday int64
	}
	var result outputResult
	todayStr := time.Now().Format("2006-01-02")
	yesterdayStr := time.Now().AddDate(0, 0, -1).Format("2006-01-02")
	err = d.db.Model(&entity.ProductionReport{}).
		Select(`
			COALESCE(SUM(CASE WHEN DATE(report_date) = ? THEN ok_qty ELSE 0 END), 0) as today,
			COALESCE(SUM(CASE WHEN DATE(report_date) = ? THEN ok_qty ELSE 0 END), 0) as yesterday
		`, todayStr, yesterdayStr).
		Where("DATE(report_date) IN (?)", []string{todayStr, yesterdayStr}).
		Scan(&result).Error
	return result.Today, result.Yesterday, err
}
func (d *DashboardDAO) GetProductionTrendByRange(start, end time.Time) ([]string, []float64, []float64, error) {
	days := int(end.Sub(start).Hours()/24) + 1
	if days <= 0 {
		days = 1
	}

	dates := make([]string, days)
	currentData := make([]float64, days)

	// 使用日期字符串参数避免 time.Time 的时区转换问题
	// 参考 GetProductionOutputs 等正常工作的查询，统一用 DATE() + 字符串
	startDateStr := start.Format("2006-01-02")
	endDateStr := end.Format("2006-01-02")

	type dayAgg struct {
		Day   string
		Total float64
	}
	var results []dayAgg
	if err := d.db.Model(&entity.ProductionReport{}).
		Select("DATE_FORMAT(report_date, '%Y-%m-%d') AS day, COALESCE(SUM(ok_qty), 0) AS total").
		Where("DATE(report_date) >= ? AND DATE(report_date) <= ?", startDateStr, endDateStr).
		Group("DATE_FORMAT(report_date, '%Y-%m-%d')").
		Scan(&results).Error; err != nil {
		return dates, currentData, nil, err
	}

	totalMap := make(map[string]float64, len(results))
	for _, r := range results {
		totalMap[r.Day] = r.Total
	}

	for i := 0; i < days; i++ {
		day := start.AddDate(0, 0, i)
		dates[i] = day.Format("01-02")
		currentData[i] = totalMap[day.Format("2006-01-02")]
	}

	return dates, currentData, nil, nil
}

// GetProductionTrendByMonth 获取月度生产趋势（按报工产量汇总）
func (d *DashboardDAO) GetProductionTrendByMonth() ([]string, []float64, []float64, error) {
	now := time.Now()
	dates := make([]string, 30)
	currentData := make([]float64, 30)
	lastMonthData := make([]float64, 30)

	for i := 29; i >= 0; i-- {
		date := now.AddDate(0, 0, -i)
		dates[29-i] = date.Format("01-02")
	}

	startDate := now.AddDate(0, 0, -29)
	startDateStr := startDate.Format("2006-01-02")
	nowStr := now.Format("2006-01-02")

	type dayAgg struct {
		Day   string
		Total float64
	}

	// 使用 DATE() + 字符串参数，避免 time.Time 的时区转换
	var currentResults []dayAgg
	if err := d.db.Model(&entity.ProductionReport{}).
		Select("DATE_FORMAT(report_date, '%Y-%m-%d') AS day, COALESCE(SUM(ok_qty), 0) AS total").
		Where("DATE(report_date) >= ? AND DATE(report_date) <= ?", startDateStr, nowStr).
		Group("DATE_FORMAT(report_date, '%Y-%m-%d')").
		Scan(&currentResults).Error; err != nil {
		return dates, currentData, lastMonthData, err
	}

	currentMap := make(map[string]float64, len(currentResults))
	for _, r := range currentResults {
		currentMap[r.Day] = r.Total
	}

	// 上月同期数据
	lastStartDate := startDate.AddDate(0, -1, 0)
	lastEndDate := now.AddDate(0, -1, 0)
	lastStartDateStr := lastStartDate.Format("2006-01-02")
	lastEndDateStr := lastEndDate.Format("2006-01-02")
	var lastResults []dayAgg
	if err := d.db.Model(&entity.ProductionReport{}).
		Select("DATE_FORMAT(report_date, '%Y-%m-%d') AS day, COALESCE(SUM(ok_qty), 0) AS total").
		Where("DATE(report_date) >= ? AND DATE(report_date) <= ?", lastStartDateStr, lastEndDateStr).
		Group("DATE_FORMAT(report_date, '%Y-%m-%d')").
		Scan(&lastResults).Error; err != nil {
		return dates, currentData, lastMonthData, err
	}

	lastMap := make(map[string]float64, len(lastResults))
	for _, r := range lastResults {
		lastMap[r.Day] = r.Total
	}

	for i := 0; i < 30; i++ {
		day := startDate.AddDate(0, 0, i)
		currentData[i] = currentMap[day.Format("2006-01-02")]
		lastMonthData[i] = lastMap[day.AddDate(0, -1, 0).Format("2006-01-02")]
	}

	return dates, currentData, lastMonthData, nil
}

// GetProductionTrendByWeek 获取周度生产趋势（按报工产量汇总）
func (d *DashboardDAO) GetProductionTrendByWeek() ([]string, []float64, []float64, error) {
	now := time.Now()
	weekday := int(now.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	monday := now.AddDate(0, 0, -(weekday - 1))

	dates := make([]string, 7)
	currentData := make([]float64, 7)
	lastWeekData := make([]float64, 7)

	for i := 0; i < 7; i++ {
		dates[i] = monday.AddDate(0, 0, i).Format("01-02")
	}

	type dayAgg struct {
		Day   string
		Total float64
	}

	// 使用 DATE() + 字符串参数，避免 time.Time 的时区转换
	sunday := monday.AddDate(0, 0, 6)
	mondayStr := monday.Format("2006-01-02")
	sundayStr := sunday.Format("2006-01-02")

	var currentResults []dayAgg
	if err := d.db.Model(&entity.ProductionReport{}).
		Select("DATE_FORMAT(report_date, '%Y-%m-%d') AS day, COALESCE(SUM(ok_qty), 0) AS total").
		Where("DATE(report_date) >= ? AND DATE(report_date) <= ?", mondayStr, sundayStr).
		Group("DATE_FORMAT(report_date, '%Y-%m-%d')").
		Scan(&currentResults).Error; err != nil {
		return dates, currentData, lastWeekData, err
	}

	currentMap := make(map[string]float64, len(currentResults))
	for _, r := range currentResults {
		currentMap[r.Day] = r.Total
	}

	// 上周同期数据
	lastMonday := monday.AddDate(0, 0, -7)
	lastSunday := sunday.AddDate(0, 0, -7)
	lastMondayStr := lastMonday.Format("2006-01-02")
	lastSundayStr := lastSunday.Format("2006-01-02")
	var lastResults []dayAgg
	if err := d.db.Model(&entity.ProductionReport{}).
		Select("DATE_FORMAT(report_date, '%Y-%m-%d') AS day, COALESCE(SUM(ok_qty), 0) AS total").
		Where("DATE(report_date) >= ? AND DATE(report_date) <= ?", lastMondayStr, lastSundayStr).
		Group("DATE_FORMAT(report_date, '%Y-%m-%d')").
		Scan(&lastResults).Error; err != nil {
		return dates, currentData, lastWeekData, err
	}

	lastMap := make(map[string]float64, len(lastResults))
	for _, r := range lastResults {
		lastMap[r.Day] = r.Total
	}

	for i := 0; i < 7; i++ {
		day := monday.AddDate(0, 0, i)
		currentData[i] = currentMap[day.Format("2006-01-02")]
		lastWeekData[i] = lastMap[day.AddDate(0, 0, -7).Format("2006-01-02")]
	}

	return dates, currentData, lastWeekData, nil
}

// GetProductionTrendByYear 获取年度生产趋势（按报工产量汇总）
func (d *DashboardDAO) GetProductionTrendByYear() ([]string, []float64, []float64, error) {
	months := []string{"1 月", "2 月", "3 月", "4 月", "5 月", "6 月", "7 月", "8 月", "9 月", "10 月", "11 月", "12 月"}
	currentData := make([]float64, 12)
	lastYearData := make([]float64, 12)

	year := time.Now().Year()

	type monthAgg struct {
		Month int
		Total float64
	}

	// 优化：单次 GROUP BY 查询替代 12 次独立查询
	var currentResults []monthAgg
	if err := d.db.Model(&entity.ProductionReport{}).
		Select("MONTH(report_date) AS month, COALESCE(SUM(ok_qty), 0) AS total").
		Where("YEAR(report_date) = ?", year).
		Group("MONTH(report_date)").
		Scan(&currentResults).Error; err != nil {
		return months, currentData, lastYearData, err
	}

	for _, r := range currentResults {
		if r.Month >= 1 && r.Month <= 12 {
			currentData[r.Month-1] = r.Total
		}
	}

	// 去年同月数据
	var lastResults []monthAgg
	if err := d.db.Model(&entity.ProductionReport{}).
		Select("MONTH(report_date) AS month, COALESCE(SUM(ok_qty), 0) AS total").
		Where("YEAR(report_date) = ?", year-1).
		Group("MONTH(report_date)").
		Scan(&lastResults).Error; err != nil {
		return months, currentData, lastYearData, err
	}

	for _, r := range lastResults {
		if r.Month >= 1 && r.Month <= 12 {
			lastYearData[r.Month-1] = r.Total
		}
	}

	return months, currentData, lastYearData, nil
}

// =============================================
// 以下为顶部四张卡片的新增 DAO 方法
// =============================================

// GetMonthlyNewMoldOrders 获取本月新增模具订单数量
func (d *DashboardDAO) GetMonthlyNewMoldOrders() (int64, error) {
	var count int64
	startOfMonth := time.Date(time.Now().Year(), time.Now().Month(), 1, 0, 0, 0, 0, time.Local)
	err := d.db.Model(&entity.MoldOrder{}).
		Where("create_time >= ?", startOfMonth).
		Count(&count).Error
	return count, err
}

// GetMonthlyNewMoldOrdersLastMonth 获取上月新增模具订单数量（用于计算增长率）
func (d *DashboardDAO) GetMonthlyNewMoldOrdersLastMonth() (int64, error) {
	var count int64
	now := time.Now()
	startOfLastMonth := time.Date(now.Year(), now.Month()-1, 1, 0, 0, 0, 0, time.Local)
	startOfThisMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.Local)
	err := d.db.Model(&entity.MoldOrder{}).
		Where("create_time >= ? AND create_time < ?", startOfLastMonth, startOfThisMonth).
		Count(&count).Error
	return count, err
}

// GetMonthlyNewProductionOrders 获取本月新增生产订单数量
func (d *DashboardDAO) GetMonthlyNewProductionOrders() (int64, error) {
	var count int64
	startOfMonth := time.Date(time.Now().Year(), time.Now().Month(), 1, 0, 0, 0, 0, time.Local)
	err := d.db.Model(&entity.ProductionOrder{}).
		Where("create_time >= ?", startOfMonth).
		Count(&count).Error
	return count, err
}

// GetMonthlyNewProductionOrdersLastMonth 获取上月新增生产订单数量
func (d *DashboardDAO) GetMonthlyNewProductionOrdersLastMonth() (int64, error) {
	var count int64
	now := time.Now()
	startOfLastMonth := time.Date(now.Year(), now.Month()-1, 1, 0, 0, 0, 0, time.Local)
	startOfThisMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.Local)
	err := d.db.Model(&entity.ProductionOrder{}).
		Where("create_time >= ? AND create_time < ?", startOfLastMonth, startOfThisMonth).
		Count(&count).Error
	return count, err
}

// GetInProductionMolds 获取正在生产流程的模具数量（life_cycle_status = PROCESSING）
func (d *DashboardDAO) GetInProductionMolds() (int64, error) {
	var count int64
	err := d.db.Model(&entity.Mold{}).
		Where("life_cycle_status = ?", constant.MoldStatusProcessing).
		Count(&count).Error
	return count, err
}

// GetInProductionMoldsYesterday 获取昨日正在生产流程的模具数量
func (d *DashboardDAO) GetInProductionMoldsYesterday() (int64, error) {
	var count int64
	yesterday := time.Now().AddDate(0, 0, -1)
	err := d.db.Model(&entity.Mold{}).
		Where("life_cycle_status = ?", constant.MoldStatusProcessing).
		Where("update_time <= ?", yesterday).
		Count(&count).Error
	return count, err
}

// GetMonthlyNewReceivable 获取本月新增应收款总额
func (d *DashboardDAO) GetMonthlyNewReceivable() (float64, error) {
	var total float64
	startOfMonth := time.Date(time.Now().Year(), time.Now().Month(), 1, 0, 0, 0, 0, time.Local)
	err := d.db.Model(&entity.AccountsReceivable{}).
		Select("COALESCE(SUM(amount), 0)").
		Where("create_time >= ?", startOfMonth).
		Scan(&total).Error
	return total, err
}

// GetMonthlyNewReceivableLastMonth 获取上月新增应收款总额
func (d *DashboardDAO) GetMonthlyNewReceivableLastMonth() (float64, error) {
	var total float64
	now := time.Now()
	startOfLastMonth := time.Date(now.Year(), now.Month()-1, 1, 0, 0, 0, 0, time.Local)
	startOfThisMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.Local)
	err := d.db.Model(&entity.AccountsReceivable{}).
		Select("COALESCE(SUM(amount), 0)").
		Where("create_time >= ? AND create_time < ?", startOfLastMonth, startOfThisMonth).
		Scan(&total).Error
	return total, err
}

// GetCustomerOrderRanking 获取客户订单排名
func (d *DashboardDAO) GetCustomerOrderRanking(limit int) ([]map[string]interface{}, error) {
	type Result struct {
		CustomerID   int64  `json:"customer_id"`
		CustomerName string `json:"customer_name"`
		OrderCount   int    `json:"order_count"`
	}

	var results []Result
	err := d.db.Table("master_orders mo").
		Select("mo.customer_id, c.cus_name as customer_name, COUNT(*) as order_count").
		Joins("JOIN customer c ON mo.customer_id = c.id").
		Group("mo.customer_id, c.cus_name").
		Order("order_count DESC").
		Limit(limit).
		Scan(&results).Error

	if err != nil {
		return nil, err
	}

	items := make([]map[string]interface{}, len(results))
	for i, r := range results {
		items[i] = map[string]interface{}{
			"customer_id":   r.CustomerID,
			"customer_name": r.CustomerName,
			"order_count":   r.OrderCount,
		}
	}

	return items, nil
}

type InProgressProductionOrderResult struct {
	OrderID      int64
	OrderCode    string
	CustomerName string
	Status       string
	PlanEndDate  string
}

// GetInProgressProductionOrders 获取进行中的生产订单（固定8条最近到期）
// 包含状态：待排产、已排产、生产中、待质检、已质检（即未完成/未取消的订单）
func (d *DashboardDAO) GetInProgressProductionOrders(limit int) ([]InProgressProductionOrderResult, int64, error) {
	var results []InProgressProductionOrderResult
	var total int64

	// 定义"进行中"状态（排除已入库、已出货、已完成、已取消）
	inProgressStatuses := []string{
		constant.ProductionOrderStatusPendingSchedule,
		constant.ProductionOrderStatusProducing,
	}

	err := d.db.Table("production_order").
		Where("status IN ?", inProgressStatuses).
		Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = d.db.Table("production_order").
		Select(`
			production_order.id as order_id,
			production_order.order_code,
			COALESCE(c.cus_name, '') as customer_name,
			production_order.status,
			COALESCE(production_order.plan_end_date, '') as plan_end_date
		`).
		Joins("LEFT JOIN customer c ON production_order.customer_id = c.id").
		Where("production_order.status IN ?", inProgressStatuses).
		Order("production_order.plan_end_date ASC").
		Limit(limit).
		Scan(&results).Error

	return results, total, err
}

type InProductionMoldListResult struct {
	MoldID            int64
	MoldCode          string
	MoldName          string
	WorkOrderID       int64
	Progress          int
	WorkOrderStatus   string
	LastCompletedStep string
}

// GetInProductionMoldList 获取正在制造的模具列表
func (d *DashboardDAO) GetInProductionMoldList(limit int) ([]InProductionMoldListResult, error) {
	var results []InProductionMoldListResult

	err := d.db.Table("mold").
		Select(`
			mold.id as mold_id,
			mold.mold_code,
			mold.mold_name,
			COALESCE(mwo.id, 0) as work_order_id,
			COALESCE(mwo.progress, 0) as progress,
			COALESCE(mwo.status, '') as work_order_status,
			COALESCE((
				SELECT pt.process_name FROM work_order_process wop
				LEFT JOIN process_template pt ON wop.process_template_id = pt.id
				WHERE wop.work_order_id = mwo.id AND wop.is_completed = 1
				ORDER BY pt.sort DESC LIMIT 1
			), '') as last_completed_step
		`).
		Joins("LEFT JOIN (SELECT MAX(id) as id, mold_id FROM mold_work_order GROUP BY mold_id) latest ON mold.id = latest.mold_id").
		Joins("LEFT JOIN mold_work_order mwo ON latest.id = mwo.id").
		Where("mold.life_cycle_status = ?", constant.MoldStatusProcessing).
		Order("mwo.progress DESC").
		Limit(limit).
		Scan(&results).Error

	return results, err
}

type InProgressMoldOrderResult struct {
	OrderID      int64
	OrderCode    string
	MoldCodes    string
	CustomerName string
	Status       string
	PlanEndDate  string
}

// GetInProgressMoldOrders 获取进行中的模具订单（固定返回8条最近到期的）
// 包含状态：未开始、加工中、待试模（即未试模的订单）
func (d *DashboardDAO) GetInProgressMoldOrders(limit int) ([]InProgressMoldOrderResult, int64, error) {
	var results []InProgressMoldOrderResult
	var total int64

	// 定义"未试模"状态：未开始、加工中、待试模
	inProgressStatuses := []string{
		constant.MoldOrderStatusNotStarted,
		constant.MoldOrderStatusProcessing,
		constant.MoldOrderStatusPendingTest,
	}

	err := d.db.Table("mold_order").
		Where("status IN ?", inProgressStatuses).
		Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	// 模具编号通过 mold_work_order（工单）关联：mold_order → mold_work_order.mold_order_id → mold_work_order.mold_id → mold.mold_code
	err = d.db.Table("mold_order").
		Select(`
			mold_order.id as order_id,
			mold_order.order_code,
			COALESCE(c.cus_name, '') as customer_name,
			mold_order.status,
			COALESCE(mold_order.plan_end_date, '') as plan_end_date,
			COALESCE((
				SELECT GROUP_CONCAT(DISTINCT m.mold_code ORDER BY m.mold_code SEPARATOR ', ')
				FROM mold_work_order mwo
				JOIN mold m ON mwo.mold_id = m.id
				WHERE mwo.mold_order_id = mold_order.id
			), '') as mold_codes
		`).
		Joins("LEFT JOIN customer c ON mold_order.customer_id = c.id").
		Where("mold_order.status IN ?", inProgressStatuses).
		Order("mold_order.plan_end_date ASC").
		Limit(limit).
		Scan(&results).Error

	return results, total, err
}

// =============================================
// 以下为仪表盘财务概览 DAO 方法
// =============================================

// GetFinanceTrend 按日期范围获取应收/应付趋势数据
// 使用日期字符串参数和 DATE()/DATE_FORMAT() 函数，避免 time.Time 的时区转换问题（与 GetProductionTrendByRange 一致）
func (d *DashboardDAO) GetFinanceTrend(startDate, endDate string) ([]dto.FinanceChartItem, error) {
	loc := time.Local
	start, err := time.ParseInLocation("2006-01-02", startDate, loc)
	if err != nil {
		return nil, err
	}
	end, err := time.ParseInLocation("2006-01-02", endDate, loc)
	if err != nil {
		return nil, err
	}

	days := int(end.Sub(start).Hours()/24) + 1
	if days <= 0 {
		days = 1
	}

	items := make([]dto.FinanceChartItem, days)

	type dayAgg struct {
		Day    string
		Amount float64
	}

	// 优化：两次 GROUP BY 查询替代循环逐天查询
	startDateStr := start.Format("2006-01-02")
	endDateStr := end.Format("2006-01-02")

	// 应收趋势
	var receivableResults []dayAgg
	d.db.Model(&entity.AccountsReceivable{}).
		Select("DATE_FORMAT(create_time, '%Y-%m-%d') AS day, COALESCE(SUM(amount), 0) AS amount").
		Where("DATE_FORMAT(create_time, '%Y-%m-%d') >= ? AND DATE_FORMAT(create_time, '%Y-%m-%d') <= ?", startDateStr, endDateStr).
		Group("DATE_FORMAT(create_time, '%Y-%m-%d')").
		Scan(&receivableResults)

	receivableMap := make(map[string]float64, len(receivableResults))
	for _, r := range receivableResults {
		receivableMap[r.Day] = r.Amount
	}

	// 应付趋势
	var payableResults []dayAgg
	d.db.Model(&entity.AccountsPayable{}).
		Select("DATE_FORMAT(create_time, '%Y-%m-%d') AS day, COALESCE(SUM(amount), 0) AS amount").
		Where("DATE_FORMAT(create_time, '%Y-%m-%d') >= ? AND DATE_FORMAT(create_time, '%Y-%m-%d') <= ?", startDateStr, endDateStr).
		Group("DATE_FORMAT(create_time, '%Y-%m-%d')").
		Scan(&payableResults)

	payableMap := make(map[string]float64, len(payableResults))
	for _, r := range payableResults {
		payableMap[r.Day] = r.Amount
	}

	for i := 0; i < days; i++ {
		day := start.AddDate(0, 0, i)
		dayStr := day.Format("2006-01-02")
		items[i] = dto.FinanceChartItem{
			Date:             day.Format("01-02"),
			ReceivableAmount: receivableMap[dayStr],
			PayableAmount:    payableMap[dayStr],
		}
	}

	return items, nil
}

// GetThisMonthReceivableTotal 获取本月应收合计
func (d *DashboardDAO) GetThisMonthReceivableTotal() (float64, error) {
	var total float64
	startOfMonth := time.Date(time.Now().Year(), time.Now().Month(), 1, 0, 0, 0, 0, time.Local)
	err := d.db.Model(&entity.AccountsReceivable{}).
		Select("COALESCE(SUM(amount), 0)").
		Where("create_time >= ?", startOfMonth).
		Scan(&total).Error
	return total, err
}

// GetLastMonthReceivableTotal 获取上月应收合计
func (d *DashboardDAO) GetLastMonthReceivableTotal() (float64, error) {
	var total float64
	now := time.Now()
	startOfLastMonth := time.Date(now.Year(), now.Month()-1, 1, 0, 0, 0, 0, time.Local)
	startOfThisMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.Local)
	err := d.db.Model(&entity.AccountsReceivable{}).
		Select("COALESCE(SUM(amount), 0)").
		Where("create_time >= ? AND create_time < ?", startOfLastMonth, startOfThisMonth).
		Scan(&total).Error
	return total, err
}

// GetThisMonthPayableTotal 获取本月应付合计
func (d *DashboardDAO) GetThisMonthPayableTotal() (float64, error) {
	var total float64
	startOfMonth := time.Date(time.Now().Year(), time.Now().Month(), 1, 0, 0, 0, 0, time.Local)
	err := d.db.Model(&entity.AccountsPayable{}).
		Select("COALESCE(SUM(amount), 0)").
		Where("create_time >= ?", startOfMonth).
		Scan(&total).Error
	return total, err
}

// GetLastMonthPayableTotal 获取上月应付合计
func (d *DashboardDAO) GetLastMonthPayableTotal() (float64, error) {
	var total float64
	now := time.Now()
	startOfLastMonth := time.Date(now.Year(), now.Month()-1, 1, 0, 0, 0, 0, time.Local)
	startOfThisMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.Local)
	err := d.db.Model(&entity.AccountsPayable{}).
		Select("COALESCE(SUM(amount), 0)").
		Where("create_time >= ? AND create_time < ?", startOfLastMonth, startOfThisMonth).
		Scan(&total).Error
	return total, err
}

// GetMoldUsageRanking 获取模具使用排名
func (d *DashboardDAO) GetMoldUsageRanking(limit int) ([]map[string]interface{}, error) {
	type Result struct {
		MoldID     int64  `json:"mold_id"`
		MoldCode   string `json:"mold_code"`
		MoldName   string `json:"mold_name"`
		UsageCount int    `json:"usage_count"`
	}

	var results []Result
	err := d.db.Model(&entity.Mold{}).
		Select("id as mold_id, mold_code, mold_name, total_orders as usage_count").
		Order("total_orders DESC").
		Limit(limit).
		Scan(&results).Error

	if err != nil {
		return nil, err
	}

	items := make([]map[string]interface{}, len(results))
	for i, r := range results {
		items[i] = map[string]interface{}{
			"mold_id":     r.MoldID,
			"mold_code":   r.MoldCode,
			"mold_name":   r.MoldName,
			"usage_count": r.UsageCount,
		}
	}

	return items, nil
}

// =============================================
// 统计报表页概览 DAO 方法
// =============================================

// GetMonthlyCompletedProductionOrders 获取本月完成的生产订单数
func (d *DashboardDAO) GetMonthlyCompletedProductionOrders() (int64, error) {
	var count int64
	startOfMonth := time.Date(time.Now().Year(), time.Now().Month(), 1, 0, 0, 0, 0, time.Local)
	startOfNextMonth := startOfMonth.AddDate(0, 1, 0)
	err := d.db.Model(&entity.ProductionOrder{}).
		Where("status = ?", constant.ProductionOrderStatusCompleted).
		Where("actual_end_date >= ? AND actual_end_date < ?", startOfMonth, startOfNextMonth).
		Count(&count).Error
	return count, err
}

// GetLastMonthCompletedProductionOrders 获取上月完成的生产订单数
func (d *DashboardDAO) GetLastMonthCompletedProductionOrders() (int64, error) {
	var count int64
	now := time.Now()
	startOfLastMonth := time.Date(now.Year(), now.Month()-1, 1, 0, 0, 0, 0, time.Local)
	startOfThisMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.Local)
	err := d.db.Model(&entity.ProductionOrder{}).
		Where("status = ?", constant.ProductionOrderStatusCompleted).
		Where("actual_end_date >= ? AND actual_end_date < ?", startOfLastMonth, startOfThisMonth).
		Count(&count).Error
	return count, err
}

// GetMonthlyStockInTotal 获取本月入库总量
func (d *DashboardDAO) GetMonthlyStockInTotal() (int64, error) {
	var total int64
	startOfMonth := time.Date(time.Now().Year(), time.Now().Month(), 1, 0, 0, 0, 0, time.Local)
	startOfNextMonth := startOfMonth.AddDate(0, 1, 0)
	err := d.db.Model(&entity.ProductionStockIn{}).
		Select("COALESCE(SUM(quantity), 0)").
		Where("stock_in_date >= ? AND stock_in_date < ?", startOfMonth, startOfNextMonth).
		Scan(&total).Error
	return total, err
}

// GetLastMonthStockInTotal 获取上月入库总量
func (d *DashboardDAO) GetLastMonthStockInTotal() (int64, error) {
	var total int64
	now := time.Now()
	startOfLastMonth := time.Date(now.Year(), now.Month()-1, 1, 0, 0, 0, 0, time.Local)
	startOfThisMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.Local)
	err := d.db.Model(&entity.ProductionStockIn{}).
		Select("COALESCE(SUM(quantity), 0)").
		Where("stock_in_date >= ? AND stock_in_date < ?", startOfLastMonth, startOfThisMonth).
		Scan(&total).Error
	return total, err
}

// GetMonthlyStockOutTotal 获取本月出库总量
func (d *DashboardDAO) GetMonthlyStockOutTotal() (int64, error) {
	var total int64
	startOfMonth := time.Date(time.Now().Year(), time.Now().Month(), 1, 0, 0, 0, 0, time.Local)
	startOfNextMonth := startOfMonth.AddDate(0, 1, 0)
	err := d.db.Model(&entity.ProductionStockOut{}).
		Select("COALESCE(SUM(quantity), 0)").
		Where("stock_out_date >= ? AND stock_out_date < ?", startOfMonth, startOfNextMonth).
		Scan(&total).Error
	return total, err
}

// GetLastMonthStockOutTotal 获取上月出库总量
func (d *DashboardDAO) GetLastMonthStockOutTotal() (int64, error) {
	var total int64
	now := time.Now()
	startOfLastMonth := time.Date(now.Year(), now.Month()-1, 1, 0, 0, 0, 0, time.Local)
	startOfThisMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.Local)
	err := d.db.Model(&entity.ProductionStockOut{}).
		Select("COALESCE(SUM(quantity), 0)").
		Where("stock_out_date >= ? AND stock_out_date < ?", startOfLastMonth, startOfThisMonth).
		Scan(&total).Error
	return total, err
}

// GetMonthlyProductionReportRatios 获取本月和上月的报工良品率
func (d *DashboardDAO) GetMonthlyProductionReportRatios() (currentRate, lastRate float64, err error) {
	now := time.Now()
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.Local)
	startOfNextMonth := startOfMonth.AddDate(0, 1, 0)
	startOfLastMonth := time.Date(now.Year(), now.Month()-1, 1, 0, 0, 0, 0, time.Local)

	type rateResult struct {
		OkQty  int64
		BadQty int64
	}

	// 本月
	var thisMonth rateResult
	d.db.Model(&entity.ProductionReport{}).
		Select("COALESCE(SUM(ok_qty), 0) AS ok_qty, COALESCE(SUM(bad_qty), 0) AS bad_qty").
		Where("report_date >= ? AND report_date < ?", startOfMonth, startOfNextMonth).
		Scan(&thisMonth)

	totalThis := thisMonth.OkQty + thisMonth.BadQty
	if totalThis > 0 {
		currentRate = float64(thisMonth.OkQty) / float64(totalThis) * 100
	}

	// 上月
	var lastMonth rateResult
	d.db.Model(&entity.ProductionReport{}).
		Select("COALESCE(SUM(ok_qty), 0) AS ok_qty, COALESCE(SUM(bad_qty), 0) AS bad_qty").
		Where("report_date >= ? AND report_date < ?", startOfLastMonth, startOfMonth).
		Scan(&lastMonth)

	totalLast := lastMonth.OkQty + lastMonth.BadQty
	if totalLast > 0 {
		lastRate = float64(lastMonth.OkQty) / float64(totalLast) * 100
	}

	return currentRate, lastRate, nil
}

// GetInProgressMoldWorkOrderCount 获取进行中的模具工单数
func (d *DashboardDAO) GetInProgressMoldWorkOrderCount() (int64, error) {
	var count int64
	notFinished := []string{
		constant.WorkOrderStatusNotStarted,
		constant.WorkOrderStatusProcessing,
		constant.WorkOrderStatusTested,
	}
	err := d.db.Model(&entity.MoldWorkOrder{}).
		Where("status IN ?", notFinished).
		Where("is_cancelled = 0").
		Count(&count).Error
	return count, err
}

// GetLastMonthInProgressMoldWorkOrderCount 获取上月同期的进行中模具工单数
func (d *DashboardDAO) GetLastMonthInProgressMoldWorkOrderCount() (int64, error) {
	var count int64
	notFinished := []string{
		constant.WorkOrderStatusNotStarted,
		constant.WorkOrderStatusProcessing,
		constant.WorkOrderStatusTested,
	}
	now := time.Now()
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.Local)

	err := d.db.Model(&entity.MoldWorkOrder{}).
		Where("status IN ?", notFinished).
		Where("is_cancelled = 0").
		Where("create_time < ?", startOfMonth).
		Count(&count).Error
	return count, err
}

// GetMonthlyCompletedMoldWorkOrders 获取本月完成的模具工单数
func (d *DashboardDAO) GetMonthlyCompletedMoldWorkOrders() (int64, error) {
	var count int64
	startOfMonth := time.Date(time.Now().Year(), time.Now().Month(), 1, 0, 0, 0, 0, time.Local)
	startOfNextMonth := startOfMonth.AddDate(0, 1, 0)
	err := d.db.Model(&entity.MoldWorkOrder{}).
		Where("status IN ?", []string{constant.WorkOrderStatusCompleted, constant.WorkOrderStatusStockedIn, constant.WorkOrderStatusDelivered}).
		Where("actual_end IS NOT NULL").
		Where("actual_end >= ? AND actual_end < ?", startOfMonth, startOfNextMonth).
		Where("is_cancelled = 0").
		Count(&count).Error
	return count, err
}

// GetLastMonthCompletedMoldWorkOrders 获取上月完成的模具工单数
func (d *DashboardDAO) GetLastMonthCompletedMoldWorkOrders() (int64, error) {
	var count int64
	now := time.Now()
	startOfLastMonth := time.Date(now.Year(), now.Month()-1, 1, 0, 0, 0, 0, time.Local)
	startOfThisMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.Local)
	err := d.db.Model(&entity.MoldWorkOrder{}).
		Where("status IN ?", []string{constant.WorkOrderStatusCompleted, constant.WorkOrderStatusStockedIn, constant.WorkOrderStatusDelivered}).
		Where("actual_end IS NOT NULL").
		Where("actual_end >= ? AND actual_end < ?", startOfLastMonth, startOfThisMonth).
		Where("is_cancelled = 0").
		Count(&count).Error
	return count, err
}

// GetInProgressProductionOrderCount 获取进行中的生产订单数（待排产/生产中）
func (d *DashboardDAO) GetInProgressProductionOrderCount() (int64, error) {
	var count int64
	inProgressStatuses := []string{
		constant.ProductionOrderStatusPendingSchedule,
		constant.ProductionOrderStatusProducing,
	}
	err := d.db.Model(&entity.ProductionOrder{}).
		Where("status IN ?", inProgressStatuses).
		Count(&count).Error
	return count, err
}

// GetLastMonthInProgressProductionOrderCount 获取上月同期的进行中生产订单数
func (d *DashboardDAO) GetLastMonthInProgressProductionOrderCount() (int64, error) {
	var count int64
	now := time.Now()
	startOfLastMonth := time.Date(now.Year(), now.Month()-1, 1, 0, 0, 0, 0, time.Local)
	startOfThisMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.Local)
	inProgressStatuses := []string{
		constant.ProductionOrderStatusPendingSchedule,
		constant.ProductionOrderStatusProducing,
	}
	// 上月同期：统计上月1日至当天的进行中生产订单
	lastMonthStart := startOfLastMonth
	lastMonthEnd := startOfThisMonth
	err := d.db.Model(&entity.ProductionOrder{}).
		Where("status IN ?", inProgressStatuses).
		Where("create_time >= ? AND create_time < ?", lastMonthStart, lastMonthEnd).
		Count(&count).Error
	return count, err
}

// GetMonthlyDetailProductionTrend 获取统计报表页的近12个月月度生产趋势数据
func (d *DashboardDAO) GetMonthlyDetailProductionTrend() ([]dto.DetailProductionTrendItem, error) {
	now := time.Now()
	startOfPeriod := time.Date(now.Year()-1, now.Month()+1, 1, 0, 0, 0, 0, time.Local) // 12个月前
	endOfPeriod := time.Date(now.Year(), now.Month()+1, 1, 0, 0, 0, 0, time.Local)      // 下个月1号

	months := 12
	dateMap := make(map[string]*dto.DetailProductionTrendItem)
	for i := 0; i < months; i++ {
		m := now.AddDate(0, -11+i, 0)
		key := m.Format("2006-01")
		dateMap[key] = &dto.DetailProductionTrendItem{Date: key}
	}

	type monthCount struct {
		Month string
		Total int64
	}

	// 1. 月度完成的生产订单数
	var completedOrderResults []monthCount
	if err := d.db.Model(&entity.ProductionOrder{}).
		Select("DATE_FORMAT(actual_end_date, '%Y-%m') AS month, COUNT(*) AS total").
		Where("status = ?", constant.ProductionOrderStatusCompleted).
		Where("actual_end_date >= ? AND actual_end_date < ?", startOfPeriod, endOfPeriod).
		Group("DATE_FORMAT(actual_end_date, '%Y-%m')").
		Scan(&completedOrderResults).Error; err != nil {
		return nil, err
	}
	for _, r := range completedOrderResults {
		if item, ok := dateMap[r.Month]; ok {
			item.CompletedProductionOrders = int(r.Total)
		}
	}

	// 2. 月度新增的进行中生产订单数
	var inProgressOrderResults []monthCount
	if err := d.db.Model(&entity.ProductionOrder{}).
		Select("DATE_FORMAT(create_time, '%Y-%m') AS month, COUNT(*) AS total").
		Where("status IN ?", []string{constant.ProductionOrderStatusPendingSchedule, constant.ProductionOrderStatusProducing}).
		Where("create_time >= ? AND create_time < ?", startOfPeriod, endOfPeriod).
		Group("DATE_FORMAT(create_time, '%Y-%m')").
		Scan(&inProgressOrderResults).Error; err != nil {
		return nil, err
	}
	for _, r := range inProgressOrderResults {
		if item, ok := dateMap[r.Month]; ok {
			item.InProgressProductionOrders = int(r.Total)
		}
	}

	// 3. 月度模具完工数
	completedStatuses := []string{
		constant.WorkOrderStatusCompleted,
		constant.WorkOrderStatusStockedIn,
		constant.WorkOrderStatusDelivered,
	}
	var completedMoldResults []monthCount
	if err := d.db.Model(&entity.MoldWorkOrder{}).
		Select("DATE_FORMAT(actual_end, '%Y-%m') AS month, COUNT(*) AS total").
		Where("status IN ?", completedStatuses).
		Where("actual_end >= ? AND actual_end < ?", startOfPeriod, endOfPeriod).
		Where("is_cancelled = 0").
		Group("DATE_FORMAT(actual_end, '%Y-%m')").
		Scan(&completedMoldResults).Error; err != nil {
		return nil, err
	}
	for _, r := range completedMoldResults {
		if item, ok := dateMap[r.Month]; ok {
			item.CompletedMoldWorkOrders = int(r.Total)
		}
	}

	// 4. 月度新增的进行中模具工单数
	inProgressStatuses := []string{
		constant.WorkOrderStatusNotStarted,
		constant.WorkOrderStatusProcessing,
		constant.WorkOrderStatusTested,
	}
	var inProgressMoldResults []monthCount
	if err := d.db.Model(&entity.MoldWorkOrder{}).
		Select("DATE_FORMAT(create_time, '%Y-%m') AS month, COUNT(*) AS total").
		Where("status IN ?", inProgressStatuses).
		Where("create_time >= ? AND create_time < ?", startOfPeriod, endOfPeriod).
		Where("is_cancelled = 0").
		Group("DATE_FORMAT(create_time, '%Y-%m')").
		Scan(&inProgressMoldResults).Error; err != nil {
		return nil, err
	}
	for _, r := range inProgressMoldResults {
		if item, ok := dateMap[r.Month]; ok {
			item.InProgressMoldWorkOrders = int(r.Total)
		}
	}

	// 按月份排序输出
	result := make([]dto.DetailProductionTrendItem, 0, months)
	for i := 0; i < months; i++ {
		m := now.AddDate(0, -11+i, 0)
		key := m.Format("2006-01")
		if item, ok := dateMap[key]; ok {
			result = append(result, *item)
		}
	}
	return result, nil
}

// GetMonthlyDeliveryPunctuality 获取近12个月生产订单交付准时率数据
func (d *DashboardDAO) GetMonthlyDeliveryPunctuality() ([]dto.DeliveryPunctualityItem, error) {
	now := time.Now()
	startOfPeriod := time.Date(now.Year()-1, now.Month()+1, 1, 0, 0, 0, 0, time.Local)
	endOfPeriod := time.Date(now.Year(), now.Month()+1, 1, 0, 0, 0, 0, time.Local)

	months := 12
	dateMap := make(map[string]*dto.DeliveryPunctualityItem)
	for i := 0; i < months; i++ {
		m := now.AddDate(0, -11+i, 0)
		key := m.Format("2006-01")
		dateMap[key] = &dto.DeliveryPunctualityItem{Month: key}
	}

	type monthStatusCount struct {
		Month    string
		Category string // early / ontime / delayed
		Total    int64
	}

	// 查询当月已完成订单的交付准时情况
	var results []monthStatusCount
	d.db.Model(&entity.ProductionOrder{}).
		Select(`DATE_FORMAT(actual_end_date, '%Y-%m') AS month,
			CASE
				WHEN actual_end_date < plan_end_date THEN 'early'
				WHEN actual_end_date = plan_end_date THEN 'ontime'
				WHEN actual_end_date > plan_end_date THEN 'delayed'
				ELSE 'delayed'
			END AS category,
			COUNT(*) AS total`).
		Where("actual_end_date IS NOT NULL").
		Where("plan_end_date IS NOT NULL").
		Where("actual_end_date >= ? AND actual_end_date < ?", startOfPeriod, endOfPeriod).
		Group("DATE_FORMAT(actual_end_date, '%Y-%m'), category").
		Scan(&results)

	for _, r := range results {
		if item, ok := dateMap[r.Month]; ok {
			switch r.Category {
			case "early":
				item.Early = int(r.Total)
			case "ontime":
				item.OnTime = int(r.Total)
			case "delayed":
				item.Delayed = int(r.Total)
			}
		}
	}

	// 按月份排序输出
	result := make([]dto.DeliveryPunctualityItem, 0, months)
	for i := 0; i < months; i++ {
		m := now.AddDate(0, -11+i, 0)
		key := m.Format("2006-01")
		if item, ok := dateMap[key]; ok {
			result = append(result, *item)
		}
	}
	return result, nil
}
