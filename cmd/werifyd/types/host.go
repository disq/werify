// Package types contains types that are used by werifyd and its worker pools. It doesn't try to be an elegant solution.
package types

import (
	"fmt"
	"net/rpc"
	"sync"
	"time"

	wrpc "github.com/disq/werify/rpc"
)

// Host is the main struct for each host, satisfies the PoolData interface
type Host struct {
	Endpoint               wrpc.Endpoint
	Added                  time.Time
	LastHealthCheckAttempt *time.Time
	IsAlive                bool

	sync.Mutex
	Conn *rpc.Client
}

// String is the stringer method for the Host
func (h *Host) String() string {
	return fmt.Sprintf("Host[%s alive=%t]", h.Endpoint, h.IsAlive)
}

// GetName satisfies the PoolData interface, returning zero-value
func (h *Host) GetName() string {
	return ""
}

// GetOperation satisfies the PoolData interface, returning zero-value
func (h *Host) GetOperation() *wrpc.Operation {
	return nil
}

// GetHost satisfies the PoolData interface, returning host
func (h *Host) GetHost() *Host {
	return h
}
