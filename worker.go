package employer

import "sync"

//===========[STATIC/CACHE]====================================================================================================

//===========[STRUCTS]====================================================================================================

type worker[TWork any] struct {
	//Provides direct access to the worker
	directAccess chan <- TWork

	//Fire this worker (terminates the go routine)
	fire chan struct{}

	//ID of the worker
	id int

	//Locks for this worker
	mx sync.RWMutex
}

func (w *worker[TW])


//===========[FUNCTIONALITY]====================================================================================================

func newWorker[TWork any]() *worker[TWork] {
	w := &worker[TWork]{
		directAccess: make(chan <- TWork),
		fire: make(chan struct{}),
	}

	return w
}