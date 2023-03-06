package pool

import (
	"context"
)

type JobPool struct {
	ctx      context.Context
	reqs     []any
	resps    []any
	fn       ProcessFn
	nworkers int
	rch      chan result
	done     chan struct{}
}

type result struct {
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
		rch:      make(chan result, nworkers),
		done:     make(chan struct{}),
	}
}

func (p *JobPool) Process() error {
	// run workers
	for w := 0; w < p.nworkers; w++ {
		go p.work(w)
	}

	// sync worker close
	defer close(p.done)

	// process result
	var count int
	for {
		select {
		case r := <-p.rch:
			count++
			if err, ok := r.val.(error); ok {
				return err
			}
			p.resps[r.idx] = r.val
			if count == len(p.reqs) {
				return nil
			}
		}
	}
}

func (p *JobPool) work(j int) {
	for i := j; i < len(p.reqs); i += p.nworkers {
		select {
		case <-p.done:
			return
		case <-p.ctx.Done():
			p.rch <- result{
				idx: i,
				val: p.ctx.Err(),
			}
			return

		default:
			rs, err := p.fn(p.ctx, p.reqs[i])
			if err != nil {
				p.rch <- result{
					idx: i,
					val: err,
				}
				return
			}
			p.rch <- result{
				idx: i,
				val: rs,
			}
		}
	}
}

func (p *JobPool) Results() []any {
	return p.resps
}
