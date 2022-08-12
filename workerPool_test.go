package workerPool

import (
	"sync"
	"testing"
)

//===========[TESTS]====================================================================================================

func TestNew(t *testing.T) {
	wpNoReq := NewWorkerPool[string](func(Worker, string) {}, nil)
	wpWithReq := NewWorkerPool[string](func(Worker, string) {}, &Requirements{
		MinWorkers:            10,
		MaxWorkers:            100,
		WorkBucketSize:        25,
		WorkerSpawnMultiplier: 10,
	})

	if wpNoReq == nil || wpWithReq == nil {
		t.Errorf("Expected to receive non nil values from function NewWorkerPool(), got %T and %T", wpNoReq, wpWithReq)
	}
}

func TestWorkerPool_WorkerCount(t *testing.T) {
	wp1 := NewWorkerPool[string](func(Worker, string) {}, nil)
	wp2 := NewWorkerPool[string](func(Worker, string) {}, &Requirements{
		MinWorkers:            10,
		MaxWorkers:            20,
		WorkBucketSize:        10,
		WorkerSpawnMultiplier: 5,
	})

	ac1 := wp1.WorkerCount()
	ac2 := wp2.WorkerCount()

	if ac1 != 1 {
		t.Errorf("Expected to have active worker count to be 1, got %d in wp1", ac1)
	}

	if ac2 != 10 {
		t.Errorf("Expected to have active worker count to be 1, got %d in wp2", ac2)
	}
}

//func TestWorkerPool_WorkHandler(t *testing.T) {
//	result := ""
//	wp1 := NewWorkerPool[string](nil)
//	expectedResult := "this_is_working"
//	wg := sync.WaitGroup{}
//
//	wp1.WorkHandler(func(w Worker, work string) {
//		wg.Done()
//		result = work
//	})
//
//	wg.Add(1)
//	wp1.AddWork(expectedResult)
//
//	wg.Wait()
//
//	if result == "" {
//		t.Errorf("Variable result should be \"%s\", got \"%s\"", expectedResult, result)
//	}
//}

func TestWorkerPool_AddWork(t *testing.T) {
	result := ""
	wg := sync.WaitGroup{}
	wp1 := NewWorkerPool[string](func(w Worker, work string) {
		wg.Done()
		result = work
	}, nil)
	expectedResult := "this_is_working"

	wg.Add(1)
	wp1.AddWork(expectedResult)

	wg.Wait()

	if result == "" {
		t.Errorf("Variable result should be \"%s\", got \"%s\"", expectedResult, result)
	}
}

func TestWorker_Id(t *testing.T) {
	wg := sync.WaitGroup{}
	result := 0
	wp := NewWorkerPool[string](func(w Worker, work string) {
		wg.Done()
		result = w.Id()
	}, nil)
	requiredResult := 1

	wg.Add(1)
	wp.AddWork("x1")

	wg.Wait()

	if result != requiredResult {
		t.Errorf("Required result is %d, got %d", requiredResult, result)
	}
}

func TestWorker_SetWorkHandler(t *testing.T) {

}
