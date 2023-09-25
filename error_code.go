package hin

import (
	"errors"
	"fmt"
	"net/http"
	"sync"
)

const ErrFailed = 9999

type ErrCoder struct {
	Code     int
	HttpCode int
	Message  string
}

type Error struct {
	ErrCoder
	error
}

// codes contains a map of error codes to metadata.
var codes = map[int]ErrCoder{}
var codeMux = &sync.Mutex{}

// Register mount a user define error code.
// It will panic when the same code already exist.
func Register(coder ErrCoder) {
	if coder.Code == 9999 {
		panic("code '9999' is reserved by errors as ErrUnknown error code")
	}

	codeMux.Lock()
	defer codeMux.Unlock()

	if _, ok := codes[coder.Code]; ok {
		panic(fmt.Sprintf("code: %d already exist", coder.Code))
	}

	codes[coder.Code] = coder
}

func ParseCoder(code int) ErrCoder {
	if coder, ok := codes[code]; ok {
		return coder
	}

	return codes[ErrFailed]
}

type ErrorOption func(*ErrCoder)

func WithErrMessage(msg string) ErrorOption {
	return func(e *ErrCoder) {
		e.Message = msg
	}
}

func WithErrHttpCode(code int) ErrorOption {
	return func(e *ErrCoder) {
		e.HttpCode = code
	}
}

func NewError(err error, code int, opts ...ErrorOption) Error {
	coder := ParseCoder(code)
	if err == nil {
		err = errors.New(coder.Message)
	}
	coder.Message = err.Error()
	for _, o := range opts {
		o(&coder)
	}
	return Error{
		error:    err,
		ErrCoder: coder,
	}
}

func init() {
	codes[ErrFailed] = ErrCoder{Code: ErrFailed, HttpCode: http.StatusInternalServerError, Message: "An internal server error occurred"}
}
