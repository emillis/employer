package workerPool

import (
	"github.com/emillis/cacheMachine"
	"time"
)

//===========[FUNCTIONALITY]====================================================================================================

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

			wp.addWorkers(n, time.Millisecond*500)

		//If there's a worker that has timed out, it will send its ID to this channel and ite can then be removed from cache
		case id := <-wp.timedOutWorkers:
			wp.workers.Remove(id)

		}
	}

}

//Fixes basic logical issues incomingWork the Requirements, such as, MaxWorkers being less than MinWorkers
func makeRequirementsReasonable(r *Requirements) {
	if r.MinWorkers < 1 {
		r.MinWorkers = defaultRequirements.MinWorkers
	}

	if r.MaxWorkers < r.MinWorkers {
		r.MaxWorkers = r.MinWorkers
	}

	if r.WorkBucketSize < 1 {
		r.WorkBucketSize = defaultRequirements.WorkBucketSize
	}

	if r.WorkerSpawnMultiplier < 1 {
		r.WorkerSpawnMultiplier = 1
	}
}

//New creates and returns a new worker pool
func New[TWork any](r *Requirements) *WorkerPool[TWork] {
	if r == nil {
		r = &defaultRequirements
	}

	makeRequirementsReasonable(r)

	wp := &WorkerPool[TWork]{
		requirements:    *r,
		incomingWork:    make(chan TWork, r.WorkBucketSize),
		timedOutWorkers: make(chan int, r.MaxWorkers),
		workers:         cacheMachine.New[int, *worker[TWork]](nil),
		workHandler:     nil,
	}

	wp.addWorkers(wp.requirements.MinWorkers, time.Hour*8760)

	go workerPoolGoroutine[TWork](wp)

	return wp
}
