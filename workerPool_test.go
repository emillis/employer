package workerPool

import (
	"sync"
	"testing"
)

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

func TestWorkerPool_ActiveWorkerCount(t *testing.T) {
	wp1 := New[string](nil)
	wp2 := New[string](&Requirements{
		MinWorkers:            10,
		MaxWorkers:            20,
		WorkBucketSize:        10,
		WorkerSpawnMultiplier: 5,
	})

	ac1 := wp1.ActiveWorkerCount()
	ac2 := wp2.ActiveWorkerCount()

	if ac1 != 1 {
		t.Errorf("Expected to have active worker count to be 1, got %d in wp1", ac1)
	}

	if ac2 != 10 {
		t.Errorf("Expected to have active worker count to be 1, got %d in wp2", ac2)
	}
}

func TestWorkerPool_WorkHandler(t *testing.T) {
	result := ""
	wp1 := New[string](nil)
	expectedResult := "this_is_working"
	wg := sync.WaitGroup{}

	wp1.WorkHandler(func(w Worker, work string) {
		wg.Done()
		result = work
	})

	wg.Add(1)
	wp1.AddWork(expectedResult)

	wg.Wait()

	if result == "" {
		t.Errorf("Variable result should be \"%s\", got \"%s\"", expectedResult, result)
	}
}

func TestWorkerPool_AddWork(t *testing.T) {
	result := ""
	wp1 := New[string](nil)
	expectedResult := "this_is_working"
	wg := sync.WaitGroup{}

	wp1.WorkHandler(func(w Worker, work string) {
		wg.Done()
		result = work
	})

	wg.Add(1)
	wp1.AddWork(expectedResult)

	wg.Wait()

	if result == "" {
		t.Errorf("Variable result should be \"%s\", got \"%s\"", expectedResult, result)
	}
}
