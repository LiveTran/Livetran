package ingest

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"sync"
	"time"
)

type UpdateResponse struct {
	Status 		string
	Update	 	string
}

type Task struct {
	mu 		    sync.Mutex
	Id 			string
	Status		string
	Webhooks 	[]string
	Abr			bool
	CancelFn	context.CancelCauseFunc
	UpdatesChan	chan UpdateResponse
	StartTime	time.Time
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

func (tm *TaskManager) GetAllStreams() (active,idle,stopped int64) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	for _, stream := range tm.TaskMap {
		switch stream.Status {
			case StreamActive:
				active++
			case StreamStopped:
				stopped++;
			default:
				idle++;
		}
	}

	return active,idle,stopped
}


// Starting a Task 
func (tm *TaskManager) StartTask(id string,webhooks []string, abr bool) {
	tm.mu.Lock()
	if _, exists := tm.TaskMap[id]; exists {
		tm.mu.Unlock()
		slog.Error("Job Exists!");
		return 
	}

	cancelCtx, cancelFunc := context.WithCancelCause(context.Background())
	task := &Task{
		Id:          id,
		CancelFn:    cancelFunc,
		Status:      StreamInit,
		Webhooks: 	 webhooks,
		Abr:		 abr || false,
		UpdatesChan: make(chan UpdateResponse, 4),
		StartTime:   time.Now(), 
	}
	tm.TaskMap[id] = task
	tm.mu.Unlock()

	// Listen for updates
	go func(updates <-chan UpdateResponse) {
		for update := range updates {

			slog.Info(update.Update)

			jsonData, err := json.Marshal(update)
			if err != nil {
				slog.Error("Failed to send webhook", "error", err);
				continue
			}

			for _,webhook := range task.Webhooks {
				resp,err := http.Post(webhook,"application/json",bytes.NewBuffer(jsonData))
				if err != nil {
					slog.Error("Failed to send webhook", "error", err);
					continue
				}
				_, _ = io.Copy(io.Discard, resp.Body)
				resp.Body.Close()
			}
			
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
		slog.Error("Job already done / Cancelled");
	}

}

