package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

func findTaskFiles(folder string) ([]string, error) {
	if folder == "" {
		return nil, fmt.Errorf("no task folder configured")
	}

	entries, err := os.ReadDir(folder)
	if err != nil {
		return nil, fmt.Errorf("cannot read folder %s: %v", folder, err)
	}

	var taskFiles []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".json") {
			taskFiles = append(taskFiles, entry.Name())
		}
	}

	if len(taskFiles) == 0 {
		
		return nil, fmt.Errorf("no task lists found")
	}

	return taskFiles, nil
}

func selectTaskFile(folder string, taskFiles []string) (string, error) {
	fmt.Printf("📋 Available task lists (%d):\n\n", len(taskFiles))
	for i, file := range taskFiles {
		displayName := strings.TrimSuffix(file, ".json")
		fmt.Printf("  %d. %s\n", i+1, displayName)
	}

	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Printf("\nSelect list (1-%d), create 'c <name>', or remove 'r <number>': ", len(taskFiles))
		if !scanner.Scan() {
			return "", fmt.Errorf("input error")
		}
		input := strings.TrimSpace(scanner.Text())

		if strings.HasPrefix(input, "c ") {
			listName := strings.TrimSpace(input[2:])
			if listName == "" {
				fmt.Println("❌ List name required")
				continue
			}
			if err := createNewList(folder, listName); err != nil {
				fmt.Printf("❌ %v\n", err)
				continue
			}
			fmt.Printf("✅ Created: %s\n", listName)
			taskFiles, err := findTaskFiles(folder)
			if err != nil {
				return "", err
			}
			displayTaskFiles(taskFiles)
			continue
		}

		if strings.HasPrefix(input, "r ") {
			numStr := strings.TrimSpace(input[2:])
			choice, err := strconv.Atoi(numStr)
			if err != nil || choice < 1 || choice > len(taskFiles) {
				fmt.Println("❌ Invalid selection")
				continue
			}
			selectedFile := taskFiles[choice-1]
			fmt.Printf("Remove '%s'? (y/N): ", selectedFile)
			if scanner.Scan() && strings.ToLower(scanner.Text()) == "y" {
				if err := os.Remove(filepath.Join(folder, selectedFile)); err != nil {
					fmt.Printf("❌ Failed to remove: %v\n", err)
					continue
				}
				fmt.Printf("✅ Removed: %s\n", selectedFile)
				taskFiles, err = findTaskFiles(folder)
				if err != nil {
					return "", err
				}
				if len(taskFiles) == 0 {
					return "", fmt.Errorf("no task lists found")
				}
				displayTaskFiles(taskFiles)
			}
			continue
		}

		choice, err := strconv.Atoi(input)
		if err != nil || choice < 1 || choice > len(taskFiles) {
			fmt.Println("❌ Invalid selection")
			continue
		}
		return filepath.Join(folder, taskFiles[choice-1]), nil
	}
}

func displayTaskFiles(taskFiles []string) {
	fmt.Printf("\n📋 Available task lists (%d):\n\n", len(taskFiles))
	for i, file := range taskFiles {
		displayName := strings.TrimSuffix(file, ".json")
		fmt.Printf("  %d. %s\n", i+1, displayName)
	}
}

func loadTasks(filePath string) (*TaskList, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var taskList TaskList
	err = json.Unmarshal(data, &taskList)
	return &taskList, err
}

func saveTasks(filePath string, taskList *TaskList) error {
	taskList.UpdatedAt = time.Now()
	data, err := json.MarshalIndent(taskList, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filePath, data, 0644)
}

func createNewList(folder string, listName string) error {
	if strings.TrimSpace(listName) == "" {
		return fmt.Errorf("list name cannot be empty")
	}

	sanitizedName := strings.ToLower(listName)
	sanitizedName = strings.Map(func(r rune) rune {
		if r == ' ' {
			return '-'
		}
		if r == '-' || r == '_' ||
			(r >= 'a' && r <= 'z') ||
			(r >= '0' && r <= '9') {
			return r
		}
		return -1
	}, sanitizedName)

	fileName := fmt.Sprintf("%s.json", sanitizedName)
	filePath := filepath.Join(folder, fileName)

	if _, err := os.Stat(filePath); err == nil {
		return fmt.Errorf("list '%s' already exists", listName)
	}

	now := time.Now()
	newTaskList := &TaskList{
		Title:     listName,
		Items:     []Task{},
		CreatedAt: now,
		UpdatedAt: now,
	}

	return saveTasks(filePath, newTaskList)
}

func displayTaskList(taskList *TaskList, fileName string) {
	listName := strings.TrimSuffix(fileName, ".json")
	fmt.Printf("┌─ 📋 %s\n", listName)
	fmt.Printf("├─ %s\n", strings.Repeat("─", len(listName)+4))

	activeCount := 0
	pendingCount := 0
	doneCount := 0

	for _, task := range taskList.Items {
		switch task.Status {
		case StatusActive:
			activeCount++
		case StatusPending, StatusPaused:
			pendingCount++
		case StatusDone:
			doneCount++
		}
	}

	fmt.Printf("├─ Active: %d │ Pending: %d │ Done: %d\n", activeCount, pendingCount, doneCount)
	fmt.Printf("└─ %s\n\n", strings.Repeat("─", 40))

	if activeCount > 0 {
		fmt.Println("🔴 ACTIVE TASKS:")
		displayTasksByStatus(taskList, StatusActive)
		fmt.Println()
	}

	if pendingCount > 0 {
		fmt.Println("⏸️ PENDING TASKS:")
		displayTasksByStatus(taskList, StatusPending, StatusPaused)
		fmt.Println()
	}

	if doneCount > 0 {
		fmt.Println("✅ COMPLETED TASKS:")
		displayTasksByStatus(taskList, StatusDone)
		fmt.Println()
	}

	fmt.Println("💡 Commands: <number> (start/stop), add <task>, remove <number>, done <number>, r (return), q (quit)")
}

