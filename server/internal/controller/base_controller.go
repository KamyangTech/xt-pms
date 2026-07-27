package controller

import (
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/KamyangTech/xt-pms/server/internal/common/response"
)

type BaseController struct{}

func NewBaseController() *BaseController {
	return &BaseController{}
}

// resolveError 将 error 转为适当的业务状态码和消息
//   如果是 *response.HttpError，使用其 Code
//   否则默认返回 500
func resolveError(err error) (code int, msg string) {
	var httpErr *response.HttpError
	if errors.As(err, &httpErr) {
		return httpErr.Code, httpErr.Message
	}
	return response.ERROR, err.Error()
}

// handleServiceError 统一处理 service 层返回的错误，自动区分业务错误和系统错误
//   如果 err 是 *response.HttpError，使用其指定的 HTTP 状态码
//   否则返回 500
// 替代手写的 response.Error(ctx, response.ERROR, err.Error())，让所有 service 错误都能被正确解析
func handleServiceError(c *gin.Context, err error) {
	code, msg := resolveError(err)
	response.Error(c, code, msg)
}

// HandleRequest 通用请求处理器（泛型函数）
// T: 请求参数的类型
// handler: 具体的业务逻辑函数，接收解析后的参数，返回数据和错误
func HandleRequest[T any](c *gin.Context, handler func(*T) (interface{}, error)) {
	var req T
	// 自动绑定并校验 JSON 参数
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, response.BAD_REQUEST, "请求参数格式错误")
		return
	}

	// 执行业务逻辑
	data, err := handler(&req)
	if err != nil {
		code, msg := resolveError(err)
		response.Error(c, code, msg)
		return
	}

	response.Success(c, data)
}

// HandlePathID 通用路径参数 ID 获取与处理
// handler: 具体的业务逻辑函数，接收 ID，返回数据和错误
func HandlePathID(c *gin.Context, handler func(int64) (interface{}, error)) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.Error(c, response.BAD_REQUEST, "无效的 ID 参数")
		return
	}

	data, err := handler(id)
	if err != nil {
		code, msg := resolveError(err)
		response.Error(c, code, msg)
		return
	}

	response.Success(c, data)
}

// HandleCreate 通用创建请求处理器
// T: 请求参数的类型
// handler: 具体的创建逻辑函数，接收解析后的参数，返回创建结果和错误
func HandleCreate[T any](c *gin.Context, handler func(*T) (interface{}, error)) {
	var req T
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, response.BAD_REQUEST, "请求参数格式错误")
		return
	}

	data, err := handler(&req)
	if err != nil {
		code, msg := resolveError(err)
		response.Error(c, code, msg)
		return
	}

	response.Success(c, data)
}

// HandleUpdate 通用更新请求处理器
// T: 请求参数的类型
// handler: 具体的更新逻辑函数，接收解析后的参数，返回错误
func HandleUpdate[T any](c *gin.Context, handler func(*T) error) {
	var req T
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, response.BAD_REQUEST, "请求参数格式错误")
		return
	}

	if err := handler(&req); err != nil {
		code, msg := resolveError(err)
		response.Error(c, code, msg)
		return
	}

	response.Success(c, req)
}

// HandleDelete 通用删除请求处理器
// handler: 具体的删除逻辑函数，接收 ID，返回错误
func HandleDelete(c *gin.Context, handler func(int64) error) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.Error(c, response.BAD_REQUEST, "无效的 ID 参数")
		return
	}

	if err := handler(id); err != nil {
		code, msg := resolveError(err)
		response.Error(c, code, msg)
		return
	}

	response.Success(c, nil)
}

// HandlePageQuery 通用分页查询处理器
// handler: 具体的查询逻辑函数，接收 page 和 pageSize，返回列表、总数和错误
func HandlePageQuery(c *gin.Context, handler func(int, int) (interface{}, int64, error)) {
	pageStr := c.DefaultQuery("page", "1")
	pageSizeStr := c.DefaultQuery("page_size", "10")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil || pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	list, total, err := handler(page, pageSize)
	if err != nil {
		code, msg := resolveError(err)
		response.Error(c, code, msg)
		return
	}

	response.Success(c, gin.H{
		"list":  list,
		"total": total,
	})
}
