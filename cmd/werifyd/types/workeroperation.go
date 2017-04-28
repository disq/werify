package types

import (
	wrpc "github.com/disq/werify/rpc"
)

// WorkerOperation is Operation with unique name attached, for the worker pool. Satisfies the PoolData interface.
type WorkerOperation struct {
	Name string
	wrpc.Operation
}

// GetName satisfies the PoolData interface, returning the name
func (w *WorkerOperation) GetName() string {
	return w.Name
}

// GetOperation satisfies the PoolData interface, returning the operation
func (w *WorkerOperation) GetOperation() *wrpc.Operation {
	return &w.Operation
}

// GetHost satisfies the PoolData interface, returning the zero-value
func (w *WorkerOperation) GetHost() *Host {
	return nil
}
