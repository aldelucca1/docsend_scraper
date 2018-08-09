package task

import (
	"sync"

	logger "github.com/sirupsen/logrus"
)

// Dispatcher - A Task dispatcher that will dispatch work to a set of Workers.
// If no Worker is avaialable, this Dispatcher will block adding new tasks until
// a Worker becomes available
type Dispatcher struct {
	workerPool      chan chan Task
	statusChannel   chan TaskStatus
	completeChannel chan Task
	errorChannel    chan Failure
	workers         []*Worker
}

// NewDispatcher - Creates a new Dispatcher with the supplied number of workers
func NewDispatcher(maxWorkers int) *Dispatcher {
	d := new(Dispatcher)
	d.workerPool = make(chan chan Task, maxWorkers)
	d.statusChannel = make(chan TaskStatus)
	d.completeChannel = make(chan Task)
	d.errorChannel = make(chan Failure)
	d.workers = make([]*Worker, 0, maxWorkers)
	for i := 0; i < maxWorkers; i++ {
		d.workers = append(d.workers, NewWorker(i, d.workerPool, d.statusChannel, d.completeChannel, d.errorChannel))
	}
	return d
}

// Status - Returns the status channel
func (d *Dispatcher) Status() <-chan TaskStatus {
	return d.statusChannel
}

// Complete - Returns the completion channel
func (d *Dispatcher) Complete() <-chan Task {
	return d.completeChannel
}

// Error - Returns the error channel
func (d *Dispatcher) Error() <-chan Failure {
	return d.errorChannel
}

// Start - Starts this dispatcher and all associated Workers
func (d *Dispatcher) Start() {
	logger.Debugf("Starting %d workers...", len(d.workers))
	for _, worker := range d.workers {
		worker.Start()
	}
}

// Stop - Stops this Dispatcher waiting for all tasks to complete
func (d *Dispatcher) Stop() {

	logger.Info("Stopping dispatcher")

	// Stop the workers
	logger.Info("Waiting for all workers to stop...")
	d.stopWorkers()

	// Close the Complete and Error channels
	close(d.statusChannel)
	close(d.completeChannel)
	close(d.errorChannel)

	logger.Info("Dispatcher stopped")
}

func (d *Dispatcher) stopWorkers() {

	// Create a WaitGroup, and add the worker count
	var wg sync.WaitGroup
	wg.Add(len(d.workers))

	// Call Stop on each work and notify the WaitGroup when done
	for _, worker := range d.workers {
		go func(worker *Worker) {
			worker.Stop()
			wg.Done()
		}(worker)
	}

	// Wait for all the Wokers to return
	wg.Wait()
}

// Dispatch - Add a Task to the list of pending tasks
func (d *Dispatcher) Dispatch(task Task) {
	// try to obtain a worker task channel that is available, this will block
	// until a worker is idle
	taskChannel := <-d.workerPool

	// dispatch the task to the worker task channel
	taskChannel <- task
}

// NonBlockingDispatcher - A Dispatcher that will not block when adding new
// tasks, rather it will queue the tasks until a worker becomes available
type NonBlockingDispatcher struct {
	*Dispatcher
	taskQueue chan Task
}

// NewNonBlockingDispatcher - Create a new non-blocking Dispatcher
func NewNonBlockingDispatcher(maxWorkers int) *NonBlockingDispatcher {
	d := new(NonBlockingDispatcher)
	d.Dispatcher = NewDispatcher(maxWorkers)
	d.taskQueue = make(chan Task)
	return d
}

// Start - Starts this dispatcher and all associated Workers
func (d *NonBlockingDispatcher) Start() {
	d.Dispatcher.Start()
	go d.dispatch()
}

// Stop - Stops this Dispatcher waiting for all tasks to complete
func (d *NonBlockingDispatcher) Stop() {

	// Close the task queue
	close(d.taskQueue)

	d.Dispatcher.Stop()
}

// Dispatch - Add a Task to the list of pending tasks
func (d *NonBlockingDispatcher) Dispatch(task Task) {
	d.taskQueue <- task
}

// dispatch - Pulls tasks off the pending task queue and executes them
func (d *NonBlockingDispatcher) dispatch() {
	// Pull tasks from the taskQueue.  This loop will complete when the taskQueue
	// chan is closed
	for task := range d.taskQueue {
		// try to obtain a worker task channel that is available, this will block
		// until a worker is idle
		taskChannel := <-d.workerPool

		// dispatch the task to the worker task channel
		taskChannel <- task
	}
}
