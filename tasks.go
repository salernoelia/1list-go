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
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".1list") {
			taskFiles = append(taskFiles, entry.Name())
		}
	}
	
	if len(taskFiles) == 0 {
		return nil, fmt.Errorf("no .1list files found")
	}
	
	return taskFiles, nil
}

func selectTaskFile(folder string, taskFiles []string) (string, error) {
	if len(taskFiles) == 1 {
		return filepath.Join(folder, taskFiles[0]), nil
	}
	
	fmt.Printf("📋 Found %d task lists:\n\n", len(taskFiles))
	for i, file := range taskFiles {
		displayName := strings.TrimSuffix(file, ".1list")
		if idx := strings.Index(displayName, "-"); idx != -1 {
			displayName = strings.TrimSpace(displayName[:idx])
		}
		fmt.Printf("%d. %s\n", i+1, displayName)
	}
	
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("\nSelect a list (1-", len(taskFiles), "), create one 'create <name>' or remove 'remove <number>': ")
		if !scanner.Scan() {
			return "", fmt.Errorf("input error")
		}
		input := strings.TrimSpace(scanner.Text())

		if strings.HasPrefix(input, "create ") {
			listName := strings.TrimSpace(input[7:])
			if listName == "" {
				fmt.Println("❌ Need list name")
				continue
			}
			err := createNewList(folder, listName)
			if err != nil {
				fmt.Printf("❌ %v\n", err)
				continue
			}
			fmt.Printf("✅ Created list: %s\n", listName)
			// Refresh taskFiles after creation
			taskFiles, err = findTaskFiles(folder)
			if err != nil {
				fmt.Printf("❌ %v\n", err)
				return "", err
			}
			// Redisplay lists
			fmt.Printf("\n📋 Found %d task lists:\n\n", len(taskFiles))
			for i, file := range taskFiles {
				displayName := strings.TrimSuffix(file, ".1list")
				if idx := strings.Index(displayName, "-"); idx != -1 {
					displayName = strings.TrimSpace(displayName[:idx])
				}
				fmt.Printf("%d. %s\n", i+1, displayName)
			}
			continue
		}

		if strings.HasPrefix(input, "remove ") {
			numStr := strings.TrimSpace(input[7:])
			choice, err := strconv.Atoi(numStr)
			if err != nil || choice < 1 || choice > len(taskFiles) {
				fmt.Println("❌ Invalid selection")
				continue
			}
			selectedFile := taskFiles[choice-1]
			fmt.Printf("Remove '%s'? (y/N): ", selectedFile)
			if scanner.Scan() && strings.ToLower(scanner.Text()) == "y" {
				err := os.Remove(filepath.Join(folder, selectedFile))
				if err != nil {
					fmt.Printf("❌ Failed to remove: %v\n", err)
					continue
				}
				fmt.Printf("✅ Removed: %s\n", selectedFile)
				// Refresh taskFiles after removal
				taskFiles, err = findTaskFiles(folder)
				if err != nil {
					fmt.Printf("❌ %v\n", err)
					return "", err
				}
				if len(taskFiles) == 0 {
					return "", fmt.Errorf("no .1list files found")
				}
				// Redisplay lists
				fmt.Printf("\n📋 Found %d task lists:\n\n", len(taskFiles))
				for i, file := range taskFiles {
					displayName := strings.TrimSuffix(file, ".1list")
					if idx := strings.Index(displayName, "-"); idx != -1 {
						displayName = strings.TrimSpace(displayName[:idx])
					}
					fmt.Printf("%d. %s\n", i+1, displayName)
				}
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
	
	sanitizedName := strings.ReplaceAll(listName, " ", "-")
	sanitizedName = strings.ReplaceAll(sanitizedName, "/", "-")
	sanitizedName = strings.ReplaceAll(sanitizedName, "\\", "-")
	
	files, _ := os.ReadDir(folder)
	counter := 1
	for _, file := range files {
		if strings.HasPrefix(file.Name(), sanitizedName+"-") && strings.HasSuffix(file.Name(), ".1list") {
			counter++
		}
	}
	
	fileName := fmt.Sprintf("%s-%d.1list", sanitizedName, counter)
	filePath := filepath.Join(folder, fileName)
	
	newTaskList := &TaskList{
		Title: listName,
		Items: []Task{},
	}
	
	return saveTasks(filePath, newTaskList)
}

func listTasks(taskList *TaskList) {
    fmt.Printf("\n📋 %s\n", taskList.Title)
    fmt.Println(strings.Repeat("=", len(taskList.Title)+4))

    fmt.Println("Not done:")
    anyNotDone := false
    for i, task := range taskList.Items {
        if !task.Done {
            fmt.Printf("%d. ❌ %s\n", i+1, task.Title)
            anyNotDone = true
        }
    }
    if anyNotDone {
        fmt.Println()
    }

    fmt.Println("Done:")
    anyDone := false
    for i, task := range taskList.Items {
        if task.Done {
            fmt.Printf("%d. ✅ %s\n", i+1, task.Title)
            anyDone = true
        }
    }
    if anyDone {
        fmt.Println()
    }
}

func addTask(taskList *TaskList, title string) {
	newTask := Task{
		ID:               time.Now().UnixNano(),
		Title:            title,
		Done:             false,
		Comment:          "",
		CommentDisplayed: false,
	}
	
	taskList.Items = append(taskList.Items, newTask)
	fmt.Printf("✨ Added: %s\n", title)
}

func markTaskDone(taskList *TaskList, index int) error {
	if index < 1 || index > len(taskList.Items) {
		return fmt.Errorf("invalid task number. Use 1-%d", len(taskList.Items))
	}
	
	task := &taskList.Items[index-1]
	task.Done = !task.Done
	
	action := "completed"
	if !task.Done {
		action = "incomplete"
	}
	
	fmt.Printf("✨ %s: %s\n", action, task.Title)
	return nil
}