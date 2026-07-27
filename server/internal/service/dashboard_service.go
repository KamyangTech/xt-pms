package service

import (
	"context"
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/KamyangTech/xt-pms/server/internal/dao"
	"github.com/KamyangTech/xt-pms/server/internal/model/constant"
	"github.com/KamyangTech/xt-pms/server/internal/model/dto"
	"github.com/KamyangTech/xt-pms/server/internal/model/entity"
	"github.com/KamyangTech/xt-pms/server/pkg/logger"
	"go.uber.org/zap"
)

type DashboardService struct {
	dao *dao.DashboardDAO
}

// NewDashboardService 创建 DashboardService 实例
func NewDashboardService() *DashboardService {
	return &DashboardService{
		dao: dao.NewDashboardDAO(),
	}
}

// GetOverview 获取概览数据
func (s *DashboardService) GetOverview() (*dto.DashboardOverview, error) {
	// 设置超时上下文，防止 goroutine 泄漏
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 并发获取所有数据
	type result struct {
		value interface{}
		err   error
		index int
	}

	var (
		totalMolds, customers, suppliers, stockWarning, pendingOrders, todayOutput, yesterdayOutput, rawMaterialStock int64
		yesterdayMolds, yesterdayStockWarning, yesterdayPendingOrders int64
		yesterdayCustomers, yesterdaySuppliers, yesterdayRawMaterialStock int64
		qualificationRate, yesterdayQualificationRate float64
		monthlyNewMoldOrders, monthlyNewProductionOrders, inProductionMolds int64
		monthlyNewMoldOrdersLastMonth, monthlyNewProductionOrdersLastMonth, inProductionMoldsYesterday int64
		monthlyNewReceivable, monthlyNewReceivableLastMonth float64
	)

	var wg sync.WaitGroup
	resultChan := make(chan result, 15)

	wg.Add(15)

	// 【优化】使用合并查询方法，一次 SQL 返回当前值和昨日值

	// 模具相关（总数+昨日）
	go func() {
		defer wg.Done()
		total, yesterday, err := s.dao.GetMoldCounts()
		select { case resultChan <- result{value: []int64{total, yesterday}, err: err, index: 0}: case <-ctx.Done(): }
	}()
	// 客户相关
	go func() {
		defer wg.Done()
		total, yesterday, err := s.dao.GetCustomerCounts()
		select { case resultChan <- result{value: []int64{total, yesterday}, err: err, index: 1}: case <-ctx.Done(): }
	}()
	// 供应商相关
	go func() {
		defer wg.Done()
		total, yesterday, err := s.dao.GetSupplierCounts()
		select { case resultChan <- result{value: []int64{total, yesterday}, err: err, index: 2}: case <-ctx.Done(): }
	}()
	// 库存预警相关
	go func() {
		defer wg.Done()
		total, yesterday, err := s.dao.GetStockWarningCounts()
		select { case resultChan <- result{value: []int64{total, yesterday}, err: err, index: 3}: case <-ctx.Done(): }
	}()
	// 待办工单相关
	go func() {
		defer wg.Done()
		total, yesterday, err := s.dao.GetPendingWorkOrderCounts()
		select { case resultChan <- result{value: []int64{total, yesterday}, err: err, index: 4}: case <-ctx.Done(): }
	}()
	// 今日/昨日产量
	go func() {
		defer wg.Done()
		today, yesterday, err := s.dao.GetProductionOutputs()
		select { case resultChan <- result{value: []int64{today, yesterday}, err: err, index: 5}: case <-ctx.Done(): }
	}()
	// 原料库存相关
	go func() {
		defer wg.Done()
		total, yesterday, err := s.dao.GetRawMaterialStockCounts()
		select { case resultChan <- result{value: []int64{total, yesterday}, err: err, index: 6}: case <-ctx.Done(): }
	}()
	// 合格率相关（返回 float64 对）
	go func() {
		defer wg.Done()
		total, yesterday, err := s.dao.GetQualificationRates()
		select { case resultChan <- result{value: []float64{total, yesterday}, err: err, index: 7}: case <-ctx.Done(): }
	}()
	// 生产中模具相关
	go func() {
		defer wg.Done()
		total, yesterday, err := s.dao.GetInProductionMoldCounts()
		select { case resultChan <- result{value: []int64{total, yesterday}, err: err, index: 8}: case <-ctx.Done(): }
	}()
	// 本月/上月新模具订单
	go func() {
		defer wg.Done()
		thisMonth, err := s.dao.GetMonthlyNewMoldOrders()
		select { case resultChan <- result{value: thisMonth, err: err, index: 9}: case <-ctx.Done(): }
	}()
	go func() {
		defer wg.Done()
		lastMonth, err := s.dao.GetMonthlyNewMoldOrdersLastMonth()
		select { case resultChan <- result{value: lastMonth, err: err, index: 10}: case <-ctx.Done(): }
	}()
	// 本月/上月新生订单
	go func() {
		defer wg.Done()
		thisMonth, err := s.dao.GetMonthlyNewProductionOrders()
		select { case resultChan <- result{value: thisMonth, err: err, index: 11}: case <-ctx.Done(): }
	}()
	go func() {
		defer wg.Done()
		lastMonth, err := s.dao.GetMonthlyNewProductionOrdersLastMonth()
		select { case resultChan <- result{value: lastMonth, err: err, index: 12}: case <-ctx.Done(): }
	}()
	// 本月/上月新应收
	go func() {
		defer wg.Done()
		thisMonth, err := s.dao.GetMonthlyNewReceivable()
		select { case resultChan <- result{value: thisMonth, err: err, index: 13}: case <-ctx.Done(): }
	}()
	go func() {
		defer wg.Done()
		lastMonth, err := s.dao.GetMonthlyNewReceivableLastMonth()
		select { case resultChan <- result{value: lastMonth, err: err, index: 14}: case <-ctx.Done(): }
	}()

	// 在后台等待所有 goroutine 完成并关闭 channel
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	for r := range resultChan {
		if r.err != nil {
			continue
		}
		switch v := r.value.(type) {
		case []int64:
			switch r.index {
			case 0: // 模具
				totalMolds = v[0]
				yesterdayMolds = v[1]
			case 1: // 客户
				customers = v[0]
				yesterdayCustomers = v[1]
			case 2: // 供应商
				suppliers = v[0]
				yesterdaySuppliers = v[1]
			case 3: // 库存预警
				stockWarning = v[0]
				yesterdayStockWarning = v[1]
			case 4: // 待办工单
				pendingOrders = v[0]
				yesterdayPendingOrders = v[1]
			case 5: // 产量
				todayOutput = v[0]
				yesterdayOutput = v[1]
			case 6: // 原料库存
				rawMaterialStock = v[0]
				yesterdayRawMaterialStock = v[1]
			case 8: // 生产中模具
				inProductionMolds = v[0]
				inProductionMoldsYesterday = v[1]
			}
		case []float64:
			switch r.index {
			case 7: // 合格率
				qualificationRate = v[0]
				yesterdayQualificationRate = v[1]
			}
		case int64:
			switch r.index {
			case 9:
				monthlyNewMoldOrders = v
			case 10:
				monthlyNewMoldOrdersLastMonth = v
			case 11:
				monthlyNewProductionOrders = v
			case 12:
				monthlyNewProductionOrdersLastMonth = v
			}
		case float64:
			switch r.index {
			case 13:
				monthlyNewReceivable = v
			case 14:
				monthlyNewReceivableLastMonth = v
			}
		}
	}

	todayOutputGrowth := calculateGrowth(todayOutput, yesterdayOutput)
	moldsGrowth := calculateGrowth(totalMolds, yesterdayMolds)
	stockWarningChange := calculateGrowth(stockWarning, yesterdayStockWarning)
	workOrdersGrowth := calculateGrowth(pendingOrders, yesterdayPendingOrders)
	customersGrowth := calculateGrowth(customers, yesterdayCustomers)
	suppliersGrowth := calculateGrowth(suppliers, yesterdaySuppliers)
	rawMaterialGrowth := calculateGrowth(rawMaterialStock, yesterdayRawMaterialStock)
	qualificationGrowth := calculateGrowthFloat(qualificationRate, yesterdayQualificationRate)

	// 新卡片增长率
	monthlyNewMoldOrdersGrowth := calculateGrowth(monthlyNewMoldOrders, monthlyNewMoldOrdersLastMonth)
	monthlyNewProductionOrdersGrowth := calculateGrowth(monthlyNewProductionOrders, monthlyNewProductionOrdersLastMonth)
	inProductionMoldsGrowth := calculateGrowth(inProductionMolds, inProductionMoldsYesterday)
	monthlyNewReceivableGrowth := calculateGrowthFloat(monthlyNewReceivable, monthlyNewReceivableLastMonth)

	overview := &dto.DashboardOverview{
		// 顶部四张卡片
		MonthlyNewMoldOrders:             int(monthlyNewMoldOrders),
		MonthlyNewMoldOrdersGrowth:       monthlyNewMoldOrdersGrowth,
		MonthlyNewProductionOrders:       int(monthlyNewProductionOrders),
		MonthlyNewProductionOrdersGrowth: monthlyNewProductionOrdersGrowth,
		InProductionMolds:                int(inProductionMolds),
		InProductionMoldsGrowth:          inProductionMoldsGrowth,
		MonthlyNewReceivable:             monthlyNewReceivable,
		MonthlyNewReceivableGrowth:       monthlyNewReceivableGrowth,

		// 原有字段保持兼容
		TodayOutput:        int(todayOutput),
		TodayOutputGrowth:  todayOutputGrowth,
		TotalMolds:         int(totalMolds),
		MoldsGrowth:        moldsGrowth,
		StockWarning:       int(stockWarning),
		StockWarningChange: stockWarningChange,
		RawMaterialStock:   int(rawMaterialStock),
		RawMaterialGrowth:  rawMaterialGrowth,
		PendingWorkOrders:  int(pendingOrders),
		WorkOrdersGrowth:   workOrdersGrowth,
		TotalCustomers:     int(customers),
		CustomersGrowth:    customersGrowth,
		TotalSuppliers:     int(suppliers),
		SuppliersGrowth:    suppliersGrowth,
		QualificationRate:   qualificationRate,
		QualificationGrowth: qualificationGrowth,
	}

	return overview, nil
}

