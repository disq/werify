package pool

import (
	"context"
	"sync"

	t "github.com/disq/werify/cmd/werifyd/types"
)

type Pool struct {
	ctx        context.Context
	cancelFunc context.CancelFunc
	in         chan t.PoolData
	wg         *sync.WaitGroup
}

type PoolCallback func(t.PoolData)

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
