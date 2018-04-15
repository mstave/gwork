package worker

import (
	"runtime"
)
// Pool Has variable number of Workers to work on jobs
type Pool struct {
	JobChannel        chan Job
	StopChannel       chan bool
	workerConcurrency int
}

// NewWorkerPool constructor for convenience
func NewWorkerPool(workerConcurrency int) *Pool {
	jobQueue := make(chan Job, 1)
	stopQ := make(chan bool, 1)
	return &Pool{jobQueue, stopQ,  workerConcurrency}
}

// AddJob Don't make me describe this simple function, domineering linter.
func (d *Pool) AddJob(j Job) {
	d.JobChannel <- j
}
// StartWorkers kick them off
func (d *Pool) StartWorkers() {
	for i := 0; i < d.workerConcurrency; i++ {
		worker := NewWorker(d)
		worker.Start()
	}
}
// DrainPool Tell workers to stop doing stuff
func (d *Pool) DrainPool() {
	d.StopChannel <- true
}
// RunJobs Convenience method - hand jobs to pool of GOMAXPROCS workers and starts them off
func RunJobs( jobs ... Job) *Pool {
	// Can throttle for now with GOMAXPROCS
	workerConcurrency := runtime.NumCPU()

	pool := NewWorkerPool(workerConcurrency)
	for _,j := range jobs {
		go pool.AddJob(j)
	}
	pool.StartWorkers()
	return pool
}
