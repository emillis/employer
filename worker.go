package workerPool

import (
	"sync/atomic"
	"time"
)

//===========[STATIC/CACHE]====================================================================================================

var uniqueIdCounter int64 = 0

//===========[STRUCTS]====================================================================================================

type worker[TWork any] struct {
	//A channel of work that this worker will be dealing with
	workBucket chan TWork

	//If this channel is closed, go routine gets terminated
	terminate chan struct{}

	//This function will be processing the work from workBucket
	workHandler func(Worker, TWork)

	//Defines amount of time until this worker quits
	timeout time.Duration

	//When the worker terminates, it sends its ID to this channel
	timedOutWorkers chan int

	//ID of the worker
	id int
}

//Terminate makes this worker redundant
func (w *worker[TWork]) Terminate() {
	w.terminate <- struct{}{}
}

//Id returns this worker's ID
func (w *worker[TWork]) Id() int {
	return w.id
}

//===========[FUNCTIONALITY]====================================================================================================

//workerGoroutine is the goroutine that will be spawned for each worker. As an argument you must pass incomingWork worker struct
func workerGoroutine[TWork any](w *worker[TWork]) {
	if w == nil {
		return
	}

	defer func() { w.timedOutWorkers <- w.id }()

	for {
		select {
		case work := <-w.workBucket:
			w.workHandler(w, work)

		case <-time.After(w.timeout):
			w.Terminate()

		case <-w.terminate:
			for {
				select {
				case work := <-w.workBucket:
					w.workHandler(w, work)

				default:
					return
				}
			}

		}
	}
}

//Creates and returns a new worker
func newWorker[TWork any](workPile chan TWork, workHandler func(Worker, TWork), timeout time.Duration, timedOutWorkers chan int) *worker[TWork] {
	if workHandler == nil {
		workHandler = func(Worker, TWork){}
	}

	w := &worker[TWork]{
		terminate:         make(chan struct{}, 2),
		workBucket:        workPile,
		workHandler:       workHandler,
		timeout:           timeout,
		timedOutWorkers:   timedOutWorkers,
		id:                int(atomic.AddInt64(&uniqueIdCounter, 1)),
	}

	go workerGoroutine(w)

	return w
}
