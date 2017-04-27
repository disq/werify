// Package types contains types that are used by werifyd and its worker pools. It doesn't try to be an elegant solution.
package types

import (
	"fmt"
	"net/rpc"
	"sync"
	"time"

	wrpc "github.com/disq/werify/rpc"
)

// Type Host is the main struct for each host, satisfies the PoolData interface.
type Host struct {
	Endpoint               wrpc.Endpoint
	Added                  time.Time
	LastHealthCheckAttempt *time.Time
	IsAlive                bool

	sync.Mutex
	Conn *rpc.Client
}

func (h *Host) String() string {
	return fmt.Sprintf("Host[%s alive=%t]", h.Endpoint, h.IsAlive)
}

func (w *Host) GetName() string {
	return ""
}
func (w *Host) GetOperation() *wrpc.Operation {
	return nil
}
func (w *Host) GetHost() *Host {
	return w
}
