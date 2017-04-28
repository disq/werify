package main

import (
	"context"
	"errors"
	"log"
	"sync"
	"time"

	"github.com/disq/werify"
	t "github.com/disq/werify/cmd/werifyd/types"
	wrpc "github.com/disq/werify/rpc"
)

// Server is our main struct
type Server struct {
	context context.Context
	env     string

	// identifier is for tracing operation results back to the coordinator
	identifier wrpc.ServerIdentifier

	numWorkers int

	// hosts is the list of hosts
	hosts  []*t.Host
	hostMu sync.RWMutex

	// opBuffer is a map of operation handles vs. data
	opBuffer     map[string]wrpc.OperationOutput
	opMu         sync.RWMutex
	nextOpHandle uint64

	forceHealthcheck chan struct{}
}

func (s *Server) getHostByEndpoint(endpoint wrpc.Endpoint, lock bool) (index int, host *t.Host) {
	if lock {
		s.hostMu.RLock()
		defer s.hostMu.RUnlock()
	}

	for i, h := range s.hosts {
		if h.Endpoint == endpoint {
			return i, h
		}
	}

	return -1, nil
}

// rpcMiddleware is poor man's net/rpc middleware for checking compatibility (? not sure it'll be enough) and env tag
func (s *Server) rpcMiddleware(input *wrpc.CommonInput, callback func() error) error {
	if input == nil {
		return errors.New("commonInput nil pointer")
	}
	if input.EnvTag != s.env {
		return errors.New("env mismatch")
	}

	return callback()
}

// AddHost is the rpc handler to add a host to our host list
func (s *Server) AddHost(input wrpc.AddHostInput, output *wrpc.AddHostOutput) error {
	return s.rpcMiddleware(&input.CommonInput, func() error {
		ep := wrpc.NewEndpoint(string(input.Endpoint), werify.DefaultPort)

		i, _ := s.getHostByEndpoint(ep, true)
		if i > -1 {
			return errors.New("endpoint already exists in host list")
		}

		s.hostMu.Lock()
		defer s.hostMu.Unlock()

		h := &t.Host{
			Endpoint: ep,
			Added:    time.Now(),
			IsAlive:  false,
		}

		err := s.setIdentifier(h)
		if err != nil {
			return err
		}

		s.hosts = append(s.hosts, h)

		err = s.healthcheck(h)
		if err != nil {
			//log.Printf("Initial healthcheck failed for %v: %s", h, err.Error())
		} else if !h.IsAlive {
			log.Printf("Initial healthcheck not OK for %v", h)
		}

		output.Ok = true
		return nil
	})
}

// RemoveHost is the rpc handler to remove a host from our host list
func (s *Server) RemoveHost(input wrpc.RemoveHostInput, output *wrpc.RemoveHostOutput) error {
	return s.rpcMiddleware(&input.CommonInput, func() error {
		s.hostMu.Lock()
		defer s.hostMu.Unlock()

		e := wrpc.NewEndpoint(string(input.Endpoint), werify.DefaultPort)

		i, h := s.getHostByEndpoint(e, false)
		if i < 0 {
			return errors.New("endpoint does not exist in host list")
		}

		s.hosts = append(s.hosts[:i], s.hosts[i+1:]...)

		if h.Conn != nil {
			h.Conn.Close()
		}

		output.Ok = true
		return nil
	})
}

// ListHost is the rpc handler to list our hosts
func (s *Server) ListHost(input wrpc.ListHostsInput, output *wrpc.ListHostsOutput) error {
	return s.rpcMiddleware(&input.CommonInput, func() error {
		s.hostMu.RLock()
		defer s.hostMu.RUnlock()

		for _, h := range s.hosts {
			if input.ListActive && h.IsAlive {
				output.ActiveHosts = append(output.ActiveHosts, h.Endpoint)
			}
			if input.ListInactive && !h.IsAlive {
				output.InactiveHosts = append(output.InactiveHosts, h.Endpoint)
			}
		}

		return nil
	})
}

// HealthCheck is a no-op rpc handler for health-check purposes
func (s *Server) HealthCheck(input wrpc.HealthCheckInput, output *wrpc.HealthCheckOutput) error {
	return s.rpcMiddleware(&input.CommonInput, func() error {
		output.Ok = true
		return nil
	})
}

// SetIdentifier is the rpc handler to set our endpoint identifier
func (s *Server) SetIdentifier(input wrpc.SetIdentifierInput, output *wrpc.SetIdentifierOutput) error {
	return s.rpcMiddleware(&input.CommonInput, func() error {
		if s.identifier == "" {
			s.identifier = input.Identifier
			output.Ok = true
			return nil
		}
		if s.identifier == input.Identifier {
			output.Ok = true
			return nil
		}

		return errors.New("identifier already set, mismatch")
	})
}

// Refresh is the rpc handler to force a health-check
func (s *Server) Refresh(input wrpc.RefreshInput, output *wrpc.RefreshOutput) error {
	return s.rpcMiddleware(&input.CommonInput, func() error {
		// Don't block if we already have another one queued up
		go func() {
			s.forceHealthcheck <- struct{}{}
		}()
		output.Ok = true
		return nil
	})
}

// newCommonInput initializes and returns a CommonInput struct using server's's information
func (s *Server) newCommonInput() wrpc.CommonInput {
	return wrpc.CommonInput{
		EnvTag: s.env,
	}
}