// GetFinanceOverview 获取仪表盘财务概览数据（按日期范围）
func (s *DashboardService) GetFinanceOverview(startDate, endDate string) (*dto.FinanceOverviewDashboard, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	type result struct {
		value interface{}
		err   error
		index int
	}

	var (
		chartData                  []dto.FinanceChartItem
		thisMonthReceivable        float64
		lastMonthReceivable        float64
		thisMonthPayable           float64
		lastMonthPayable           float64
	)

	var wg sync.WaitGroup
	resultChan := make(chan result, 5)

	wg.Add(5)

	go func() {
		defer wg.Done()
		data, err := s.dao.GetFinanceTrend(startDate, endDate)
		select { case resultChan <- result{value: data, err: err, index: 0}: case <-ctx.Done(): }
	}()
	go func() {
		defer wg.Done()
		total, err := s.dao.GetThisMonthReceivableTotal()
		select { case resultChan <- result{value: total, err: err, index: 1}: case <-ctx.Done(): }
	}()
	go func() {
		defer wg.Done()
		total, err := s.dao.GetLastMonthReceivableTotal()
		select { case resultChan <- result{value: total, err: err, index: 2}: case <-ctx.Done(): }
	}()
	go func() {
		defer wg.Done()
		total, err := s.dao.GetThisMonthPayableTotal()
		select { case resultChan <- result{value: total, err: err, index: 3}: case <-ctx.Done(): }
	}()
	go func() {
		defer wg.Done()
		total, err := s.dao.GetLastMonthPayableTotal()
		select { case resultChan <- result{value: total, err: err, index: 4}: case <-ctx.Done(): }
	}()

	go func() {
		wg.Wait()
		close(resultChan)
	}()

	for r := range resultChan {
		if r.err != nil {
			continue
		}
		switch v := r.value.(type) {
		case []dto.FinanceChartItem:
			if r.index == 0 {
				chartData = v
			}
		case float64:
			switch r.index {
			case 1:
				thisMonthReceivable = v
			case 2:
				lastMonthReceivable = v
			case 3:
				thisMonthPayable = v
			case 4:
				lastMonthPayable = v
			}
		}
	}

	if chartData == nil {
		chartData = []dto.FinanceChartItem{}
	}

	overview := &dto.FinanceOverviewDashboard{
		ChartData:                   chartData,
		ThisMonthReceivable:         thisMonthReceivable,
		ThisMonthReceivableGrowth:   calculateGrowthFloat(thisMonthReceivable, lastMonthReceivable),
		ThisMonthPayable:            thisMonthPayable,
		ThisMonthPayableGrowth:      calculateGrowthFloat(thisMonthPayable, lastMonthPayable),
	}

	return overview, nil
}

