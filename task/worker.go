package task

import (
	logger "github.com/sirupsen/logrus"
)

// Worker represents the worker that executes the task
type Worker struct {
	id              int
	pool            chan chan Task
	taskChannel     chan Task
	statusChannel   chan<- TaskStatus
	completeChannel chan<- Task
	errorChannel    chan<- Failure
	stoppedChannel  chan bool
}

// NewWorker - Creates a new Worker
func NewWorker(id int, pool chan chan Task, statusChannel chan<- TaskStatus, completeChannel chan<- Task, errorChannel chan<- Failure) *Worker {
	w := new(Worker)
	w.id = id
	w.pool = pool
	w.statusChannel = statusChannel
	w.completeChannel = completeChannel
	w.errorChannel = errorChannel
	return w
}

// Start method starts the run loop for the worker, listening for a quit channel
// in case we need to stop it
func (w *Worker) Start() {
	w.taskChannel = make(chan Task)
	w.stoppedChannel = make(chan bool, 1)
	w.pool <- w.taskChannel
	go w.run()
}

func (w *Worker) run() {
	// Pull tasks from this Worker's task channel. This loop will complete when
	// the taskChannel is closed
	for task := range w.taskChannel {
		logger.Infof("Worker %d: Got task with id: %s", w.id, task.ID())
		if err := task.Execute(w.statusChannel); err != nil {
			w.errorChannel <- Failure{Task: task, Error: err}
		} else {
			w.completeChannel <- task
		}
		w.pool <- w.taskChannel
	}
	w.stoppedChannel <- true
}

// Stop signals the worker to stop listening for work requests.
func (w *Worker) Stop() {
	close(w.taskChannel)
	<-w.stoppedChannel
	logger.Debugf("Stopped worker %d", w.id)
}
