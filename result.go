package hin

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	grpcCodes "google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"log"
	"net/http"
)

var Result = result{}

type response struct {
	HttpCode int    `json:"-"`
	GrpcCode int    `json:"http_code,omitempty"`
	Code     int    `json:"code"`
	Message  string `json:"message"`
	Request  string `json:"request,omitempty"`
}

func (r response) toJson() string {
	b, err := json.Marshal(r)
	if err != nil {
		panic(err)
	}

	return string(b)
}

type result struct{}

type resultOption func(*response)

func (r result) Fail(c *gin.Context, code int, opts ...resultOption) {
	r.Json(c, code, opts...)
}

func (r result) Success(c *gin.Context, opts ...resultOption) {
	r.Json(c, ErrOk, opts...)
}

func (r result) Created(c *gin.Context, opts ...resultOption) {
	r.Json(c, ErrCreated, opts...)
}

func (r result) Updated(c *gin.Context, opts ...resultOption) {
	r.Json(c, ErrUpdated, opts...)
}

func (r result) Deleted(c *gin.Context, opts ...resultOption) {
	r.Json(c, ErrDeleted, opts...)
}

func (r result) WithMessage(message string) resultOption {
	return func(r *response) {
		r.Message = message
	}
}

func (r result) WithHttpCode(code int) resultOption {
	return func(r *response) {
		r.HttpCode = code
	}
}

func (r result) WithCode(code int) resultOption {
	return func(r *response) {
		r.Code = code
	}
}

func (r result) Json(ctx *gin.Context, val any, opts ...resultOption) {
	res := &response{
		HttpCode: http.StatusOK,
		Request:  fmt.Sprintf("%s %s", ctx.Request.Method, ctx.Request.URL),
	}

	doOpts := func() {
		for _, o := range opts {
			o(res)
		}
	}

	if v, ok := val.(Error); ok {
		log.Printf("%#+v", val)
		res.HttpCode = v.HttpCode
		res.Code = v.Code
		res.Message = v.Message
		doOpts()
		ctx.AbortWithStatusJSON(res.HttpCode, res)
		return
	}

	if v, ok := val.(error); ok {
		log.Printf("%#+v", val)
		res.HttpCode = http.StatusInternalServerError
		res.Code = 9999
		res.Message = v.Error()
		doOpts()
		ctx.AbortWithStatusJSON(res.HttpCode, res)
		return
	}

	if code, ok := val.(int); ok {
		coder := ParseCoder(code)
		res.HttpCode = coder.HttpCode
		res.Code = coder.Code
		res.Message = coder.Message
		doOpts()
		ctx.AbortWithStatusJSON(res.HttpCode, res)
		return
	}

	doOpts()
	ctx.AbortWithStatusJSON(res.HttpCode, val)
}

func (r result) GrpcStatus(ctx context.Context, val any, opts ...resultOption) error {
	res := &response{
		GrpcCode: http.StatusOK,
	}

	doOpts := func() {
		for _, o := range opts {
			o(res)
		}
	}

	if v, ok := val.(Error); ok {
		log.Printf("%#+v", val)
		res.GrpcCode = v.HttpCode
		res.Code = v.Code
		res.Message = v.Message
		doOpts()
		return status.Error(grpcCodes.Aborted, res.toJson())
	}

	if v, ok := val.(error); ok {
		log.Printf("%#+v", val)
		res.GrpcCode = http.StatusInternalServerError
		res.Code = 9999
		res.Message = v.Error()
		doOpts()
		return status.Error(grpcCodes.Aborted, res.toJson())
	}

	if code, ok := val.(int); ok {
		coder := ParseCoder(code)
		res.GrpcCode = coder.HttpCode
		res.Code = coder.Code
		res.Message = coder.Message
		doOpts()
		return status.Error(grpcCodes.Aborted, res.toJson())
	}

	doOpts()
	return status.Error(grpcCodes.Aborted, res.toJson())
}

const (
	ErrOk int = iota + 0
	ErrCreated
	ErrUpdated
	ErrDeleted
)

const (
	// ErrParameterError 参数错误 400
	ErrParameterError int = iota + 10001
	// ErrResRepeat 资源重复 302
	ErrResRepeat
	// ErrNotFound 资源不存在 404
	ErrNotFound
	// ErrAuthHeaderRequired 缺少认证头 401
	ErrAuthHeaderRequired
	// ErrAuthHeaderInvalid 认证头无效 401
	ErrAuthHeaderInvalid
	// ErrForbidden 禁止访问 403
	ErrForbidden
	// ErrRequestBodyRequired 缺少请求体 400
	ErrRequestBodyRequired

	ErrTokenInvalid
	ErrTokenRequired
	ErrPermissionDenied
)

func init() {
	Register(ErrCoder{ErrOk, http.StatusOK, "ok"})
	Register(ErrCoder{ErrUpdated, http.StatusOK, "updated"})
	Register(ErrCoder{ErrCreated, http.StatusCreated, "created"})
	Register(ErrCoder{ErrDeleted, http.StatusNoContent, "deleted"})
	Register(ErrCoder{ErrParameterError, http.StatusBadRequest, "parameter error"})
	Register(ErrCoder{ErrResRepeat, http.StatusFound, "resource repeat"})
	Register(ErrCoder{ErrNotFound, http.StatusNotFound, "resource notfound"})
	Register(ErrCoder{ErrAuthHeaderRequired, http.StatusUnauthorized, "auth header required"})
	Register(ErrCoder{ErrAuthHeaderInvalid, http.StatusUnauthorized, "auth header is invalid"})
	Register(ErrCoder{ErrForbidden, http.StatusForbidden, "forbidden"})
	Register(ErrCoder{ErrRequestBodyRequired, http.StatusBadRequest, "request body required"})
	Register(ErrCoder{ErrTokenInvalid, http.StatusUnauthorized, "token invalid"})
	Register(ErrCoder{ErrTokenRequired, http.StatusUnauthorized, "token required"})
	Register(ErrCoder{ErrPermissionDenied, http.StatusForbidden, "permission denied"})
}
