package workerPool

import "github.com/emillis/cacheMachine"

//===========[STATIC/CACHE]====================================================================================================

var defaultRequirements = Requirements{
	MinWorkers: 1,
	MaxWorkers: 1,
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

	//Maximum number of workers that the Hiring Manager is allowed to hire in case if workload increases. If
	//MaxWorkers is set below MinWorkers, this will automatically be set to MinWorkers count
	MaxWorkers int `json:"max_workers" bson:"max_workers"`

	//How much work can the channel take in before starting to block
	WorkBucketSize int `json:"work_bucket_size" bson:"work_bucket_size"`
}

type WorkerPool[TWork any] struct {
	Requirements //TODO: make immutable

	//The channel that all workers will get the jobs from
	in chan TWork

	//The channel that all the workers will send results to
	out chan TWork

	//Channel that if closed, terminates all the workers
	terminateAllWorkers chan struct{}

	//pool of the actual workers
	workers cacheMachine.Cache[int, *worker[TWork]]
}

func (wp *WorkerPool[TWork]) addWorkers(n int) {
	for i := 0; i < n; i++ {
		w := newWorker[TWork](wp.in, nil, nil)

		wp.workers.Add(w.id, w)
	}
}

//===========[FUNCTIONALITY]====================================================================================================

//Fixes basic logical issues in the Requirements, such as, MaxWorkers being less than MinWorkers
func makeRequirementsReasonable(r *Requirements) {
	if r.MinWorkers < 1 {
		r.MinWorkers = 1
	}

	if r.MaxWorkers < r.MinWorkers {
		r.MaxWorkers = r.MinWorkers
	}
}

//New creates and returns a new worker pool
func New[TWork any](r *Requirements) *WorkerPool[TWork] {
	if r == nil {
		r = &defaultRequirements
	}

	makeRequirementsReasonable(r)

	wp := &WorkerPool[TWork]{
		Requirements:        *r,
		in:                  make(chan TWork),
		out:                 make(chan TWork),
		terminateAllWorkers: make(chan struct{}),
		workers:             cacheMachine.New[int, *worker[TWork]](nil),
	}

	wp.addWorkers(wp.Requirements.MinWorkers)

	return wp
}