// GetProductionTrend 获取生产趋势数据
func (s *DashboardService) GetProductionTrend(period string, startDate, endDate string) (*dto.ProductionTrend, error) {
	var dates []string
	var currentData, lastPeriodData []float64
	var err error

	// 如果传入了日期范围，按日期范围查询（单条线）
	if startDate != "" && endDate != "" {
		start, err1 := time.ParseInLocation("2006-01-02", startDate, time.Local)
		end, err2 := time.ParseInLocation("2006-01-02", endDate, time.Local)
		if err1 == nil && err2 == nil && !end.Before(start) {
			dates, currentData, lastPeriodData, err = s.dao.GetProductionTrendByRange(start, end)
		} else {
			err = s.fallbackProductionTrend(period, &dates, &currentData, &lastPeriodData)
		}
	} else {
		err = s.fallbackProductionTrend(period, &dates, &currentData, &lastPeriodData)
	}

	if err != nil {
		return nil, err
	}

	// 有上期数据时返回两条线，否则只返回一条线
	datasets := []dto.TrendDataset{
		{
			Name: "日产量",
			Data: currentData,
		},
	}
	if lastPeriodData != nil {
		datasets = append(datasets, dto.TrendDataset{
			Name: "上期",
			Data: lastPeriodData,
		})
	}

	return &dto.ProductionTrend{
		Categories: dates,
		Datasets:   datasets,
	}, nil
}

