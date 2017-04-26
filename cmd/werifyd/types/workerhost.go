package types

import (
	wrpc "github.com/disq/werify/rpc"
)

// WorkerHost is Host for the worker pool. Satisfies the PoolData interface.
type WorkerHost Host

func (w *WorkerHost) GetName() string {
	return ""
}
func (w *WorkerHost) GetOperation() *wrpc.Operation {
	return nil
}
func (w *WorkerHost) GetHost() *Host {
	h := Host(*w)
	return &h
}
