package main

import (
	"log"
	"time"

	"github.com/disq/werify/cmd/werifyd/pool"
	t "github.com/disq/werify/cmd/werifyd/types"
)

const healthCheckInterval = 60 * time.Second

func (s *Server) healthchecker() {
	for {
		select {
		case <-s.context.Done():
			return
		case <-s.forceHealthcheck:
			log.Println("Starting forced health checks...")
			s.runHealthcheck()
		case <-time.After(healthCheckInterval):
			s.runHealthcheck()
		}
	}
}

func (s *Server) runHealthcheck() {
	ch := make(chan t.PoolData)
	p := pool.NewPool(s.context, ch)

	p.Start(s.numWorkers, func(pd t.PoolData) {
		s.healthcheck(pd.GetHost())
		//log.Printf("HC for %s: %v\n", pd.GetHost().Endpoint, err)
	})

	s.hostMu.RLock()
	defer s.hostMu.RUnlock()

	for _, h := range s.hosts {
		// Run each RPC call for each Host in a worker concurrently
		ch <- h
	}

	close(ch)
	p.Wait()
}