// fallbackProductionTrend 降级为按周期（week/month/year）查询
func (s *DashboardService) fallbackProductionTrend(period string, dates *[]string, currentData, lastPeriodData *[]float64) error {
	var err error
	switch period {
	case "week":
		*dates, *currentData, *lastPeriodData, err = s.dao.GetProductionTrendByWeek()
	case "year":
		*dates, *currentData, *lastPeriodData, err = s.dao.GetProductionTrendByYear()
	default:
		*dates, *currentData, *lastPeriodData, err = s.dao.GetProductionTrendByMonth()
	}
	return err
}

// GetMoldStatus 获取模具状态统计
func (s *DashboardService) GetMoldStatus() (*dto.MoldStatus, error) {
	// 优化：单次 GROUP BY 查询替代 8 次独立查询
	type statusCount struct {
		LifeCycleStatus string
		Count           int64
	}
	var results []statusCount
	dao.GetDB().Model(&entity.Mold{}).
		Select("life_cycle_status, COUNT(*) as count").
		Group("life_cycle_status").
		Scan(&results)

	var total, inUse, idle, maintaining, scrap int64
	for _, r := range results {
		total += r.Count
		switch r.LifeCycleStatus {
		// 使用中/生产中：包括 PRODUCING、DELIVERED、TESTED
		case constant.MoldStatusProducing, constant.MoldStatusDelivered, constant.MoldStatusTested:
			inUse += r.Count
		// 闲置/库存
		case constant.MoldStatusInStock:
			idle = r.Count
		// 维修/维护中
		case constant.MoldStatusRepairing:
			maintaining = r.Count
		// 报废
		case constant.MoldStatusScrapped:
			scrap = r.Count
		}
	}

	logger.Info("模具状态统计",
		zap.Int64("total", total),
		zap.Int64("inUse", inUse),
		zap.Int64("idle", idle),
		zap.Int64("maintaining", maintaining),
		zap.Int64("scrap", scrap))

	return &dto.MoldStatus{
		Total:       int(total),
		InUse:       int(inUse),
		Idle:        int(idle),
		Maintenance: int(maintaining),
		Scrap:       int(scrap),
	}, nil
}

