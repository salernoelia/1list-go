package main

import (
	"time"
)

type Task struct {
	ID               int64      `json:"id"`
	Title            string     `json:"title"`
	Status           TaskStatus `json:"status"`
	Comment          string     `json:"comment"`
	Sessions         []Session  `json:"sessions"`
	TotalDuration    int64      `json:"total_duration"`
	ActiveStartTime  *time.Time `json:"active_start_time,omitempty"`
	CompletedAt      *time.Time `json:"completed_at,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
}

type Session struct {
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	Duration  int64     `json:"duration"`
}

type TaskStatus string

const (
	StatusPending TaskStatus = "pending"
	StatusActive  TaskStatus = "active"
	StatusPaused  TaskStatus = "paused"
	StatusDone    TaskStatus = "done"
)

type TaskList struct {
	Title     string `json:"title"`
	Items     []Task `json:"items"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Config struct {
	TaskDir string `json:"task_folder"`
}

func (t *Task) IsActive() bool {
	return t.Status == StatusActive
}

func (t *Task) IsDone() bool {
	return t.Status == StatusDone
}

func (t *Task) GetFormattedDuration() string {
	return formatDuration(t.TotalDuration)
}
