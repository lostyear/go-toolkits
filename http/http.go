package http

import (
	"fmt"
	"net/http"
	"time"

	"github.com/fvbock/endless"
	"github.com/gin-gonic/gin"

	"github.com/lostyear/go-toolkits/http/middlewares/n9emetric"
	"github.com/lostyear/go-toolkits/http/middlewares/prommetric"
	"github.com/lostyear/go-toolkits/http/middlewares/recovery"
	"github.com/lostyear/go-toolkits/http/middlewares/requestlog"
	"github.com/lostyear/go-toolkits/http/middlewares/timeout"
	"github.com/lostyear/go-toolkits/http/response"
)

// RegisterHandler used for start func, it shoud register all handlers
type RegisterHandler func(*gin.Engine)

// NOTE: 信号处理需要在外部实现，reload config 需要有个读写锁

// Config http server
type Config struct {
	Listen string

	Metric           string
	LogPath          string
	LogRotationHours uint
	LogMaxDays       uint

	HTTPTimeoutMilliseSecond  int
	ReadTimeoutMilliseSecond  int
	WriteTimeoutMilliseSecond int
}

var (
	emptyHandler = func(*gin.Context) {}
)

// StartHTTPServer run http server with config, handler should have all route regist action,
// and all middlewares will be used.
// also it has some default middlewares use
func StartHTTPServer(cfg Config, handler RegisterHandler, middlewares gin.HandlersChain) {
	eng := gin.New()

	eng.Use(GetMetricMiddleWare(cfg.Metric))
	eng.Use(requestlog.RequestFileLogMiddleware(cfg.LogPath, cfg.LogRotationHours, cfg.LogMaxDays))
	eng.Use(recovery.Recovery())
	eng.Use(timeout.Middleware(
		time.Duration(cfg.HTTPTimeoutMilliseSecond)*time.Millisecond,
		`{"status":"timeout","msg":"Gateway Timeout"}`,
	))

	eng.Use(middlewares...)
	eng.NoRoute(noRouteHandler)
	eng.NoMethod(noMethodHandler)

	handler(eng)

	GracefulRun(
		eng,
		cfg.Listen,
		time.Duration(cfg.ReadTimeoutMilliseSecond)*time.Millisecond,
		time.Duration(cfg.WriteTimeoutMilliseSecond)*time.Millisecond,
	)
}

// GracefulRun serve http server with graceful stop
func GracefulRun(engine *gin.Engine, listenAddr string, readTimeout, writeTimeout time.Duration) {
	endless.DefaultReadTimeOut = readTimeout
	endless.DefaultWriteTimeOut = writeTimeout

	//TODO: 信号处理函数，支持reload和restart

	endless.ListenAndServe(listenAddr, engine)
}

func noRouteHandler(c *gin.Context) {
	c.JSON(http.StatusNotFound, response.DefaultResponse{
		Status:  http.StatusNotFound,
		Message: fmt.Sprintf("No route to your request: %s %s", c.Request.Method, c.Request.RequestURI),
	})
}

func noMethodHandler(c *gin.Context) {
	c.JSON(http.StatusNotFound, response.DefaultResponse{
		Status:  http.StatusNotFound,
		Message: fmt.Sprintf("Not support Method to your request: %s %s", c.Request.Method, c.Request.RequestURI),
	})
}

// GetMetricMiddleWare choose use n9e or prometheus metric middleware
func GetMetricMiddleWare(metric string) gin.HandlerFunc {
	switch metric {
	case "n9e":
		return n9emetric.MetricMiddleware()
	case "prom":
		prommetric.StartDefaultServer()
		return prommetric.MetricHandler()
	default:
		return emptyHandler
	}
}
