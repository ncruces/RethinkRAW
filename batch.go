package main

import (
	"sync"

	"rethinkraw/internal/util"
)

var batches Batches

type Batches struct {
	lock  sync.Mutex
	queue [16]struct {
		id    string
		paths []string
	}
}

func (p *Batches) New(paths []string) string {
	p.lock.Lock()
	defer p.lock.Unlock()

	for i := len(p.queue) - 1; i > 0; i-- {
		p.queue[i] = p.queue[i-1]
	}

	id := util.RandomID()
	p.queue[0].id = id
	p.queue[0].paths = paths
	return id
}

func (p *Batches) Get(id string) []string {
	p.lock.Lock()
	defer p.lock.Unlock()

	for j, t := range p.queue {
		if t.id == id {
			for i := j; i > 0; i-- {
				p.queue[i] = p.queue[i-1]
			}
			p.queue[0] = t
			return t.paths
		}
	}

	return nil
}
