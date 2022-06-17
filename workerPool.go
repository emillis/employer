package workerPool

import (
	"github.com/emillis/cacheMachine"
)

//===========[STATIC/CACHE]====================================================================================================

var defaultRequirements = Requirements{
	MinWorkers:     1,
	MaxWorkers:     1,
	WorkBucketSize: 10,
}

//===========[INTERFACES]====================================================================================================

type Worker interface {
}

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
}

//WorkerPool provides the main public API to worker pools
type WorkerPool[TWork any] struct {
	requirements Requirements

	//The channel that all workers will get the jobs from
	incomingWork chan TWork

	//The channel that all the workers will send results to
	out chan TWork

	//Channel that if closed, terminates all the workers
	terminateAllWorkers chan struct{}

	//pool of the actual workers
	workers cacheMachine.Cache[int, *worker[TWork]]

	//This will be passed to each worker to use for work processing
	workHandler func(...TWork)
}

//------PRIVATE------

//addWorkers add n number of workers to the pool
func (wp *WorkerPool[TWork]) addWorkers(n int) {
	for i := 0; i < n; i++ {
		w := newWorker[TWork](wp.incomingWork, wp.workHandler, nil, wp.terminateAllWorkers)

		wp.workers.Add(w.id, w)
	}
}

//terminateWorkers removes a number of workers specified in the argument
func (wp *WorkerPool[TWork]) terminateWorkers(n int) {
	if n < 1 {
		return
	}

	count := wp.workers.Count()
	if n > count {
		n = count
	}

	for i := 0; i < n; i++ {
		wp.terminateAllWorkers <- struct{}{}
	}
}

//------PUBLIC------

//WorkHandler is a function that every worker will use to primarily process all incoming work
//Note, you can set WorkHandler only once. Once set, the subsequent calls to change it will be ignored
func (wp *WorkerPool[TWork]) WorkHandler(f func(w ...TWork)) {
	if wp.workHandler != nil {
		return
	}

	wp.workHandler = f
}

//===========[FUNCTIONALITY]====================================================================================================

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
}

//New creates and returns a new worker pool
func New[TWork any](r *Requirements) *WorkerPool[TWork] {
	if r == nil {
		r = &defaultRequirements
	}

	makeRequirementsReasonable(r)

	wp := &WorkerPool[TWork]{
		requirements:        *r,
		incomingWork:        make(chan TWork, r.WorkBucketSize),
		out:                 make(chan TWork),
		terminateAllWorkers: make(chan struct{}, r.MaxWorkers),
		workers:             cacheMachine.New[int, *worker[TWork]](nil),
		workHandler:         func(w ...TWork) {},
	}

	wp.addWorkers(wp.requirements.MinWorkers)

	return wp
}
