package task

// TaskStatus represents a status update from a Task
type TaskStatus struct {
	Task    Task
	Message string
}

// Failure represents a failed task
type Failure struct {
	Task  Task
	Error error
}

// Task represents the task to be run
type Task interface {
	Execute(status chan<- TaskStatus) error
	ID() string
}
