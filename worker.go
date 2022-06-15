package workerPool

import (
	"sync/atomic"
)

//===========[STATIC/CACHE]====================================================================================================

var uniqueIdCounter int64 = 0

//===========[STRUCTS]====================================================================================================

type worker[TWork any] struct {
	//Provides direct access to the worker
	directWork chan TWork

	//If this channel is closed, go routine gets terminated
	terminate chan struct{}

	//A channel of work that this worker will be dealing with
	workBucket chan TWork

	//This function will be processing the work from workBucket
	workHandler func(work ...TWork)

	//A different function can be set for direct work send to a worker. If not set, it simply uses workHandler
	directWorkHandler func(work ...TWork)

	//ID of the worker
	id int
}

//DirectWork sends some work directly to this worker
func (w *worker[TWork]) DirectWork(work TWork) {
	w.directWork <- work
}

//Terminate makes this worker redundant
func (w *worker[TWork]) Terminate() {
	close(w.terminate)
}

//SetWorkHandler sets a new handler for all the work done
func (w *worker[TWork]) SetWorkHandler(handler func(work ...TWork)) {
	w.workHandler = handler
}

//SetDirectWorkHandler sets a new worker for all the work that is coming directly for this worker
func (w *worker[TWork]) SetDirectWorkHandler(handler func(work ...TWork)) {
	w.directWorkHandler = handler
}

//===========[FUNCTIONALITY]====================================================================================================

//workerGoroutine is the goroutine that will be spawned for each worker. As an argument you must pass in worker struct
func workerGoroutine[TWork any](w *worker[TWork]) {
	if w == nil {
		return
	}

	for {
		select {
		case work := <-w.directWork:
			if w.directWorkHandler == nil {
				w.workHandler(work)
				continue
			}
			w.directWorkHandler(work)

		case work := <-w.workBucket:
			w.workHandler(work)

		case <-w.terminate:
			for {
				select {
				case work := <-w.directWork:
					w.directWorkHandler(work)

				default:
					break
				}
			}
			break
		}
	}
}

//Creates and returns a new worker
func newWorker[TWork any](workPile chan TWork, workHandler, directWorkHandler func(work ...TWork)) *worker[TWork] {

	defaultWorkHandler := func(work ...TWork) {}

	w := &worker[TWork]{
		directWork:        make(chan TWork, 2),
		terminate:         make(chan struct{}),
		workBucket:        workPile,
		workHandler:       workHandler,
		directWorkHandler: directWorkHandler,
		id:                int(atomic.AddInt64(&uniqueIdCounter, 1)),
	}

	if w.workHandler == nil {
		w.SetWorkHandler(defaultWorkHandler)
	}

	go workerGoroutine(w)

	return w
}