var productionOrderStatusLabels = map[string]string{
	constant.ProductionOrderStatusPendingSchedule: "待排产",
	constant.ProductionOrderStatusProducing:       "生产中",
	constant.ProductionOrderStatusStockedIn:       "已入库",
	constant.ProductionOrderStatusShipped:         "已出货",
	constant.ProductionOrderStatusCompleted:       "已完成",
	constant.ProductionOrderStatusCancelled:       "已取消",
}

var moldOrderStatusLabels = map[string]string{
	constant.MoldOrderStatusNotStarted:  "未开始",
	constant.MoldOrderStatusProcessing:  "加工中",
	constant.MoldOrderStatusPendingTest: "待试模",
	constant.MoldOrderStatusTested:      "已试模",
	constant.MoldOrderStatusCompleted:   "已完成",
	constant.MoldOrderStatusCancelled:   "已取消",
	constant.MoldOrderStatusDelivered:   "已交付",
}

// GetInProgressProductionOrders 获取进行中的生产订单列表（固定8条最近到期）
func (s *DashboardService) GetInProgressProductionOrders(limit int) (*dto.InProgressProductionOrders, error) {
	results, total, err := s.dao.GetInProgressProductionOrders(limit)
	if err != nil {
		return nil, err
	}

	items := make([]dto.InProgressProductionOrderItem, len(results))
	for i, r := range results {
		label := productionOrderStatusLabels[r.Status]
		if label == "" {
			label = r.Status
		}
		items[i] = dto.InProgressProductionOrderItem{
			OrderID:      r.OrderID,
			OrderCode:    r.OrderCode,
			CustomerName: r.CustomerName,
			Status:       r.Status,
			StatusLabel:  label,
			PlanEndDate:  r.PlanEndDate,
		}
	}

	return &dto.InProgressProductionOrders{
		Items: items,
		Total: total,
	}, nil
}

// GetInProgressMoldOrders 获取进行中的模具订单列表（固定8条最近到期）
func (s *DashboardService) GetInProgressMoldOrders(limit int) (*dto.InProgressMoldOrders, error) {
	results, total, err := s.dao.GetInProgressMoldOrders(limit)
	if err != nil {
		return nil, err
	}

	items := make([]dto.InProgressMoldOrderItem, len(results))
	for i, r := range results {
		label := moldOrderStatusLabels[r.Status]
		if label == "" {
			label = r.Status
		}
		items[i] = dto.InProgressMoldOrderItem{
			OrderID:      r.OrderID,
			OrderCode:    r.OrderCode,
			MoldCodes:    r.MoldCodes,
			CustomerName: r.CustomerName,
			Status:       r.Status,
			StatusLabel:  label,
			PlanEndDate:  r.PlanEndDate,
		}
	}

	return &dto.InProgressMoldOrders{
		Items:    items,
		Total:    total,
		Page:     1,
		PageSize: limit,
	}, nil
}

// GetInProductionMoldList 获取正在制造的模具列表
func (s *DashboardService) GetInProductionMoldList(limit int) (*dto.InProductionMolds, error) {
	results, err := s.dao.GetInProductionMoldList(limit)
	if err != nil {
		return nil, err
	}
	items := make([]dto.InProductionMold, len(results))
	for i, r := range results {
		items[i] = dto.InProductionMold{
			MoldID:            r.MoldID,
			MoldCode:          r.MoldCode,
			MoldName:          r.MoldName,
			WorkOrderID:       r.WorkOrderID,
			Progress:          r.Progress,
			WorkOrderStatus:   r.WorkOrderStatus,
			LastCompletedStep: r.LastCompletedStep,
		}
	}
	return &dto.InProductionMolds{Items: items}, nil
}

