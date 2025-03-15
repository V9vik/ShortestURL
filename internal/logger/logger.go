package logger

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"time"
)

func ginLogger(log *zap.SugaredLogger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		duration := time.Since(start)
		status := c.Writer.Status()
		size := c.Writer.Size()

		log.Infoln(
			"uri", c.Request.RequestURI,
			"method", c.Request.Method,
			"duration", duration,
			"status", status,
			"size", size,
		)
	}
}
