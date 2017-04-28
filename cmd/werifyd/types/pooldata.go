package types

import (
	wrpc "github.com/disq/werify/rpc"
)

// PoolData is used by the worker pool, which can be either an operation pool or an host pool
type PoolData interface {
	GetName() string
	GetOperation() *wrpc.Operation
	GetHost() *Host
}
