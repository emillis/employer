package workerPool

import (
	"github.com/emillis/cacheMachine"
	"time"
)

//===========[STRUCTS]====================================================================================================

//WorkerPool provides the main public API to worker pools
type WorkerPool[TWork any] struct {
	requirements Requirements

	//The channel that all workers will get the jobs from
	incomingWork chan TWork

	//If a worker times out, it will send its ID via this channel
	timedOutWorkers chan int

	//pool of the actual workers
	workers cacheMachine.Cache[int, *worker[TWork]]

	//This will be passed to each worker to use for work processing
	workHandler func(Worker, TWork)
}

//------PRIVATE------

//addWorkers add n number of workers to the pool
func (wp *WorkerPool[TWork]) addWorkers(n int, timeout time.Duration) {
	for i := 0; i < n; i++ {
		w := newWorker[TWork](wp.incomingWork, wp.workHandler, nil, timeout, wp.timedOutWorkers)

		wp.workers.Add(w.id, w)
	}
}

//terminateWorkers removes a number of workers specified in the argument
func (wp *WorkerPool[TWork]) terminateWorkers(n int) {
	//if n < 1 {
	//	return
	//}
	//
	//count := wp.workers.Count()
	//if n > count {
	//	n = count
	//}
	//
	//for i := 0; i < n; i++ {
	//	wp.workerTerminator <- struct{}{}
	//}
}

//------PUBLIC------

//WorkHandler is a function that every worker will use to primarily process all incoming work
//Note, you can set WorkHandler only once. Once set, the subsequent calls to change it will be ignored
func (wp *WorkerPool[TWork]) WorkHandler(f func(Worker, TWork)) {
	if wp.workHandler != nil {
		return
	}

	wp.workHandler = f

	for _, worker := range wp.workers.GetAll() {
		worker.SetWorkHandler(f)
	}
}

//AddWork sends work to workers
func (wp *WorkerPool[TWork]) AddWork(w TWork) {
	wp.incomingWork <- w
}

//ActiveWorkerCount returns number of active workers in the worker pool
func (wp *WorkerPool[TWork]) ActiveWorkerCount() int {
	return wp.workers.Count()
}