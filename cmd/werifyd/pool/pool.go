package pool

import (
	"context"
	"sync"

	wrpc "github.com/disq/werify/rpc"
)

type Pool struct {
	ctx        context.Context
	cancelFunc context.CancelFunc
	in         chan wrpc.WorkerOperation
	wg         *sync.WaitGroup
}

type PoolCallback func(wrpc.WorkerOperation)

func NewPool(ctx context.Context, ch chan wrpc.WorkerOperation) *Pool {
	ctxWithCancel, cancelFunc := context.WithCancel(ctx)
	p := &Pool{
		ctx:        ctxWithCancel,
		cancelFunc: cancelFunc,
		in:         ch,
		wg:         &sync.WaitGroup{},
	}
	return p
}

func (p *Pool) Start(numWorkers int, callback PoolCallback) {
	p.wg.Add(numWorkers)
	for i := 0; i < numWorkers; i++ {
		go p.worker(callback)
	}
}

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

func (p *Pool) Cancel() {
	p.cancelFunc()
}
