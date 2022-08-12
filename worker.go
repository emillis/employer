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

	//This function will be processing the work from workBucket
	workHandler func(Worker, TWork)

	//Holds pointer to the pool this worker resides in
	workerPool *WorkerPool[TWork]

	timeout time.Duration

	//ID of the worker
	id int
}

//spawnGoroutine spawns a new goroutine that monitors this worker
func (w *worker[TWork]) spawnGoroutine() {
	go func() {
		defer w.workerPool.workers.Remove(w.id)

		exit := make(chan struct{})

		t := time.AfterFunc(w.timeout, func() {
			exit <- struct{}{}
		})

		for {
			select {
			case work := <-w.workBucket:
				stoppedOnTime := t.Stop()
				w.workHandler(w, work)
				//If this check is not made, you can restart the timer after this goroutine has already exited
				//and once the timer fires it will try to send signal down the channel that nobody is listening to
				if !stoppedOnTime {
					continue
				}
				t.Reset(w.timeout)

			case <-exit:
				break
			}
		}
	}()
}

//===========[PUBLIC]====================================================================================================

//Id returns this worker's ID
func (w *worker[TWork]) Id() int {
	return w.id
}

//===========[FUNCTIONALITY]====================================================================================================

//issueNewWorkerId returns an int incremented by one every time this function is invoked
func issueNewWorkerId() int {
	return int(atomic.AddInt64(&uniqueIdCounter, 1))
}
