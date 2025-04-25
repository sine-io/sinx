// Package middleware 提供Gin框架的中间件
package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sine-io/sinx/pkg/logger"
	"go.uber.org/zap"
)

const (
	// TraceIDHeader 是HTTP请求头中传递追踪ID的字段名
	TraceIDHeader = "X-Trace-ID"
)

// TraceID 中间件，用于为每个请求添加追踪ID
// 追踪ID将从请求头中的X-Trace-ID获取，如果不存在则生成新的
func TraceID() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 尝试从请求头获取追踪ID
		traceID := c.GetHeader(TraceIDHeader)
		if traceID == "" {
			traceID = logger.NewTraceID() // 生成新的追踪ID
		}

		// 将追踪ID添加到上下文
		ctx := logger.ContextWithTraceID(c.Request.Context(), traceID)
		c.Request = c.Request.WithContext(ctx)

		// 在响应头中添加追踪ID
		c.Header(TraceIDHeader, traceID)

		// 记录请求开始
		logger.InfoContext(
			ctx,
			"开始处理请求",
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.String("client_ip", c.ClientIP()),
			zap.String("user_agent", c.Request.UserAgent()),
		)

		// 继续处理请求
		c.Next()

		// 记录请求结束
		logger.InfoContext(
			ctx,
			"完成处理请求",
			zap.Int("status", c.Writer.Status()),
			zap.Int("size", c.Writer.Size()),
		)
	}
}

// RequestLogger 记录请求信息的中间件
// 比TraceID中间件更详细，包含请求参数、响应时间等
func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 确保有追踪ID
		ctx := c.Request.Context()
		if logger.TraceIDFromContext(ctx) == "" {
			traceID := c.GetHeader(TraceIDHeader)
			if traceID == "" {
				traceID = logger.NewTraceID()
			}
			ctx = logger.ContextWithTraceID(ctx, traceID)
			c.Request = c.Request.WithContext(ctx)
			c.Header(TraceIDHeader, traceID)
		}

		// 开始时间
		start := logger.TimeNow()

		// 记录请求参数
		var requestParams []zap.Field

		// 添加查询参数
		if len(c.Request.URL.RawQuery) > 0 {
			requestParams = append(requestParams, zap.String("query", c.Request.URL.RawQuery))
		}

		// 添加请求头(排除敏感信息)
		headers := make(map[string]string)
		for k, v := range c.Request.Header {
			// 跳过敏感的Authorization头，只保留类型
			if strings.ToLower(k) == "authorization" {
				if len(v) > 0 && len(v[0]) > 0 {
					parts := strings.SplitN(v[0], " ", 2)
					if len(parts) > 0 {
						headers[k] = parts[0] + " [REDACTED]" // 只保留认证类型
					}
				}
			} else {
				headers[k] = strings.Join(v, ",")
			}
		}

		if len(headers) > 0 {
			requestParams = append(requestParams, zap.Any("headers", headers))
		}

		// 记录请求详情
		logger.InfoContext(
			ctx,
			"收到API请求",
			append(requestParams,
				zap.String("method", c.Request.Method),
				zap.String("path", c.Request.URL.Path),
				zap.String("client_ip", c.ClientIP()),
				zap.String("user_agent", c.Request.UserAgent()),
			)...,
		)

		// 处理请求
		c.Next()

		// 计算处理时间
		duration := logger.TimeNow().Sub(start)

		// 记录响应详情
		logger.InfoContext(
			ctx,
			"API请求完成",
			zap.Int("status", c.Writer.Status()),
			zap.Int("size", c.Writer.Size()),
			zap.Duration("duration", duration),
			zap.Int("errors", len(c.Errors)),
		)

		// 记录错误(如果有)
		if len(c.Errors) > 0 {
			for i, err := range c.Errors {
				logger.ErrorContext(
					ctx,
					"API请求错误",
					zap.Int("error_index", i),
					zap.String("error", err.Error()),
				)
			}
		}
	}
}

// RecoveryWithLogger 从panic恢复并记录错误的中间件
func RecoveryWithLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// 获取上下文中的追踪ID，如果没有则生成一个
				ctx := c.Request.Context()
				if logger.TraceIDFromContext(ctx) == "" {
					ctx = logger.ContextWithTraceID(ctx, logger.NewTraceID())
					c.Request = c.Request.WithContext(ctx)
				}

				// 记录panic错误
				logger.ErrorContext(
					ctx,
					"服务器内部错误",
					zap.Any("error", err),
					zap.String("method", c.Request.Method),
					zap.String("path", c.Request.URL.Path),
				)

				// 返回500错误
				c.AbortWithStatus(500)
			}
		}()

		c.Next()
	}
}
