package n9emetric

import (
	"time"

	"github.com/gin-gonic/gin"
	statsd "github.com/n9e/metrics-go/statsdlib"
)

// MetricMiddleware is a gin framework middleware.
// it will report metric to n9e monitor system
func MetricMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		st := time.Now()

		c.Next()

		caller := c.HandlerName()
		callee := c.Request.RequestURI
		method := c.Request.Method
		status := c.Request.Response.StatusCode
		latency := time.Since(st)

		statsd.RpcMetric(
			"http.request",
			caller, callee, latency, status,
			map[string]string{"method": method},
		)
	}
}
