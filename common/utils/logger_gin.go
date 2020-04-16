package utils

import (
	"bufio"
	"bytes"
	"github.com/apache/dubbo-go/common/logger"
	"github.com/gin-gonic/gin"
	"time"
)

func LoggerWithWriter(notlogged ...string) gin.HandlerFunc {
	var skip map[string]struct{}
	if length := len(notlogged); length > 0 {
		skip = make(map[string]struct{}, length)
		for _, path := range notlogged {
			skip[path] = struct{}{}
		}
	}

	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery
		w := bufio.NewWriter(c.Writer)
		buff := bytes.Buffer{}
		newWriter := &bufferedWriter{c.Writer, w, buff}
		c.Writer = newWriter
		c.Next()
		if _, ok := skip[path]; !ok {
			end := time.Now()
			latency := end.Sub(start)
			clientIP := c.ClientIP()
			method := c.Request.Method
			statusCode := c.Writer.Status()
			if raw != "" {
				path = path + "?" + raw
			}
			logger.Infof(" | %d | %13v | %15s | %s | %s /n %s", statusCode, latency, clientIP, method, path, newWriter.Buffer.Bytes())
			_ = w.Flush()
		}
	}
}

type bufferedWriter struct {
	gin.ResponseWriter
	out    *bufio.Writer
	Buffer bytes.Buffer
}

func (g *bufferedWriter) Write(data []byte) (int, error) {
	g.Buffer.Write(data)
	return g.out.Write(data)
}