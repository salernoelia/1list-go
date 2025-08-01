package main

type Task struct {
	Comment          string `json:"comment"`
	CommentDisplayed bool   `json:"commentDisplayed"`
	Done             bool   `json:"done"`
	ID               int64  `json:"id"`
	Title            string `json:"title"`
}

type TaskList struct {
	Items []Task `json:"items"`
	Title string `json:"title"`
}

type Config struct {
	TaskFolder string `json:"task_folder"`
}