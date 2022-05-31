package employer

//===========[STATIC/CACHE]====================================================================================================

const ()

var defaultRequirements = Requirements{
	MinWorkers: 1,
	MaxWorkers: 1,
}

//===========[STRUCTS]====================================================================================================

//Requirements define the rules for Hiring Manager. Such as number of workers
type Requirements struct {
	//Minimum number of workers that the Hiring Manager must always maintain
	MinWorkers int `json:"min_workers" bson:"min_workers"`

	//Maximum number of workers that the Hiring Manager is allowed to hire in case if workload increases
	MaxWorkers int `json:"max_workers" bson:"max_workers"`
}

type HiringManager[TWork any] struct {
	Requirements //TODO: make immutable

	//The channel that all workers will get the jobs from
	in chan<- TWork

	//The channel that all the workers will send results to
	out <-chan TWork

	//Channel that if closed, terminates all the workers
	terminateAllWorkers chan struct{}
}

//init initializes the HiringManager based on Requirements provided
func (hm *HiringManager[TWork]) init() {
	hm.in = make(chan TWork)
	hm.out = make(chan TWork)
	hm.terminateAllWorkers = make(chan struct{})

}

//===========[FUNCTIONALITY]====================================================================================================

func New[TWork any](r *Requirements) *HiringManager[TWork] {
	if r == nil {
		r = &defaultRequirements
	}

	hm := &HiringManager[TWork]{
		Requirements: *r,
	}

	hm.init()

	return hm
}
