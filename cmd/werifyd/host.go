package main

import (
	"fmt"
	"log"
	"net"
	"net/rpc"
	"sync"
	"time"

	"github.com/disq/werify"
	wrpc "github.com/disq/werify/rpc"
)

type Host struct {
	Endpoint               wrpc.Endpoint
	Added                  time.Time
	LastHealthCheckAttempt *time.Time
	IsAlive                bool

	mu   sync.Mutex
	conn *rpc.Client
}

func (h *Host) String() string {
	return fmt.Sprintf("Host[%s alive=%t]", h.Endpoint, h.IsAlive)
}

func (s *Server) connect(h *Host) error {
	if h.conn != nil {
		return nil
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	connection, err := net.DialTimeout("tcp", string(h.Endpoint), werify.DefaultTimeoutServerToServer)
	if err != nil {
		h.conn = nil
		return err
	}

	log.Printf("Connected to %v", h)

	h.conn = rpc.NewClient(connection)
	return nil
}

func (s *Server) healthcheck(h *Host) (err error) {
	defer func() {
		if err != nil {
			log.Printf("Healthcheck not OK: %v: %s", h, err.Error())

		} else {
			log.Printf("Healthcheck OK: %v", h)
		}
	}()

	err = s.connect(h)
	if err != nil {
		return err
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	out := wrpc.HealthCheckOutput{}
	in := wrpc.HealthCheckInput{CommonInput: s.newCommonInput()}
	t := time.Now()

	h.LastHealthCheckAttempt = &t
	err = h.conn.Call(wrpc.BuildMethod("HealthCheck"), in, &out)
	if err != nil {
		h.IsAlive = false
		return err
	}

	h.IsAlive = out.Ok
	return nil
}

func (s *Server) setIdentifier(h *Host) (err error) {
	err = s.connect(h)
	if err != nil {
		return err
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	out := wrpc.SetIdentifierOutput{}
	in := wrpc.SetIdentifierInput{
		CommonInput: s.newCommonInput(),
		Identifier:  string(h.Endpoint),
	}

	return h.conn.Call(wrpc.BuildMethod("SetIdentifier"), in, &out)
}
