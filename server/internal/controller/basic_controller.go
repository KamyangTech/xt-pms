package controller

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/KamyangTech/xt-pms/server/internal/common/response"
	"github.com/KamyangTech/xt-pms/server/internal/model/entity"
	"github.com/KamyangTech/xt-pms/server/internal/service"
)

// parseID 将字符串ID转换为int64
func parseID(idStr string) int64 {
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return 0
	}
	return id
}

type BasicController struct {
	customerService  *service.CustomerService
	warehouseService *service.WarehouseService
	processService   *service.ProcessService
	supplierService  *service.SupplierService
	processorService *service.ProcessorService
}

func NewBasicController() *BasicController {
	return &BasicController{
		customerService:  service.NewCustomerService(),
		warehouseService: service.NewWarehouseService(),
		processService:   service.NewProcessService(),
		supplierService:  service.NewSupplierService(),
		processorService: service.NewProcessorService(),
	}
}

// GetCustomers 获取客户列表
func (b *BasicController) GetCustomers(c *gin.Context) {
	cusCode := c.Query("cusCode")
	cusName := c.Query("cusName")
	contact := c.Query("contact")
	phone := c.Query("phone")
	statusStr := c.DefaultQuery("status", "0")
	status, _ := strconv.Atoi(statusStr)

	HandlePageQuery(c, func(page, pageSize int) (interface{}, int64, error) {
		return b.customerService.GetCustomersByPageWithFilter(page, pageSize, cusCode, cusName, contact, phone, status)
	})
}

// GetCustomerByID 根据 ID 获取客户
func (b *BasicController) GetCustomerByID(c *gin.Context) {
	HandlePathID(c, func(id int64) (interface{}, error) {
		return b.customerService.GetCustomerByID(id)
	})
}

// CreateCustomer 创建客户
func (b *BasicController) CreateCustomer(c *gin.Context) {
	c.Set("oper_log_title", "创建客户")
	c.Set("oper_log_type", 1)
	c.Set("oper_log_business_type", "customer")
	HandleCreate(c, func(customer *entity.Customer) (interface{}, error) {
		return b.customerService.CreateCustomer(customer)
	})
}

// UpdateCustomer 更新客户
func (b *BasicController) UpdateCustomer(c *gin.Context) {
	c.Set("oper_log_title", "更新客户")
	c.Set("oper_log_type", 3)
	c.Set("oper_log_business_type", "customer")
	HandleUpdate(c, func(customer *entity.Customer) error {
		return b.customerService.UpdateCustomer(customer)
	})
}

// DeleteCustomer 删除客户
func (b *BasicController) DeleteCustomer(c *gin.Context) {
	c.Set("oper_log_title", "删除客户")
	c.Set("oper_log_type", 2)
	c.Set("oper_log_business_type", "customer")
	HandleDelete(c, func(id int64) error {
		return b.customerService.DeleteCustomer(id)
	})
}

