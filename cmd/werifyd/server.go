package main

import (
	"context"
	"errors"
	"log"
	"sync"
	"time"

	"github.com/disq/werify"
	wrpc "github.com/disq/werify/rpc"
)

type Server struct {
	context context.Context
	env     string

	// identifier is for tracing operation results back to the coordinator
	identifier string

	hosts  []*Host
	hostMu sync.RWMutex
}

func (s *Server) getHostByEndpoint(endpoint wrpc.Endpoint, lock bool) (index int, host *Host) {
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

func (s *Server) AddHost(input wrpc.AddHostInput, output *wrpc.AddHostOutput) error {
	return s.rpcMiddleware(&input.CommonInput, func() error {
		i, _ := s.getHostByEndpoint(input.Endpoint, true)
		if i > -1 {
			return errors.New("endpoint already exists in host list")
		}

		s.hostMu.Lock()
		defer s.hostMu.Unlock()

		h := &Host{
			Endpoint: wrpc.NewEndpoint(string(input.Endpoint), werify.DefaultPort),
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

		if h.conn != nil {
			h.conn.Close()
		}

		output.Ok = true
		return nil
	})
}

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

func (s *Server) HealthCheck(input wrpc.HealthCheckInput, output *wrpc.HealthCheckOutput) error {
	return s.rpcMiddleware(&input.CommonInput, func() error {
		output.Ok = true
		return nil
	})
}

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

// newCommonInput initializes and returns a CommonInput struct using server's's information
func (s *Server) newCommonInput() wrpc.CommonInput {
	return wrpc.CommonInput{
		EnvTag: s.env,
	}
}