// GetOrderRanking 获取订单排名
func (s *DashboardService) GetOrderRanking(rankingType string, limit int) (*dto.OrderRanking, error) {
	var items []dto.RankingItem

	switch rankingType {
	case "customer":
		customerData, err := s.dao.GetCustomerOrderRanking(limit)
		if err != nil {
			return nil, err
		}
		for i, data := range customerData {
			items = append(items, dto.RankingItem{
				Rank:       i + 1,
				Name:       data["customer_name"].(string),
				Change:     0, // 暂时设为 0，后续可以计算环比变化
				OrderCount: data["order_count"].(int),
			})
		}
	case "mold":
		moldData, err := s.dao.GetMoldUsageRanking(limit)
		if err != nil {
			return nil, err
		}
		for i, data := range moldData {
			name := data["mold_code"].(string)
			if moldName, ok := data["mold_name"].(string); ok && moldName != "" {
				name = name + " " + moldName
			}
			items = append(items, dto.RankingItem{
				Rank:       i + 1,
				Name:       name,
				Change:     0, // 暂时设为 0
				OrderCount: data["usage_count"].(int),
			})
		}
	default:
		return nil, fmt.Errorf("invalid ranking type: %s", rankingType)
	}

	return &dto.OrderRanking{
		Items: items,
	}, nil
}

func calculateGrowth(current, last int64) float64 {
	if last == 0 {
		if current > 0 {
			return 100.0
		}
		return 0.0
	}
	return math.Round(float64(current-last)/float64(last)*100*10) / 10
}

// calculateGrowthFloat 计算增长率（浮点数版本）
func calculateGrowthFloat(current, last float64) float64 {
	if last == 0 {
		if current > 0 {
			return 100.0
		}
		return 0.0
	}
	return math.Round((current-last)/last*100*10) / 10
}

// GetDetailOverview 获取统计报表页概览数据（4张卡片）
func (s *DashboardService) GetDetailOverview() (*dto.DashboardDetailOverview, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	type result struct {
		value interface{}
		err   error
		index int
	}

	var (
		completedOrders, completedOrdersLast         int64
		progressProductionOrders, progressProductionOrdersLast int64
		progressOrders, progressOrdersLast           int64
		completedMolds, completedMoldsLast           int64
	)

	var wg sync.WaitGroup
	resultChan := make(chan result, 8)
	wg.Add(8)

	go func() {
		defer wg.Done()
		v, err := s.dao.GetMonthlyCompletedProductionOrders()
		select { case resultChan <- result{value: v, err: err, index: 0}: case <-ctx.Done(): }
	}()
	go func() {
		defer wg.Done()
		v, err := s.dao.GetLastMonthCompletedProductionOrders()
		select { case resultChan <- result{value: v, err: err, index: 1}: case <-ctx.Done(): }
	}()
	go func() {
		defer wg.Done()
		v, err := s.dao.GetInProgressProductionOrderCount()
		select { case resultChan <- result{value: v, err: err, index: 2}: case <-ctx.Done(): }
	}()
	go func() {
		defer wg.Done()
		v, err := s.dao.GetLastMonthInProgressProductionOrderCount()
		select { case resultChan <- result{value: v, err: err, index: 3}: case <-ctx.Done(): }
	}()
	go func() {
		defer wg.Done()
		v, err := s.dao.GetInProgressMoldWorkOrderCount()
		select { case resultChan <- result{value: v, err: err, index: 4}: case <-ctx.Done(): }
	}()
	go func() {
		defer wg.Done()
		v, err := s.dao.GetLastMonthInProgressMoldWorkOrderCount()
		select { case resultChan <- result{value: v, err: err, index: 5}: case <-ctx.Done(): }
	}()
	go func() {
		defer wg.Done()
		v, err := s.dao.GetMonthlyCompletedMoldWorkOrders()
		select { case resultChan <- result{value: v, err: err, index: 6}: case <-ctx.Done(): }
	}()
	go func() {
		defer wg.Done()
		v, err := s.dao.GetLastMonthCompletedMoldWorkOrders()
		select { case resultChan <- result{value: v, err: err, index: 7}: case <-ctx.Done(): }
	}()

	go func() {
		wg.Wait()
		close(resultChan)
	}()

	for r := range resultChan {
		if r.err != nil {
			continue
		}
		switch v := r.value.(type) {
		case int64:
			switch r.index {
			case 0:
				completedOrders = v
			case 1:
				completedOrdersLast = v
			case 2:
				progressProductionOrders = v
			case 3:
				progressProductionOrdersLast = v
			case 4:
				progressOrders = v
			case 5:
				progressOrdersLast = v
			case 6:
				completedMolds = v
			case 7:
				completedMoldsLast = v
			}
		}
	}

	overview := &dto.DashboardDetailOverview{
		CompletedProductionOrders:              int(completedOrders),
		CompletedProductionOrdersGrowth:        calculateGrowth(completedOrders, completedOrdersLast),
		InProgressProductionOrders:             int(progressProductionOrders),
		InProgressProductionOrdersGrowth:       calculateGrowth(progressProductionOrders, progressProductionOrdersLast),
		InProgressMoldWorkOrders:               int(progressOrders),
		InProgressMoldWorkOrdersGrowth:         calculateGrowth(progressOrders, progressOrdersLast),
		CompletedMoldWorkOrders:                int(completedMolds),
		CompletedMoldWorkOrdersGrowth:          calculateGrowth(completedMolds, completedMoldsLast),
	}

	return overview, nil
}

