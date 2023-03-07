package pool

import (
	"context"
	"sync"
)

type JobPool struct {
	ctx      context.Context
	reqs     []any
	resps    []any
	fn       ProcessFn
	nworkers int
	jch      chan job
	rch      chan result
	quit     chan struct{}
	wg       sync.WaitGroup
}

type result struct {
	idx int
	val any
}

type job struct {
	idx int
	val any
}

type ProcessFn func(ctx context.Context, req any) (any, error)

func NewJobPool(ctx context.Context, reqs []any, nworkers int, fn ProcessFn) *JobPool {
	return &JobPool{
		ctx:      ctx,
		reqs:     reqs,
		resps:    make([]any, len(reqs)),
		nworkers: nworkers,
		fn:       fn,
		jch:      make(chan job, nworkers),
		rch:      make(chan result, nworkers),
		quit:     make(chan struct{}),
	}
}

func (p *JobPool) Process() (err error) {
	// run workers
	p.wg.Add(p.nworkers + 1)
	for w := 0; w < p.nworkers; w++ {
		go p.work(w)
	}

	go func() {
		defer p.wg.Done()
		// send jobs
	sendloop:
		for idx, req := range p.reqs {
			select {
			case <-p.quit:
				break sendloop
			case <-p.ctx.Done():
				break sendloop
			default:
				p.jch <- job{
					idx: idx,
					val: req,
				}
			}
		}
		close(p.jch)
	}()

	// process results
ploop:
	for c := 0; c < len(p.reqs); {
		select {
		case r := <-p.rch:
			c++
			var ok bool
			if err, ok = r.val.(error); ok {
				break ploop
			}
			p.resps[r.idx] = r.val
		}
	}
	// sync workers
	close(p.quit)

	// wait workers
	p.wg.Wait()

	return err
}

func (p *JobPool) work(j int) {
	defer p.wg.Done()

	for job := range p.jch {
		select {
		case <-p.quit:
			return
		case <-p.ctx.Done():
			p.rch <- result{
				idx: job.idx,
				val: p.ctx.Err(),
			}
			return

		default:
			rs, err := p.fn(p.ctx, job.val)
			if err != nil {
				p.rch <- result{
					idx: job.idx,
					val: err,
				}
				return
			}
			p.rch <- result{
				idx: job.idx,
				val: rs,
			}
		}
	}
}

func (p *JobPool) Results() []any {
	return p.resps
}
