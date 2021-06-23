package timeout

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"net/http"
	"path"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// Handler handle a func with timeout
func Handler(timeout time.Duration, timeoutMsg string, handler gin.HandlerFunc) gin.HandlerFunc {
	return timeoutHandlerFunc(timeout, timeoutMsg, handler)
}

// Middleware handles timeout exception
func Middleware(timeout time.Duration, timeoutMsg string) gin.HandlerFunc {
	handler := func(c *gin.Context) {
		c.Next()
	}
	return timeoutHandlerFunc(timeout, timeoutMsg, handler)
}

func timeoutHandlerFunc(timeout time.Duration, timeoutMsg string, handler gin.HandlerFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		// if gin framework already run serverError, this is no need
		if c.Writer.Written() {
			return
		}
		if c.Writer.Status() != 200 {
			return
		}

		// ensure write response before middleware exit
		wg := sync.WaitGroup{}
		wg.Add(1)
		defer wg.Wait()

		ctx := c.Request.Context()
		timeoutCtx, cancel := context.WithTimeout(ctx, timeout)

		w := c.Writer
		done := make(chan struct{})
		defer close(done)

		c.Request = c.Request.WithContext(timeoutCtx)
		tw := &timeoutWriter{
			ResponseWriter: w,
			h:              make(http.Header),
			req:            c.Request,
		}
		c.Writer = tw

		go func() {
			defer cancel()
			defer wg.Done()
			select {
			case <-done:
				tw.Lock()
				defer tw.Unlock()

				dst := w.Header()
				for k, vv := range tw.h {
					dst[k] = vv
				}

				if !tw.wroteHeader {
					if w.Status() > 0 {
						tw.code = w.Status()
					} else {
						tw.code = http.StatusOK
					}
				}
				w.WriteHeader(tw.code)
				w.Write(tw.wbuf.Bytes())
			case <-timeoutCtx.Done():
				tw.Lock()
				defer tw.Unlock()
				w.WriteHeader(http.StatusGatewayTimeout)
				w.WriteString(timeoutMsg)
				tw.timedOut = true
			}
		}()

		handler(c)
	}
}

type timeoutWriter struct {
	gin.ResponseWriter
	req  *http.Request
	h    http.Header
	wbuf bytes.Buffer

	sync.RWMutex
	timedOut    bool
	wroteHeader bool
	code        int
}

func (tw *timeoutWriter) Header() http.Header {
	tw.RLock()
	defer tw.RUnlock()
	return tw.h
}

func (tw *timeoutWriter) Write(p []byte) (int, error) {
	tw.Lock()
	defer tw.Unlock()
	if tw.timedOut {
		// TODO: 超时处理时间记录
		// 返回error会导致panic，暂时不返回error，后期可以考虑在panic中记录超时处理时间
		// return 0, http.ErrHandlerTimeout
		return 0, nil
	}
	if !tw.wroteHeader {
		tw.writeHeaderLocked(http.StatusOK)
	}
	return tw.wbuf.Write(p)
}

func (tw *timeoutWriter) writeHeaderLocked(code int) {
	checkWriteHeaderCode(code)

	switch {
	case tw.timedOut:
		return
	case tw.wroteHeader:
		if tw.req != nil {
			caller := relevantCaller()
			logf(tw.req, "http: superfluous response.WriteHeader call from %s (%s:%d)", caller.Function, path.Base(caller.File), caller.Line)
		}
	default:
		tw.wroteHeader = true
		tw.code = code
	}
}

func (tw *timeoutWriter) Status() int {
	tw.RLock()
	defer tw.RUnlock()
	if tw.code != 0 {
		return tw.code
	}
	return tw.ResponseWriter.Status()
}

func (tw *timeoutWriter) Size() int {
	tw.RLock()
	defer tw.RUnlock()
	return tw.wbuf.Len()
}

func (tw *timeoutWriter) Wroten() bool {
	tw.RLock()
	defer tw.RUnlock()
	return tw.wroteHeader
}

func (tw *timeoutWriter) WriteHeader(code int) {
	tw.Lock()
	defer tw.Unlock()
	tw.writeHeaderLocked(code)
}

func (tw *timeoutWriter) WriteHeaderNow() {
	if !tw.Wroten() {
		tw.WriteHeader(tw.code)
	}
}

func (tw *timeoutWriter) WriteString(s string) (n int, err error) {
	tw.Lock()
	defer tw.Unlock()
	return tw.wbuf.WriteString(s)

}

// FROM net/http/server.go
func checkWriteHeaderCode(code int) {
	// Issue 22880: require valid WriteHeader status codes.
	// For now we only enforce that it's three digits.
	// In the future we might block things over 599 (600 and above aren't defined
	// at https://httpwg.org/specs/rfc7231.html#status.codes)
	// and we might block under 200 (once we have more mature 1xx support).
	// But for now any three digits.
	//
	// We used to send "HTTP/1.1 000 0" on the wire in responses but there's
	// no equivalent bogus thing we can realistically send in HTTP/2,
	// so we'll consistently panic instead and help people find their bugs
	// early. (We can't return an error from WriteHeader even if we wanted to.)
	if code < 100 || code > 999 {
		panic(fmt.Sprintf("invalid WriteHeader code %v", code))
	}
}

// FROM net/http/server.go
// relevantCaller searches the call stack for the first function outside of net/http.
// The purpose of this function is to provide more helpful error messages.
func relevantCaller() runtime.Frame {
	pc := make([]uintptr, 16)
	n := runtime.Callers(1, pc)
	frames := runtime.CallersFrames(pc[:n])
	var frame runtime.Frame
	for {
		frame, more := frames.Next()
		if !strings.HasPrefix(frame.Function, "net/http.") {
			return frame
		}
		if !more {
			break
		}
	}
	return frame
}

// FROM net/http/server.go
// logf prints to the ErrorLog of the *Server associated with request r
// via ServerContextKey. If there's no associated server, or if ErrorLog
// is nil, logging is done via the log package's standard logger.
func logf(r *http.Request, format string, args ...interface{}) {
	s, _ := r.Context().Value(http.ServerContextKey).(*http.Server)
	if s != nil && s.ErrorLog != nil {
		s.ErrorLog.Printf(format, args...)
	} else {
		log.Printf(format, args...)
	}
}
