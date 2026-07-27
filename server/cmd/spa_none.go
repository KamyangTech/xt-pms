//go:build !release

package main

import "github.com/gin-gonic/gin"

// serveEmbeddedSPA 开发环境下为空实现，前端由 Vite dev server 独立提供。
func serveEmbeddedSPA(r *gin.Engine) {}
