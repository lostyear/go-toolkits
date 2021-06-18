package prommetric

import (
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	httpCounter = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "http",
			Subsystem: "requests",
			Name:      "total_counter",
			Help:      "request counter",
		},
		[]string{"method"},
	)
	httpHistogram = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "http",
			Subsystem: "requests",
			Name:      "latency_histogram",
			Help:      "http latency",
		},
		[]string{"method", "handler", "status"},
	)
	httpSummary = promauto.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace: "http",
			Subsystem: "requests",
			Name:      "latency_summary",
			Help:      "http latency",
		},
		[]string{"method", "handler", "status"},
	)
	httpReqSummary = promauto.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace: "http",
			Subsystem: "requests",
			Name:      "size_summary",
			Help:      "http request size",
		},
		[]string{"method", "handler", "status"},
	)
	httpResSummary = promauto.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace: "http",
			Subsystem: "response",
			Name:      "size_summary",
			Help:      "http response size",
		},
		[]string{"method", "handler", "status"},
	)

	metricHandlerPath = ""
)

// MetricMiddleware is a gin framework middleware.
// it will support metric for prometheus monitor system
func MetricMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if metricHandlerPath == c.Request.RequestURI {
			return
		}

		st := time.Now()
		method := c.Request.Method
		httpCounter.WithLabelValues(method).Inc()

		c.Next()

		status := strconv.Itoa(c.Writer.Status())
		handler := c.HandlerName()
		// uri := c.Request.RequestURI
		reqSz := float64(c.Request.ContentLength)
		resSz := float64(c.Writer.Size())
		latency := float64(time.Since(st)) / float64(time.Millisecond)

		httpHistogram.WithLabelValues(method, handler, status).Observe(latency)
		httpSummary.WithLabelValues(method, handler, status).Observe(latency)
		httpReqSummary.WithLabelValues(method, handler, status).Observe(reqSz)
		httpResSummary.WithLabelValues(method, handler, status).Observe(resSz)
	}
}

// MetricHandler return a gin handler function to handler prometheus metric
func MetricHandler() gin.HandlerFunc {
	handler := promhttp.Handler()
	return func(c *gin.Context) {
		handler.ServeHTTP(c.Writer, c.Request)
	}
}

var runServer sync.Once

// StartSingleServer start a http server to handle prometheus metric request
func StartSingleServer(addr, path string) {
	runServer.Do(func() {
		eng := gin.New()
		eng.Use(gin.RecoveryWithWriter(emptyWriter{}))
		eng.GET(path, MetricHandler())
		eng.Run(addr)
	})
}

// StartDefaultServer start a http server to handle prometheus metric request with default exporter
func StartDefaultServer() {
	StartSingleServer(":9090", "/metrics")
}

type emptyWriter struct{}

func (emptyWriter) Write([]byte) (int, error)