// GetDetailProductionTrend 获取统计报表页的趋势图数据（4条折线，近12个月）
func (s *DashboardService) GetDetailProductionTrend() (*dto.DetailProductionTrend, error) {
	items, err := s.dao.GetMonthlyDetailProductionTrend()
	if err != nil {
		return nil, err
	}

	categories := make([]string, len(items))
	completedOrdersData := make([]float64, len(items))
	inProgressOrdersData := make([]float64, len(items))
	completedMoldsData := make([]float64, len(items))
	inProgressMoldsData := make([]float64, len(items))

	for i, item := range items {
		// item.Date = "2026-06"，格式化为 "6月"
		if len(item.Date) >= 7 {
			monthStr := item.Date[5:]
			if monthStr[0] == '0' {
				monthStr = monthStr[1:]
			}
			categories[i] = monthStr + "月"
		} else {
			categories[i] = item.Date
		}
		completedOrdersData[i] = float64(item.CompletedProductionOrders)
		inProgressOrdersData[i] = float64(item.InProgressProductionOrders)
		completedMoldsData[i] = float64(item.CompletedMoldWorkOrders)
		inProgressMoldsData[i] = float64(item.InProgressMoldWorkOrders)
	}

	datasets := []dto.TrendDataset{
		{Name: "completedProductionOrders", Data: completedOrdersData},
		{Name: "inProgressProductionOrders", Data: inProgressOrdersData},
		{Name: "completedMoldWorkOrders", Data: completedMoldsData},
		{Name: "inProgressMoldWorkOrders", Data: inProgressMoldsData},
	}

	return &dto.DetailProductionTrend{
		Categories: categories,
		Datasets:   datasets,
	}, nil
}

// GetDeliveryPunctuality 获取近12个月交付准时率数据（提前/准时/延迟）
func (s *DashboardService) GetDeliveryPunctuality() (*dto.DeliveryPunctuality, error) {
	items, err := s.dao.GetMonthlyDeliveryPunctuality()
	if err != nil {
		return nil, err
	}

	categories := make([]string, len(items))
	earlyData := make([]float64, len(items))
	onTimeData := make([]float64, len(items))
	delayedData := make([]float64, len(items))

	for i, item := range items {
		if len(item.Month) >= 7 {
			monthStr := item.Month[5:]
			if monthStr[0] == '0' {
				monthStr = monthStr[1:]
			}
			categories[i] = monthStr + "月"
		} else {
			categories[i] = item.Month
		}
		earlyData[i] = float64(item.Early)
		onTimeData[i] = float64(item.OnTime)
		delayedData[i] = float64(item.Delayed)
	}

	datasets := []dto.TrendDataset{
		{Name: "early", Data: earlyData},
		{Name: "onTime", Data: onTimeData},
		{Name: "delayed", Data: delayedData},
	}

	return &dto.DeliveryPunctuality{
		Categories: categories,
		Datasets:   datasets,
	}, nil
}

// GetDashboardService 获取 DashboardService 实例
func GetDashboardService() *DashboardService {
	return NewDashboardService()
}
