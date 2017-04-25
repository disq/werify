package rpc

import (
	"fmt"
	"strings"
)

// Endpoint is the host:port of a (hopefully) running werifyd
type Endpoint string

func NewEndpoint(hostPort string, defaultPort int) Endpoint {
	if strings.Index(hostPort, ":") == -1 {
		return Endpoint(fmt.Sprintf("%s:%d", hostPort, defaultPort))
	}

	return Endpoint(hostPort)
}