func displayTasksByStatus(taskList *TaskList, statuses ...TaskStatus) {
	statusMap := make(map[TaskStatus]bool)
	for _, status := range statuses {
		statusMap[status] = true
	}

	for i, task := range taskList.Items {
		if !statusMap[task.Status] {
			continue
		}

		var statusIcon string
		var timeInfo string

		switch task.Status {
		case StatusActive:
			statusIcon = "🟢"
			if task.ActiveStartTime != nil {
				elapsed := time.Since(*task.ActiveStartTime)
				timeInfo = fmt.Sprintf(" [Running: %s]", formatDuration(elapsed.Nanoseconds()))
			}
		case StatusPending:
			statusIcon = "⚪"
			if task.TotalDuration > 0 {
				timeInfo = fmt.Sprintf(" [Total: %s]", task.GetFormattedDuration())
			}
		case StatusPaused:
			statusIcon = "🟡"
			timeInfo = fmt.Sprintf(" [Paused: %s]", task.GetFormattedDuration())
		case StatusDone:
			statusIcon = "✅"
			if task.TotalDuration > 0 {
				timeInfo = fmt.Sprintf(" [Total: %s]", task.GetFormattedDuration())
			}
			if task.CompletedAt != nil {
				timeInfo += fmt.Sprintf(" [Completed: %s]", task.CompletedAt.Format("15:04"))
			}
		}

		fmt.Printf("  %d. %s %s%s\n", i+1, statusIcon, task.Title, timeInfo)

		if len(task.Sessions) > 0 && (task.Status == StatusDone || task.Status == StatusPaused) {
			fmt.Printf("     Sessions: %d │ ", len(task.Sessions))
			if len(task.Sessions) <= 3 {
				for j, session := range task.Sessions {
					fmt.Printf("%s", formatDuration(session.Duration))
					if j < len(task.Sessions)-1 {
						fmt.Print(", ")
					}
				}
			} else {
				for j := 0; j < 2; j++ {
					fmt.Printf("%s, ", formatDuration(task.Sessions[j].Duration))
				}
				fmt.Printf("... +%d more", len(task.Sessions)-2)
			}
			fmt.Println()
		}
	}
}

func addTask(taskList *TaskList, title string) {
	newTask := Task{
		ID:            time.Now().UnixNano(),
		Title:         title,
		Status:        StatusPending,
		Comment:       "",
		Sessions:      []Session{},
		TotalDuration: 0,
		CreatedAt:     time.Now(),
	}

	taskList.Items = append(taskList.Items, newTask)
	fmt.Printf("✨ Added: %s\n", title)
}

func removeTask(taskList *TaskList, index int) error {
	if index < 1 || index > len(taskList.Items) {
		return fmt.Errorf("invalid task number. Use 1-%d", len(taskList.Items))
	}

	removedTask := taskList.Items[index-1]
	taskList.Items = append(taskList.Items[:index-1], taskList.Items[index:]...)

	fmt.Printf("🗑️ Removed: %s\n", removedTask.Title)
	return nil
}

func toggleTaskTimer(taskList *TaskList, index int) error {
	if index < 1 || index > len(taskList.Items) {
		return fmt.Errorf("invalid task number. Use 1-%d", len(taskList.Items))
	}

	task := &taskList.Items[index-1]

	if task.Status == StatusDone {
		return fmt.Errorf("cannot start timer for completed task")
	}

	now := time.Now()

	switch task.Status {
	case StatusPending, StatusPaused:
		for i := range taskList.Items {
			if taskList.Items[i].Status == StatusActive {
				stopTaskTimer(&taskList.Items[i], now)
			}
		}
		task.Status = StatusActive
		task.ActiveStartTime = &now
		fmt.Printf("▶️ Started: %s\n", task.Title)

	case StatusActive:
		stopTaskTimer(task, now)
		fmt.Printf("⏸️ Paused: %s [Session: %s] [Total: %s]\n", 
			task.Title, 
			formatDuration(task.Sessions[len(task.Sessions)-1].Duration),
			task.GetFormattedDuration())
	}

	return nil
}

func stopTaskTimer(task *Task, endTime time.Time) {
	if task.ActiveStartTime == nil {
		return
	}

	duration := endTime.Sub(*task.ActiveStartTime)
	session := Session{
		StartTime: *task.ActiveStartTime,
		EndTime:   endTime,
		Duration:  duration.Nanoseconds(),
	}

	task.Sessions = append(task.Sessions, session)
	task.TotalDuration += duration.Nanoseconds()
	task.Status = StatusPaused
	task.ActiveStartTime = nil
}

func markTaskComplete(taskList *TaskList, index int) error {
	if index < 1 || index > len(taskList.Items) {
		return fmt.Errorf("invalid task number. Use 1-%d", len(taskList.Items))
	}

	task := &taskList.Items[index-1]
	now := time.Now()

	if task.Status == StatusActive {
		stopTaskTimer(task, now)
	}

	task.Status = StatusDone
	task.CompletedAt = &now

	totalTime := ""
	if task.TotalDuration > 0 {
		totalTime = fmt.Sprintf(" [Total time: %s]", task.GetFormattedDuration())
	}

	fmt.Printf("✅ Completed: %s%s\n", task.Title, totalTime)
	return nil
}