// BatchDeleteCustomers 批量删除客户
func (b *BasicController) BatchDeleteCustomers(c *gin.Context) {
	c.Set("oper_log_title", "批量删除客户")
	c.Set("oper_log_type", 2)
	c.Set("oper_log_business_type", "customer")
	var req struct {
		IDs []int64 `json:"ids"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, response.ERROR, "请求参数格式错误")
		return
	}
	if err := b.customerService.BatchDeleteCustomers(req.IDs); err != nil {
		response.Error(c, response.ERROR, err.Error())
		return
	}
	response.Success(c, nil)
}

// GetWarehouses 获取仓库列表
func (b *BasicController) GetWarehouses(c *gin.Context) {
	whCode := c.Query("whCode")
	whName := c.Query("whName")
	manager := c.Query("manager")
	statusStr := c.DefaultQuery("status", "0")
	status, _ := strconv.Atoi(statusStr)

	HandlePageQuery(c, func(page, pageSize int) (interface{}, int64, error) {
		return b.warehouseService.GetWarehousesByPageWithFilter(page, pageSize, whCode, whName, manager, status)
	})
}

// GetWarehouseByID 根据 ID 获取仓库
func (b *BasicController) GetWarehouseByID(c *gin.Context) {
	HandlePathID(c, func(id int64) (interface{}, error) {
		return b.warehouseService.GetWarehouseByID(id)
	})
}

// CreateWarehouse 创建仓库
func (b *BasicController) CreateWarehouse(c *gin.Context) {
	c.Set("oper_log_title", "创建仓库")
	c.Set("oper_log_type", 1)
	c.Set("oper_log_business_type", "warehouse")
	HandleCreate(c, func(warehouse *entity.Warehouse) (interface{}, error) {
		return b.warehouseService.CreateWarehouse(warehouse)
	})
}

// UpdateWarehouse 更新仓库
func (b *BasicController) UpdateWarehouse(c *gin.Context) {
	c.Set("oper_log_title", "更新仓库")
	c.Set("oper_log_type", 3)
	c.Set("oper_log_business_type", "warehouse")
	HandleUpdate(c, func(warehouse *entity.Warehouse) error {
		return b.warehouseService.UpdateWarehouse(warehouse)
	})
}

// DeleteWarehouse 删除仓库
func (b *BasicController) DeleteWarehouse(c *gin.Context) {
	c.Set("oper_log_title", "删除仓库")
	c.Set("oper_log_type", 2)
	c.Set("oper_log_business_type", "warehouse")
	HandleDelete(c, func(id int64) error {
		return b.warehouseService.DeleteWarehouse(id)
	})
}

// BatchDeleteWarehouses 批量删除仓库
func (b *BasicController) BatchDeleteWarehouses(c *gin.Context) {
	c.Set("oper_log_title", "批量删除仓库")
	c.Set("oper_log_type", 2)
	c.Set("oper_log_business_type", "warehouse")
	var req struct {
		IDs []int64 `json:"ids"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, response.ERROR, "请求参数格式错误")
		return
	}
	if err := b.warehouseService.BatchDeleteWarehouses(req.IDs); err != nil {
		response.Error(c, response.ERROR, err.Error())
		return
	}
	response.Success(c, nil)
}

// GetProcesses 获取工序列表
func (b *BasicController) GetProcesses(c *gin.Context) {
	processes, err := b.processService.GetAllProcesses()
	if err != nil {
		response.Error(c, response.ERROR, "获取工序列表失败")
		return
	}

	response.Success(c, gin.H{
		"list":  processes,
		"total": len(processes),
	})
}

// GetProcessByID 根据 ID 获取工序
func (b *BasicController) GetProcessByID(c *gin.Context) {
	HandlePathID(c, func(id int64) (interface{}, error) {
		return b.processService.GetProcessByID(id)
	})
}

// CreateProcess 创建工序
func (b *BasicController) CreateProcess(c *gin.Context) {
	c.Set("oper_log_title", "创建工序")
	c.Set("oper_log_type", 1)
	c.Set("oper_log_business_type", "process")
	HandleCreate(c, func(process *entity.Process) (interface{}, error) {
		return b.processService.CreateProcess(process)
	})
}

// UpdateProcess 更新工序
func (b *BasicController) UpdateProcess(c *gin.Context) {
	c.Set("oper_log_title", "更新工序")
	c.Set("oper_log_type", 3)
	c.Set("oper_log_business_type", "process")
	HandleUpdate(c, func(process *entity.Process) error {
		return b.processService.UpdateProcess(process)
	})
}

// DeleteProcess 删除工序
func (b *BasicController) DeleteProcess(c *gin.Context) {
	c.Set("oper_log_title", "删除工序")
	c.Set("oper_log_type", 2)
	c.Set("oper_log_business_type", "process")
	HandleDelete(c, func(id int64) error {
		return b.processService.DeleteProcess(id)
	})
}

// GetSuppliers 获取供应商列表
func (b *BasicController) GetSuppliers(c *gin.Context) {
	supCode := c.Query("supCode")
	supName := c.Query("supName")
	contact := c.Query("contact")
	phone := c.Query("phone")
	statusStr := c.DefaultQuery("status", "0")
	status, _ := strconv.Atoi(statusStr)

	HandlePageQuery(c, func(page, pageSize int) (interface{}, int64, error) {
		return b.supplierService.GetSuppliersByPageWithFilter(page, pageSize, supCode, supName, contact, phone, status)
	})
}

// GetSupplierByID 根据 ID 获取供应商
func (b *BasicController) GetSupplierByID(c *gin.Context) {
	HandlePathID(c, func(id int64) (interface{}, error) {
		return b.supplierService.GetSupplierByID(id)
	})
}

