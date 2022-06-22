package workerPool

import (
	"github.com/emillis/cacheMachine"
	"time"
)

//===========[STATIC/CACHE]====================================================================================================

var defaultRequirements = Requirements{
	MinWorkers:            1,
	MaxWorkers:            1,
	WorkBucketSize:        10,
	WorkerSpawnMultiplier: 1,
}

//===========[INTERFACES]===================================================================================================

//===========[STRUCTS]====================================================================================================

//Requirements define the rules for worker pool management. Such as number of workers
type Requirements struct {
	//Minimum number of workers that the Hiring Manager must always maintain
	//If set below 1, it will automatically bring MinWorkers count to 1
	MinWorkers int `json:"min_workers" bson:"min_workers"`

	//Maximum number of workers that the Hiring Manager is allowed to hire incomingWork case if workload increases. If
	//MaxWorkers is set below MinWorkers, this will automatically be set to MinWorkers count
	MaxWorkers int `json:"max_workers" bson:"max_workers"`

	//How much work can the channel take incomingWork before starting to block
	WorkBucketSize int `json:"work_bucket_size" bson:"work_bucket_size"`

	//How many workers to spawn every time a shortage of workers is detected. E.g. If you select this to be 10, this
	//means, every time there are not enough workers to handle all the work, there will be another 10 spawned at a time
	//until either they can handle all the work or ceiling of MaxWorkers is reached
	WorkerSpawnMultiplier int `json:"worker_spawn_multiplier" bson:"worker_spawn_multiplier"`
}

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
