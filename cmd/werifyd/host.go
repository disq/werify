package main

import (
	"log"
	"net"
	"net/rpc"
	"time"

	"errors"
	t "github.com/disq/werify/cmd/werifyd/types"
	wrpc "github.com/disq/werify/rpc"
)

const defaultTimeoutServerToServer = 10 * time.Second
const rpcHealthCheckTimeout = 5 * time.Second

func (s *Server) connect(h *t.Host) error {
	if h.Conn != nil {
		return nil
	}

	h.Lock()
	defer h.Unlock()

	connection, err := net.DialTimeout("tcp", string(h.Endpoint), defaultTimeoutServerToServer)
	if err != nil {
		h.Conn = nil
		h.IsAlive = false
		return err
	}

	//log.Printf("Connected to %v", h)

	h.Conn = rpc.NewClient(connection)
	return nil
}

func (s *Server) healthcheck(h *t.Host) (err error) {
	var expectedLiveness bool

	defer func() {
		// Log only state changes
		if err != nil && expectedLiveness {
			log.Printf("Healthcheck not OK: %v: %s", h, err.Error())

		} else if !expectedLiveness {
			log.Printf("Healthcheck OK: %v", h)
		}
	}()

	h.Lock()
	tm := time.Now()
	h.LastHealthCheckAttempt = &tm
	h.Unlock()

	err = s.connect(h)
	if err != nil {
		return err
	}

	h.Lock()
	defer h.Unlock()

	expectedLiveness = h.IsAlive

	out := wrpc.HealthCheckOutput{}
	in := wrpc.HealthCheckInput{CommonInput: s.newCommonInput()}

	call := h.Conn.Go(wrpc.BuildMethod("HealthCheck"), in, &out, nil)
	select {
	case ret := <-call.Done:
		err = ret.Error
	case <-time.After(rpcHealthCheckTimeout):
		err = errors.New("RPC call timed out")
	}
	if err != nil {
		h.IsAlive = false

		// Close HC-failed connection so that we reconnect the next time
		h.Conn.Close()
		h.Conn = nil
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
