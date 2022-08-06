package workerPool

import "testing"

//===========[TESTS]====================================================================================================

func TestNew(t *testing.T) {
	wpNoReq := New[string](nil)
	wpWithReq := New[string](&Requirements{
		MinWorkers:            10,
		MaxWorkers:            100,
		WorkBucketSize:        25,
		WorkerSpawnMultiplier: 10,
	})

	if wpNoReq == nil || wpWithReq == nil {
		t.Errorf("Expected to receive non nil values from function New(), got %T and %T", wpNoReq, wpWithReq)
	}
}