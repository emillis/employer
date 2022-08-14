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

	//pool of the actual workers
	workers cacheMachine.Cache[int, *worker[TWork]]

	//This will be passed to each worker to use for work processing
	workHandler func(TWork)
}

//------PRIVATE------

//addWorkers add n number of workers to the pool
func (wp *WorkerPool[TWork]) addWorkers(n int, timeout time.Duration) {
	for i := 0; i < n; i++ {

		w := &worker[TWork]{
			workBucket:  wp.incomingWork,
			workHandler: wp.workHandler,
			timeout:     timeout,
			workerPool:  wp,
			id:          issueNewWorkerId(),
		}

		wp.workers.Add(w.id, w)

		w.spawnGoroutine()
	}
}

//------PUBLIC------

//AddWork sends work to workers
func (wp *WorkerPool[TWork]) AddWork(w TWork) {
	wp.incomingWork <- w
}

//WorkerCount returns number of active workers in the worker pool
func (wp *WorkerPool[TWork]) WorkerCount() int {
	return wp.workers.Count()
}

//===========[FUNCTIONS]================================================================================================

//This is called as goroutine and it overseas a single WorkerPool element
func workerPoolGoroutine[TWork any](wp *WorkerPool[TWork]) {
	if wp == nil {
		return
	}

	var length, remainingPoolCapacity int
	n := wp.requirements.WorkerSpawnMultiplier

	for {
		select {
		//Spawning more workers if there is too much work
		case <-time.After(time.Microsecond * 100):
			length = len(wp.incomingWork) //TODO: Benchmark these. Perhaps not assigning variables is faster?
			if length <= wp.requirements.MinWorkers {
				continue
			}

			remainingPoolCapacity = wp.requirements.MaxWorkers - wp.workers.Count()
			n = wp.requirements.WorkerSpawnMultiplier

			if n > remainingPoolCapacity {
				n = remainingPoolCapacity
			}

			wp.addWorkers(n, wp.requirements.Timeout)
		}
	}

}

//NewWorkerPool creates and returns a new worker pool
func NewWorkerPool[TWork any](workHandler func(TWork), r *Requirements) *WorkerPool[TWork] {
	if r == nil {
		r = &defaultRequirements
	} else {
		makeRequirementsReasonable(r)
	}

	wp := &WorkerPool[TWork]{
		requirements: *r,
		incomingWork: make(chan TWork, r.WorkBucketSize),
		workers:      cacheMachine.New[int, *worker[TWork]](nil),
		workHandler:  workHandler,
	}

	wp.addWorkers(wp.requirements.MinWorkers, time.Hour*8760)

	go workerPoolGoroutine[TWork](wp)

	return wp
}
