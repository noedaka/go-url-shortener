package objectpool

import "sync"

type Pool[T Resetter] struct {
	pool sync.Pool
}

func New[T Resetter]() *Pool[T] {
	return &Pool[T]{
		pool: sync.Pool{
			New: func() any {
				var zero T
				return zero
			},
		},
	}
}

func (p *Pool[T]) Get() T {
	return p.pool.Get().(T)
}

func (p *Pool[T]) Put(obj T) {
	obj.Reset()
	p.pool.Put(obj)
}
