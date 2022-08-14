package workerPool

import (
	"sync"
	"testing"
	"time"
)

//===========[TESTS]====================================================================================================

func TestNew(t *testing.T) {
	wpNoReq := NewWorkerPool[string](func(string) {}, nil)
	wpWithReq := NewWorkerPool[string](func(string) {}, &Requirements{
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
	wp1 := NewWorkerPool[string](func(string) {}, nil)
	wp2 := NewWorkerPool[string](func(string) {}, &Requirements{
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

func TestWorkerPool_AddWork(t *testing.T) {
	result := ""
	wg := sync.WaitGroup{}
	wp1 := NewWorkerPool[string](func(work string) {
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

func TestWorker_SetWorkHandler(t *testing.T) {

}

func TestWorkerPool_AutoWorkerManagement(t *testing.T) {
	wp := NewWorkerPool[int](func(work int) {
		time.Sleep(time.Millisecond * 25)
	}, &Requirements{
		MinWorkers:            1,
		MaxWorkers:            25,
		WorkBucketSize:        500,
		WorkerSpawnMultiplier: 25,
		Timeout:               time.Millisecond * 50,
	})

	initialPoolSize := wp.WorkerCount()

	for i := 0; i < 100; i++ {
		wp.AddWork(i)
	}

	time.Sleep(time.Millisecond * 100)

	midPoolSize := wp.WorkerCount()

	time.Sleep(time.Millisecond * 250)

	endPoolSize := wp.WorkerCount()

	if initialPoolSize != 1 {
		t.Errorf("Initially expected to have 1 worker, got %d", initialPoolSize)
	}

	if midPoolSize != 25 {
		t.Errorf("Mid way through the test expected to have 25 workers, got %d", midPoolSize)
	}

	if endPoolSize != 1 {
		t.Errorf("At the end of the test expected to have the worker count dropped back to 1, got %d", endPoolSize)
	}
}
