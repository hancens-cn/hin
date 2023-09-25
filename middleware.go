package hin

import (
	"bytes"
	"github.com/casbin/casbin/v2"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/samber/lo"
	"github.com/spf13/viper"
	"strings"
)

var (
	headerXRequestID = ""
	admins           []string
)

func PreRequestContext() gin.HandlerFunc {
	headerXRequestID = viper.GetString("server.header.request_id")
	if headerXRequestID == "" {
		headerXRequestID = "X-Request-ID"
	}

	admins = viper.GetStringSlice("jwt.admins")

	return func(c *gin.Context) {
		// Get id from request
		rid := c.GetHeader(headerXRequestID)
		if rid == "" {
			rid = uuid.New().String()
			c.Request.Header.Add(headerXRequestID, rid)
		}
		c.Header(headerXRequestID, rid)
	}
}

func TokenParse(logger *Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader(headerXRequestID)
		if requestID == "" {
			panic("request id is empty, please check your middleware or configuration. server -> header -> request_id")
		}

		token := c.GetHeader("Authorization")
		if token == "" || !strings.HasPrefix(token, "Bearer") {
			c.Set(headerXRequestID, ErrTokenRequired)
			return
		}

		claims, err := ParseToken(strings.TrimPrefix(token, "Bearer "))
		if err != nil || claims.Issuer != JwtAccessKey {
			c.Set(headerXRequestID, ErrTokenInvalid)
			return
		}

		RegisterContext(requestID, &CurrentContext{
			claims,
			logger,
		})
		c.Set(headerXRequestID, requestID)
		c.Next()
		UnRegisterContext(requestID)
	}
}

func LoginRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, exist := c.Get(headerXRequestID)
		if !exist {
			Result.Fail(c, ErrTokenRequired)
			c.Abort()
			return
		}

		if code, ok := claims.(int); ok {
			Result.Fail(c, code)
			c.Abort()
			return
		}
	}
}

func PermissionRequired(casbin *casbin.Enforcer) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, exist := c.Get(headerXRequestID)
		if !exist {
			Result.Fail(c, ErrTokenRequired)
			c.Abort()
			return
		}

		if code, ok := claims.(int); ok {
			Result.Fail(c, code)
			c.Abort()
			return
		}

		ctx := GetCurrentContext(c)

		if lo.Contains(admins, ctx.Claims.ID) {
			return
		}

		if allowed, err := casbin.Enforce(ctx.Claims.ID, c.Request.URL.Path, c.Request.Method); err != nil || !allowed {
			Result.Fail(c, ErrPermissionDenied)
			c.Abort()
			return
		}
	}
}

type ResponseWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func NewResponseWriter(w gin.ResponseWriter) *ResponseWriter {
	rw := &ResponseWriter{
		ResponseWriter: w,
		body:           bytes.NewBuffer([]byte{}),
	}
	return rw
}

func (w ResponseWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

func (w ResponseWriter) WriteString(s string) (int, error) {
	w.body.WriteString(s)
	return w.ResponseWriter.WriteString(s)
}

func (w ResponseWriter) Body() string {
	return w.body.String()
}