// CreateSupplier 创建供应商
func (b *BasicController) CreateSupplier(c *gin.Context) {
	c.Set("oper_log_title", "创建供应商")
	c.Set("oper_log_type", 1)
	c.Set("oper_log_business_type", "supplier")
	HandleCreate(c, func(supplier *entity.Supplier) (interface{}, error) {
		return b.supplierService.CreateSupplier(supplier)
	})
}

// UpdateSupplier 更新供应商
func (b *BasicController) UpdateSupplier(c *gin.Context) {
	c.Set("oper_log_title", "更新供应商")
	c.Set("oper_log_type", 3)
	c.Set("oper_log_business_type", "supplier")
	HandleUpdate(c, func(supplier *entity.Supplier) error {
		return b.supplierService.UpdateSupplier(supplier)
	})
}

// DeleteSupplier 删除供应商
func (b *BasicController) DeleteSupplier(c *gin.Context) {
	c.Set("oper_log_title", "删除供应商")
	c.Set("oper_log_type", 2)
	c.Set("oper_log_business_type", "supplier")
	HandleDelete(c, func(id int64) error {
		return b.supplierService.DeleteSupplier(id)
	})
}

// BatchDeleteSuppliers 批量删除供应商
func (b *BasicController) BatchDeleteSuppliers(c *gin.Context) {
	c.Set("oper_log_title", "批量删除供应商")
	c.Set("oper_log_type", 2)
	c.Set("oper_log_business_type", "supplier")
	var req struct {
		IDs []int64 `json:"ids"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, response.ERROR, "请求参数格式错误")
		return
	}
	if err := b.supplierService.BatchDeleteSuppliers(req.IDs); err != nil {
		response.Error(c, response.ERROR, err.Error())
		return
	}
	response.Success(c, nil)
}

// GetProcessors 获取加工商列表
func (b *BasicController) GetProcessors(c *gin.Context) {
	processorCode := c.Query("processorCode")
	processorName := c.Query("processorName")
	contact := c.Query("contact")
	phone := c.Query("phone")
	statusStr := c.DefaultQuery("status", "0")
	status, _ := strconv.Atoi(statusStr)

	HandlePageQuery(c, func(page, pageSize int) (interface{}, int64, error) {
		return b.processorService.GetProcessorsByPageWithFilter(page, pageSize, processorCode, processorName, contact, phone, status)
	})
}

// GetProcessorByID 根据 ID 获取加工商
func (b *BasicController) GetProcessorByID(c *gin.Context) {
	HandlePathID(c, func(id int64) (interface{}, error) {
		return b.processorService.GetProcessorByID(id)
	})
}

// CreateProcessor 创建加工商
func (b *BasicController) CreateProcessor(c *gin.Context) {
	c.Set("oper_log_title", "创建加工商")
	c.Set("oper_log_type", 1)
	c.Set("oper_log_business_type", "processor")
	HandleCreate(c, func(processor *entity.Processor) (interface{}, error) {
		return b.processorService.CreateProcessor(processor)
	})
}

// UpdateProcessor 更新加工商
func (b *BasicController) UpdateProcessor(c *gin.Context) {
	c.Set("oper_log_title", "更新加工商")
	c.Set("oper_log_type", 3)
	c.Set("oper_log_business_type", "processor")
	HandleUpdate(c, func(processor *entity.Processor) error {
		return b.processorService.UpdateProcessor(processor)
	})
}

// DeleteProcessor 删除加工商
func (b *BasicController) DeleteProcessor(c *gin.Context) {
	c.Set("oper_log_title", "删除加工商")
	c.Set("oper_log_type", 2)
	c.Set("oper_log_business_type", "processor")
	HandleDelete(c, func(id int64) error {
		return b.processorService.DeleteProcessor(id)
	})
}

// BatchDeleteProcessors 批量删除加工商
func (b *BasicController) BatchDeleteProcessors(c *gin.Context) {
	c.Set("oper_log_title", "批量删除加工商")
	c.Set("oper_log_type", 2)
	c.Set("oper_log_business_type", "processor")
	var req struct {
		IDs []int64 `json:"ids"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, response.ERROR, "请求参数格式错误")
		return
	}
	if err := b.processorService.BatchDeleteProcessors(req.IDs); err != nil {
		response.Error(c, response.ERROR, err.Error())
		return
	}
	response.Success(c, nil)
}
