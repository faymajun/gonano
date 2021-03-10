// Package echologrus provides a middleware for echo that logs request details
// via the logrus logging library
package logger // fknsrs.biz/p/echo-logrus

import (
	"bytes"
	"fmt"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"strconv"
	"time"

	"go.uber.org/zap"

	"github.com/labstack/echo/v4"
)

func ZapLogger() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) (err error) {
			req := c.Request()

			var reqBody []byte
			if c.Request().Body != nil { // Read
				reqBody, _ = ioutil.ReadAll(req.Body)
			}
			req.Body = ioutil.NopCloser(bytes.NewBuffer(reqBody)) // Reset

			if len(reqBody) > 512 {
				reqBody = []byte{}
			}

			var queryData string
			for k, v := range c.QueryParams() {
				var val string
				if len(v) > 0 {
					val = v[0]
				}
				queryData += fmt.Sprintf("%s:%s ", k, val)
			}
			res := c.Response()
			start := time.Now()
			if err = next(c); err != nil {
				c.Error(err)
			}
			stop := time.Now()
			errMsg := ""
			if err != nil {
				errMsg = err.Error()
			}
			logrus.Infoln(req.Method,
				zap.String("uri", req.RequestURI),
				zap.Int("status", res.Status),
				zap.String("remote_ip", c.RealIP()),
				zap.String("host", req.Host),
				zap.Int64("latency", int64(stop.Sub(start))),
				zap.String("latency_human", stop.Sub(start).String()),
				zap.String("bytes_in", req.Header.Get(echo.HeaderContentLength)),
				zap.String("bytes_out", strconv.FormatInt(res.Size, 10)),
				zap.String("user_agent", req.UserAgent()),
				zap.String("error", errMsg),
				zap.String("reqBody", string(reqBody)),
				zap.String("query", queryData),
			)
			return
		}
	}
}
