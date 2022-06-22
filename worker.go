package workerPool

import (
	"sync/atomic"
	"time"
)

//===========[STATIC/CACHE]====================================================================================================

var uniqueIdCounter int64 = 0

//===========[INTERFACES]====================================================================================================

type Worker interface {
	Id() int
}

//===========[STRUCTS]====================================================================================================

type worker[TWork any] struct {
	//Provides direct access to the worker
	directWork chan TWork

	//A channel of work that this worker will be dealing with
	workBucket chan TWork

	//If this channel is closed, go routine gets terminated
	terminate chan struct{}

	//This function will be processing the work from workBucket
	workHandler func(Worker, TWork)

	//A different function can be set for direct work send to a worker. If not set, it simply uses workHandler
	directWorkHandler func(Worker, TWork)

	//Defines amount of time until this worker quits
	timeout time.Duration

	//When the worker terminates, it sends its ID to this channel
	timedOutWorkers chan int

	//ID of the worker
	id int
}

//DirectWork sends some work directly to this worker
func (w *worker[TWork]) DirectWork(work TWork) {
	w.directWork <- work
}

//Terminate makes this worker redundant
func (w *worker[TWork]) Terminate() {
	w.terminate <- struct{}{}
}

//SetWorkHandler sets a new handler for all the work done
func (w *worker[TWork]) SetWorkHandler(handler func(Worker, TWork)) {
	w.workHandler = handler
}

//SetDirectWorkHandler sets a new worker for all the work that is coming directly for this worker
func (w *worker[TWork]) SetDirectWorkHandler(handler func(Worker, TWork)) {
	w.directWorkHandler = handler
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
		case work := <-w.directWork:
			if w.directWorkHandler == nil {
				w.workHandler(w, work)
				continue
			}
			w.directWorkHandler(w, work)

		case work := <-w.workBucket:
			w.workHandler(w, work)

		case <-time.After(w.timeout):
			w.Terminate()

		case <-w.terminate:
			for {
				select {
				case work := <-w.directWork:
					if w.directWorkHandler == nil {
						w.workHandler(w, work)
						continue
					}
					w.directWorkHandler(w, work)

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
func newWorker[TWork any](workPile chan TWork, workHandler, directWorkHandler func(Worker, TWork), timeout time.Duration, timedOutWorkers chan int) *worker[TWork] {

	defaultWorkHandler := func(Worker, TWork) {}

	w := &worker[TWork]{
		directWork:        make(chan TWork, 2),
		terminate:         make(chan struct{}, 2),
		workBucket:        workPile,
		workHandler:       workHandler,
		directWorkHandler: directWorkHandler,
		timeout:           timeout,
		timedOutWorkers:   timedOutWorkers,
		id:                int(atomic.AddInt64(&uniqueIdCounter, 1)),
	}

	if w.workHandler == nil {
		w.SetWorkHandler(defaultWorkHandler)
	}

	go workerGoroutine(w)

	return w
}
