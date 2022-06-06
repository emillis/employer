package employer

import (
	"sync"
	"sync/atomic"
)

//===========[STATIC/CACHE]====================================================================================================

var uniqueIdCounter int64 = 0

//===========[STRUCTS]====================================================================================================

type worker[TWork any] struct {
	//Provides direct access to the worker
	directAccess chan TWork

	//If this channel is closed, go routine gets terminated
	fire chan struct{}

	//ID of the worker
	id int
}

//DirectWork sends some work directly to this worker
func (w *worker[TWork]) DirectWork(work TWork) {
	w.directAccess <- work
}

//Fire makes this worker redundant
func (w *worker[TWork]) Fire() {
	close(w.fire)
}

//===========[FUNCTIONALITY]====================================================================================================

//workerGoroutine is the goroutine that will be spawned for each worker. As an argument you must pass in worker struct
func workerGoroutine[TWork any](w *worker[TWork], workPile chan TWork) {
	if w == nil {
		return
	}

	for {
		select {
		case work := <-w.directAccess:
		case work := <-workPile:
		case <-w.fire:
			break
		}
	}
}

func newWorker[TWork any](workPile chan TWork) *worker[TWork] {

	w := &worker[TWork]{
		directAccess: make(chan TWork, 2),
		fire:         make(chan struct{}),
		id:           int(atomic.AddInt64(&uniqueIdCounter, 1)),
	}

	go workerGoroutine(w, workPile)

	return w
}
