package main

import (
	"fmt"
	"log"
	"sync"

	"github.com/disq/werify/cmd/werifyd/checkers"
	"github.com/disq/werify/cmd/werifyd/pool"
	wrpc "github.com/disq/werify/rpc"
)

func (s *Server) RunOperation(input wrpc.OperationInput, output *wrpc.OperationOutput) error {
	return s.rpcMiddleware(&input.CommonInput, func() error {
		log.Printf("Forward: %t", input.Forward)
		log.Printf("Operations: %+v", input.Ops)

		output.Results = make(map[wrpc.ServerIdentifier]map[string]wrpc.OperationResult)

		if !input.Forward {
			// Really run the checks, return results

			var mu sync.Mutex
			res := map[string]wrpc.OperationResult{}

			ch := make(chan wrpc.WorkerOperation)
			p := pool.NewPool(s.context, ch)
			p.Start(s.numWorkers, func(op wrpc.WorkerOperation) {
				r := s.operationRunner(op.Operation)
				mu.Lock()
				defer mu.Unlock()
				res[op.Name] = *r
			})

			for opId, op := range input.Ops {
				ch <- wrpc.WorkerOperation{
					Name:      opId,
					Operation: op,
				}
			}

			close(ch)
			p.Wait()

			output.Results[s.identifier] = res
			return nil
		}

		// TODO: Forward checks to (alive?) hosts (worker pool?) and reap results

		return nil
	})
}

// operationRunner runs the Operation (checks) and returns the result
func (s *Server) operationRunner(op wrpc.Operation) *wrpc.OperationResult {
	res := &wrpc.OperationResult{}

	switch op.OpType {
	case "file_exists":
		ok, err := checkers.DoesFileExist(string(op.PathArg))
		res.Success = ok
		res.Err = err

	case "file_contains":
		ok, err := checkers.DoesFileHasWord(string(op.PathArg), string(op.CheckArg))
		res.Success = ok
		res.Err = err

	case "process_running":
		ok, err := checkers.IsProcessRunning(string(op.CheckArg), string(op.PathArg))
		res.Success = ok
		res.Err = err

	default:
		res.Err = fmt.Errorf("Unhandled operation type: %s", op.OpType)
	}

	return res
}
