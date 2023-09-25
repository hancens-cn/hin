package hin

import (
	"context"
	"go.uber.org/zap"
	"sync"
)

type CurrentContext struct {
	Claims *JwtClaims
	Logger *Logger
}

var (
	ctxMap = map[string]*CurrentContext{}
	ctxMux = &sync.Mutex{}
)

func RegisterContext(key string, ctx *CurrentContext) {
	ctxMux.Lock()
	defer ctxMux.Unlock()

	if _, ok := ctxMap[key]; ok {
		ctx.Logger.Error("########## !!! current context key repeat !!! ##########", zap.String("key", key), zap.Any("ctxMap", ctxMap))
	}

	ctxMap[key] = ctx
}

func UnRegisterContext(key string) {
	ctxMux.Lock()
	defer ctxMux.Unlock()

	delete(ctxMap, key)
}

func GetCurrentContext(ctx context.Context) *CurrentContext {
	key := ctx.Value(headerXRequestID).(string)
	if ctx, ok := ctxMap[key]; ok {
		return ctx
	}

	return nil
}
