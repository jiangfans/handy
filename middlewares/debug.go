package middlewares

import (
	"bytes"
	"github.com/gin-gonic/gin"
	"github.com/jiangfans/handy/utils"
	"github.com/sirupsen/logrus"
	"io/ioutil"
)

type bodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w bodyLogWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

func Debug(env string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if env != utils.ProductionMode {
			var bodyBytes []byte
			if c.Request.Body != nil {
				bodyBytes, _ = ioutil.ReadAll(c.Request.Body)
			}
			c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))
			blw := &bodyLogWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
			c.Writer = blw

			c.Next()

			logrus.Debugf("req_uri: %s, method: %s, req_body: %s, resp_code: %d, resp_body: %s", c.Request.RequestURI, c.Request.Method, string(bodyBytes), c.Writer.Status(), blw.body.String())
		}
	}
}
