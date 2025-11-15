package objectpool

import "sync"

type Pool[T Resetter] struct {
	objects []T
	mu      sync.Mutex
}

func New[T Resetter]() *Pool[T] {
	return &Pool[T]{
		objects: make([]T, 0),
	}
}

func (p *Pool[T]) Get() T {
	p.mu.Lock()
	defer p.mu.Unlock()

	if len(p.objects) == 0 {
		var zero T
		return zero
	}

	obj := p.objects[len(p.objects)-1]
	p.objects = p.objects[:len(p.objects)-1]

	return obj
}

func (p *Pool[T]) Put(obj T) {
	obj.Reset()

	p.mu.Lock()
	defer p.mu.Unlock()

	p.objects = append(p.objects, obj)
}
