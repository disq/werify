package pool

import (
	"context"
	"sync"

	t "github.com/disq/werify/cmd/werifyd/types"
)

// Pool is a generic worker pool using t.PoolData as the datatype
type Pool struct {
	ctx        context.Context
	cancelFunc context.CancelFunc
	in         chan t.PoolData
	wg         *sync.WaitGroup
}

// PoolCallback is the data process callback
type PoolCallback func(t.PoolData)

// NewPool creates a new Pool
func NewPool(ctx context.Context, ch chan t.PoolData) *Pool {
	ctxWithCancel, cancelFunc := context.WithCancel(ctx)
	p := &Pool{
		ctx:        ctxWithCancel,
		cancelFunc: cancelFunc,
		in:         ch,
		wg:         &sync.WaitGroup{},
	}
	return p
}

// Start starts numWorkers number of workers on the pool, each running callback
func (p *Pool) Start(numWorkers int, callback PoolCallback) {
	p.wg.Add(numWorkers)
	for i := 0; i < numWorkers; i++ {
		go p.worker(callback)
	}
}

// Wait waits for the pool jobs to complete
func (p *Pool) Wait() {
	p.wg.Wait()
}

func (p *Pool) worker(callback PoolCallback) {
	defer p.wg.Done()

	for {
		select {
		case data, ok := <-p.in:
			if !ok { // channel closed
				return
			}
			callback(data)
		case <-p.ctx.Done():
			return
		}
	}
}

// Cancel cancels the jobs in the pool
func (p *Pool) Cancel() {
	p.cancelFunc()
}
