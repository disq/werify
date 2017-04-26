package types

import (
	wrpc "github.com/disq/werify/rpc"
)

// WorkerOperation is Operation with unique name attached, for the worker pool. Satisfies the PoolData interface.
type WorkerOperation struct {
	Name string
	wrpc.Operation
}

func (w *WorkerOperation) GetName() string {
	return w.Name
}
func (w *WorkerOperation) GetOperation() *wrpc.Operation {
	return &w.Operation
}
func (w *WorkerOperation) GetHost() *Host {
	return nil
}
