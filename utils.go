package verkle

import (
	"sync"
	"sync/atomic"
)

type ProcessorPool struct {
	isOpen uint32
	quitCh chan struct{}
	wg     sync.WaitGroup

	queues [8]chan func()
}

func (p *ProcessorPool) Perform(task func(), queue byte) {
	if atomic.LoadUint32(&p.isOpen) == 0 {
		// Already closed, do nothing
		return
	}
	p.queues[queue] <- task
}

func worker(id int, jobs chan func()) {
	for {
		select {
		case job, open := <-jobs:
			if !open {
				return
			}
			job()
		}
	}
}

func (p *ProcessorPool) Start() {
	// Spin up one worker per queue
	p.wg.Add(len(p.queues))
	for i := range p.queues {
		p.queues[i] = make(chan func(), 100)
		go func(i int) {
			worker(i, p.queues[i])
			p.wg.Done()
		}(i)
	}
	// open for business
	atomic.SwapUint32(&p.isOpen, 1)
}

// Shutdown will wait for all the current jobs to execute before shutting down.
// While waiting for shutdown, no new jobs will be accepted.
func (p *ProcessorPool) Shutdown() {
	// first off, stop new jobs from entering
	if atomic.SwapUint32(&p.isOpen, 0) == 0 {
		// Already closed, do nothing
		return
	}
	// Send the closer on the queue, which means it will be the last thing
	// ever executed on the queue.
	for _, queue := range p.queues {
		q := queue
		queue <- func() {
			close(q)
		}
	}
	// Now wait for all goroutines to exit
	p.wg.Wait()
}

// Shutdown will wait until all currently executing jobs are done.
func (p *ProcessorPool) WaitIdle() {
	var wg sync.WaitGroup
	wg.Add(8)
	for _, queue := range p.queues {
		queue <- func() {
			wg.Done()
		}
	}
	wg.Wait()
}
