package main

import (
	"log"
	"net"
	"net/rpc"
	"time"

	t "github.com/disq/werify/cmd/werifyd/types"
	wrpc "github.com/disq/werify/rpc"
)

const defaultTimeoutServerToServer = 10 * time.Second

func (s *Server) connect(h *t.Host) error {
	if h.Conn != nil {
		return nil
	}

	h.Lock()
	defer h.Unlock()

	connection, err := net.DialTimeout("tcp", string(h.Endpoint), defaultTimeoutServerToServer)
	if err != nil {
		h.Conn = nil
		return err
	}

	log.Printf("Connected to %v", h)

	h.Conn = rpc.NewClient(connection)
	return nil
}

func (s *Server) healthcheck(h *t.Host) (err error) {
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

	h.Lock()
	defer h.Unlock()

	out := wrpc.HealthCheckOutput{}
	in := wrpc.HealthCheckInput{CommonInput: s.newCommonInput()}
	t := time.Now()

	h.LastHealthCheckAttempt = &t
	err = h.Conn.Call(wrpc.BuildMethod("HealthCheck"), in, &out)
	if err != nil {
		h.IsAlive = false
		return err
	}

	h.IsAlive = out.Ok
	return nil
}

func (s *Server) setIdentifier(h *t.Host) (err error) {
	err = s.connect(h)
	if err != nil {
		return err
	}

	h.Lock()
	defer h.Unlock()

	out := wrpc.SetIdentifierOutput{}
	in := wrpc.SetIdentifierInput{
		CommonInput: s.newCommonInput(),
		Identifier:  wrpc.ServerIdentifier(h.Endpoint),
	}

	return h.Conn.Call(wrpc.BuildMethod("SetIdentifier"), in, &out)
}
