package http

import (
	"io"
	"time"

	"github.com/fvbock/endless"
	"github.com/gin-gonic/gin"

	rlogs "github.com/lestrrat-go/file-rotatelogs"
	"github.com/lostyear/gin-middlewares/recovery"
	"github.com/lostyear/gin-middlewares/timeout"
)

type RegisterHandler func(*gin.RouterGroup)

type Config struct {
	Listen string

	LogPath          string
	LogRotationHours uint
	LogMaxDays       uint

	HTTPTimeoutMilliseSecond  int
	ReadTimeoutMilliseSecond  int
	WriteTimeoutMilliseSecond int
}

func StartHTTPServer(cfg Config, handler RegisterHandler, middlewares gin.HandlersChain) {
	eng := gin.New()

	eng.Use(recovery.Recovery())

	//TODO: use request log middleware
	//TODO: use metric middleware
	eng.Use(createLogger(cfg.LogPath, cfg.LogRotationHours, cfg.LogMaxDays))
	eng.Use(timeout.TimeoutMiddleware(
		time.Duration(cfg.HTTPTimeoutMilliseSecond)*time.Millisecond,
		`{"status":"timeout","msg":"Gateway Timeout"}`,
	))

	eng.Use(middlewares...)
	handler(eng.Group(""))

	GracefulRun(
		eng,
		cfg.Listen,
		time.Duration(cfg.ReadTimeoutMilliseSecond)*time.Millisecond,
		time.Duration(cfg.WriteTimeoutMilliseSecond)*time.Millisecond,
	)
}

func createLogger(filePath string, rotation, maxDays uint) gin.HandlerFunc {
	var w io.Writer

	if len(filePath) <= 0 {
		return gin.Logger()
	}

	w, err := rlogs.New(
		filePath,
		rlogs.WithRotationTime(time.Duration(rotation)*time.Hour),
		rlogs.WithMaxAge(time.Duration(maxDays)*time.Hour*24),
	)
	if err != nil {
		return gin.Logger()
	}

	return gin.LoggerWithWriter(w)

}

func GracefulRun(engine *gin.Engine, listenAddr string, readTimeout, writeTimeout time.Duration) {
	endless.DefaultReadTimeOut = readTimeout
	endless.DefaultWriteTimeOut = writeTimeout

	//TODO: 信号处理函数，支持reload和restart

	endless.ListenAndServe(listenAddr, engine)
}
