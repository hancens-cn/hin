package hin

import (
	"context"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/jinzhu/copier"
	"io"
	"net/http"
	"strings"
)

func handleBind(ctx *gin.Context, err error) error {
	var errs validator.ValidationErrors
	if err.Error() == io.EOF.Error() {
		Result.Fail(ctx, ErrRequestBodyRequired)
	} else if errors.As(err, &errs) {
		validationErrors := map[string]any{}

		for _, validationErr := range errs {
			validationErrors[strings.ToLower(validationErr.Field())] = validationErr.ActualTag()
		}

		Result.Json(ctx, H{"code": ErrParameterError, "message": validationErrors, "request": fmt.Sprintf("%s %s", ctx.Request.Method, ctx.Request.URL)}, Result.WithHttpCode(http.StatusBadRequest))
	} else {
		Result.Fail(ctx, ErrParameterError, Result.WithMessage(err.Error()))
	}

	return err
}

func BindJSON(ctx *gin.Context, o any) error {
	if err := ctx.ShouldBindJSON(o); err != nil {
		return handleBind(ctx, err)
	}
	return nil
}

func BindQuery(ctx *gin.Context, o any) error {
	if err := ctx.ShouldBindQuery(o); err != nil {
		return handleBind(ctx, err)
	}
	return nil
}

func BindUri(ctx *gin.Context, o any) error {
	if err := ctx.ShouldBindUri(o); err != nil {
		return handleBind(ctx, err)
	}
	return nil
}

func BindGrpcRequest(ctx context.Context, to, from any) error {
	validate := validator.New()
	validate.SetTagName("binding")
	if err := copier.Copy(to, from); err != nil {
		return err
	}

	if err := validate.Struct(to); err != nil {
		return err
	}
	return nil
}
