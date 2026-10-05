package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func redactTokenQuery(path string) string {
	queryStart := strings.IndexByte(path, '?')
	if queryStart < 0 || queryStart == len(path)-1 {
		return path
	}
	parts := strings.Split(path[queryStart+1:], "&")
	changed := false
	for i, part := range parts {
		key, _, _ := strings.Cut(part, "=")
		if key != "token" {
			continue
		}
		parts[i] = key + "=REDACTED"
		changed = true
	}
	if !changed {
		return path
	}
	return path[:queryStart+1] + strings.Join(parts, "&")
}

func accessLogFormatter(param gin.LogFormatterParams) string {
	var statusColor, methodColor, resetColor string
	if param.IsOutputColor() {
		statusColor = param.StatusCodeColor()
		methodColor = param.MethodColor()
		resetColor = param.ResetColor()
	}
	if param.Latency > time.Minute {
		param.Latency = param.Latency.Truncate(time.Second)
	}
	return fmt.Sprintf("[GIN] %v |%s %3d %s| %13v | %15s |%s %-7s %s %#v\n%s",
		param.TimeStamp.Format("2006/01/02 - 15:04:05"),
		statusColor, param.StatusCode, resetColor,
		param.Latency,
		param.ClientIP,
		methodColor, param.Method, resetColor,
		redactTokenQuery(param.Path),
		param.ErrorMessage,
	)
}
