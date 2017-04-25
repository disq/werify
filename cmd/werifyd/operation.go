package main

import (
	"log"

	wrpc "github.com/disq/werify/rpc"
)

func (s *Server) RunOperation(input wrpc.OperationInput, output *wrpc.OperationOutput) error {
	return s.rpcMiddleware(&input.CommonInput, func() error {
		log.Printf("Forward: %t", input.Forward)
		log.Printf("Operations: %+v", input.Ops)

		if !input.Forward {
			// TODO: Really run the checks, return results

			return nil
		}

		// TODO: Forward checks to (alive?) hosts (worker pool?) and reap results

		return nil
	})
}
