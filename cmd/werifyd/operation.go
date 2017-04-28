package main

import (
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/disq/werify/cmd/werifyd/checkers"
	"github.com/disq/werify/cmd/werifyd/pool"
	t "github.com/disq/werify/cmd/werifyd/types"
	wrpc "github.com/disq/werify/rpc"
)

const rpcOperationTimeout = 30 * time.Second

// OperationStatusCheck is the rpc handler to check the status of an ongoing or ended operation
func (s *Server) OperationStatusCheck(input wrpc.OperationStatusCheckInput, output *wrpc.OperationStatusCheckOutput) error {
	return s.rpcMiddleware(&input.CommonInput, func() error {
		o := s.getOpBuffer(input.Handle)
		if o == nil {
			return errors.New("Invalid handle")
		}

		output.Results = o.Results
		output.StartedAt = o.StartedAt
		output.EndedAt = o.EndedAt
		return nil
	})
}

func (s *Server) generateHandle() string {
	// increment and get the new value
	val := atomic.AddUint64(&(s.nextOpHandle), 1)
	return fmt.Sprintf("%s%d", randStringBytes(3), val)
}

func (s *Server) setOpBuffer(handle string, o *wrpc.OperationOutput) {
	s.opMu.Lock()
	defer s.opMu.Unlock()
	s.opBuffer[handle] = *o
}

func (s *Server) getOpBuffer(handle string) *wrpc.OperationOutput {
	s.opMu.RLock()
	defer s.opMu.RUnlock()

	o, ok := s.opBuffer[handle]
	if !ok {
		return nil
	}
	return &o
}

// RunOperation is the rpc handler to run a host check operation
func (s *Server) RunOperation(input wrpc.OperationInput, output *wrpc.OperationOutput) error {
	return s.rpcMiddleware(&input.CommonInput, func() error {

		if input.Forward {
			// Forward checks to alive hosts in a worker pool and reap results
			// But first generate a Handle and return it
			handle := s.generateHandle()
			output.Handle = handle

			go s.runAsyncOperation(handle, input)
			return nil
		}

		output.StartedAt = time.Now()
		output.Results = make(map[wrpc.ServerIdentifier]map[string]wrpc.OperationResult)

		ch := make(chan t.PoolData)
		p := pool.NewPool(s.context, ch)

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
		tm := time.Now()
		output.EndedAt = &tm
		return nil
	})
}

func (s *Server) runAsyncOperation(handle string, input wrpc.OperationInput) {
	// We modify the input struct and use it below, same struct for all hosts
	// TODO further distribute work among peers using Forward=true calls?
	input.Forward = false

	ch := make(chan t.PoolData)
	p := pool.NewPool(s.context, ch)

	rpcCmd := wrpc.BuildMethod(wrpc.RunOperationRpcCommand)
	var mu sync.Mutex

	output := wrpc.OperationOutput{
		Results:   make(map[wrpc.ServerIdentifier]map[string]wrpc.OperationResult),
		StartedAt: time.Now(),
	}
	s.setOpBuffer(handle, &output)

	p.Start(s.numWorkers, func(pd t.PoolData) {
		h := pd.GetHost()
		if !h.IsAlive {
			return
		}

		out := wrpc.OperationOutput{}

		var err error

		call := h.Conn.Go(rpcCmd, input, &out, nil)
		select {
		case ret := <-call.Done:
			err = ret.Error
		case <-time.After(rpcOperationTimeout):
			err = errors.New("RPC call timed out")
		}

		mu.Lock()
		defer mu.Unlock()
		if err != nil {
			// A failed RPC call is a failed RPC call for all the commands.
			// We won't know the identifier of the server, so make one from the Endpoint (it should match, else we wouldn't have added this Host to our list)
			id := wrpc.ServerIdentifier(h.Endpoint)
			output.Results[id] = make(map[string]wrpc.OperationResult)
			for k := range input.Ops {
				s := output.Results[id][k]
				s.Err = err.Error()
				output.Results[id][k] = s
			}
			s.setOpBuffer(handle, &output)
			return
		}

		for id, r := range out.Results {
			output.Results[id] = r
		}
		s.setOpBuffer(handle, &output)
		return
	})

	s.hostMu.RLock()
	defer s.hostMu.RUnlock()

	for _, h := range s.hosts {
		// Run each RPC call for each Host in a worker concurrently
		ch <- h
	}

	close(ch)
	p.Wait()

	tm := time.Now()
	output.EndedAt = &tm
	s.setOpBuffer(handle, &output)
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
