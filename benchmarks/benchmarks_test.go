package benchmarks

import (
	"github.com/emillis/workerPool"
	"testing"
)

var pool *workerPool.WorkerPool[int]

func BenchmarkThroughput(b *testing.B) {
	for n := 0; n < b.N; n++ {
		pool.AddWork(n)
	}
}

func init() {
	pool = workerPool.New[int](&workerPool.Requirements{
		MinWorkers:            10,
		MaxWorkers:            100,
		WorkBucketSize:        500,
		WorkerSpawnMultiplier: 10,
	})

	pool.WorkHandler(func(n ...int) {})
}
