package hin

import "github.com/google/wire"

var HinSet = wire.NewSet(
	NewMongoDB,
	NewCasbin,
)
