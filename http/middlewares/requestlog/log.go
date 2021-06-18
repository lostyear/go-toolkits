package requestlog

import (
	"fmt"
	"io"
	"log"
	"time"

	"github.com/gin-gonic/gin"
	rlogs "github.com/lestrrat-go/file-rotatelogs"
)

// RequestFileLogMiddleware create rotated log file to record gin request
func RequestFileLogMiddleware(filePath string, rotationHours, maxDays uint) gin.HandlerFunc {
	var w io.Writer

	if len(filePath) <= 0 {
		return gin.Logger()
	}

	w, err := rlogs.New(
		filePath+".%Y%m%d%H",
		rlogs.WithRotationTime(time.Duration(rotationHours)*time.Hour),
		rlogs.WithMaxAge(time.Duration(maxDays)*time.Hour*24),
	)
	if err != nil {
		log.Fatalf("init gin request log writer got error: %s.\n", err.Error())
	}

	return gin.LoggerWithConfig(gin.LoggerConfig{
		Formatter: Formatter,
		Output:    w,
		SkipPaths: []string{"/ping", "/health", "/healthz"},
	})
}

// Formatter is gin log formatter
func Formatter(params gin.LogFormatterParams) string {
	return fmt.Sprintf(
		"[access] %s\t%s\t[%s]\t%d\t%v\t%s\t%s\t%s\t[%s]\t[%s]\t%s\n",
		params.Request.Host,
		params.ClientIP,
		params.TimeStamp.Format(time.RFC3339Nano),
		params.StatusCode,
		params.Latency,
		params.Method,
		params.Path,
		params.Request.Proto,
		params.Request.Referer(),
		params.Request.UserAgent(),
		params.ErrorMessage,
	)
}
