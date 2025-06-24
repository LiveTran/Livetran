package ingest

import (
	"context"
	"fmt"
	"sync"
)

type UpdateResponse struct {
	Status 		string
	Update	 	string
}
type Task struct {
	mu 		    sync.Mutex
	Id 			string
	Status		string
	CancelFn	context.CancelCauseFunc
	UpdatesChan	chan UpdateResponse
}

const (
	StreamInit = "INITIALISED"
	StreamReady = "READY"
	StreamStopped = "STOPPED"
	StreamActive = "STREAMING"
)

type TaskManager struct {
	mu		sync.Mutex
	TaskMap	map[string]*Task
}

func NewTaskManager() *TaskManager {
	return &TaskManager{
		TaskMap: make(map[string]*Task),
	}
}

func (task *Task) UpdateStatus(status string, update string) {
	task.mu.Lock()
	defer task.mu.Unlock()

	task.Status = status

	task.UpdatesChan <- UpdateResponse{
		Status: status,
		Update: update,
	}
}



// Starting a Task 
func (tm *TaskManager) StartTask(id string) {
	tm.mu.Lock()
	if _, exists := tm.TaskMap[id]; exists {
		tm.mu.Unlock()
		fmt.Println("Job exists!")
		return 
	}

	cancelCtx, cancelFunc := context.WithCancelCause(context.Background())
	task := &Task{
		Id:          id,
		CancelFn:    cancelFunc,
		Status:      StreamInit,
		UpdatesChan: make(chan UpdateResponse, 4),
	}
	tm.TaskMap[id] = task
	tm.mu.Unlock()

	// Listen for updates
	go func(updates <-chan UpdateResponse) {
		for update := range updates {
			fmt.Println("Update:", update.Update)
			// TODO: sendWebhook
		}
	}(task.UpdatesChan)

	
	go func() {
		SrtConnectionTask(cancelCtx, task)
		tm.StopTask(id, context.Canceled)
	}()
}


// Stopping a task
func (tm *TaskManager) StopTask(id string,reason error) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	if task, exists := tm.TaskMap[id]; exists {
		task.CancelFn(reason)
	} else {
		fmt.Println("Job already done / Cancelled")
	}

}

