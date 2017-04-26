package main

import (
	"fmt"
	"sync"

	"github.com/disq/werify/cmd/werifyd/checkers"
	"github.com/disq/werify/cmd/werifyd/pool"
	t "github.com/disq/werify/cmd/werifyd/types"
	wrpc "github.com/disq/werify/rpc"
)

func (s *Server) RunOperation(input wrpc.OperationInput, output *wrpc.OperationOutput) error {
	return s.rpcMiddleware(&input.CommonInput, func() error {
		output.Results = make(map[wrpc.ServerIdentifier]map[string]wrpc.OperationResult)

		ch := make(chan t.PoolData)
		p := pool.NewPool(s.context, ch)

		if !input.Forward {
			// Really run the checks, return results

			var mu sync.Mutex
			res := map[string]wrpc.OperationResult{}

			p.Start(s.numWorkers, func(pd t.PoolData) {
				r := s.operationRunner(pd.GetOperation())
				mu.Lock()
				defer mu.Unlock()
				res[pd.GetName()] = *r
			})

			for opId, op := range input.Ops {
				// Run each Operation in a worker concurrently
				o := t.WorkerOperation{
					Name:      opId,
					Operation: op,
				}
				ch <- &o
			}

			close(ch)
			p.Wait()

			output.Results[s.identifier] = res
			return nil
		}

		// Forward checks to alive hosts in a worker pool and reap results

		// We modify the input struct and use it below, same struct for all hosts
		// TODO further distribute work among peers using Forward=true calls?
		input.Forward = false

		rpcCmd := wrpc.BuildMethod(wrpc.RunOperationRpcCommand)
		var mu sync.Mutex

		p.Start(s.numWorkers, func(pd t.PoolData) {
			h := pd.GetHost()
			if !h.IsAlive {
				return
			}

			out := wrpc.OperationOutput{}

			// TODO what if it takes too long? Add a timeout
			err := h.Conn.Call(rpcCmd, input, &out)

			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				// A failed RPC call is a failed RPC call for all the commands.
				// We don't know the identifier of the server, so make one from the Endpoint (it should match, else we wouldn't have added this Host to our list)
				id := wrpc.ServerIdentifier(h.Endpoint)
				output.Results[id] = make(map[string]wrpc.OperationResult)
				for k := range input.Ops {
					s := output.Results[id][k]
					s.Err = err.Error()
					output.Results[id][k] = s
				}
				return
			}

			for id, r := range out.Results {
				output.Results[id] = r
			}
			return
		})

		s.hostMu.RLock()
		defer s.hostMu.RUnlock()

		for _, h := range s.hosts {
			// Run each RPC call for each Host in a worker concurrently
			o := t.WorkerHost(*h)
			ch <- &o
		}

		close(ch)
		p.Wait()

		return nil
	})
}

// operationRunner runs the Operation (checks) and returns the result
func (s *Server) operationRunner(op *wrpc.Operation) *wrpc.OperationResult {
	res := &wrpc.OperationResult{}

	var ok bool
	var err error

	switch op.OpType {
	case "file_exists":
		ok, err = checkers.DoesFileExist(string(op.PathArg))

	case "file_contains":
		ok, err = checkers.DoesFileHasWord(string(op.PathArg), string(op.CheckArg))

	case "process_running":
		ok, err = checkers.IsProcessRunning(string(op.CheckArg), string(op.PathArg))

	default:
		err = fmt.Errorf("Unhandled operation type: %s", op.OpType)
	}

	res.Success = ok
	if err != nil {
		res.Err = err.Error()
	}

	return res
}
