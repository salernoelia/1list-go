package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func printUsage() {
	fmt.Print(`
⏰ Enhanced Task CLI Manager

Usage:
  tgo                      - Interactive task management
  tgo start <number>       - Start/stop task timer
  tgo done <number>        - Mark task complete
  tgo set-folder <path>    - Configure task directory
  tgo create-list <name>   - Create new task list
  tgo remove-list          - Remove task list
  tgo help                 - Show this help

Interactive Commands:
  <number>        - Start/stop task timer
  add <task>      - Add new task
  remove <number> - Remove task
  done <number>   - Mark task complete
  r | return      - Return to main menu
  q | quit        - Exit program

Examples:
  tgo set-folder ~/Tasks
  tgo create-list "Sprint Planning"
  tgo
  tgo start 3
`)
}

func runCLI() {
	config, err := loadConfig()
	if err != nil {
		fmt.Printf("Configuration error: %v\n", err)
		os.Exit(1)
	}

	if len(os.Args) < 2 {
		runInteractiveMode(config)
		return
	}

	command := os.Args[1]
	switch command {
	case "set-folder":
		handleSetFolder(config)
	case "create-list":
		handleCreateList(config)
	case "remove-list":
		handleRemoveList(config)
	case "start":
		handleStartTask(config)
	case "done":
		handleMarkDone(config)
	case "help", "-h", "--help":
		printUsage()
	default:
		fmt.Printf("❌ Unknown command: %s\n", command)
		printUsage()
	}
}

func runInteractiveMode(config *Config) {
	if config.TaskDir == "" {
		fmt.Println("🔧 No task directory configured")
		fmt.Println("Use: tgo set-folder <path>")
		return
	}

	taskFiles, err := findTaskFiles(config.TaskDir)
	if err != nil {
		fmt.Printf("❌ %v\n", err)
		showDirContents(config.TaskDir)
		return
	}

	taskFile, err := selectTaskFile(config.TaskDir, taskFiles)
	if err != nil {
		fmt.Printf("❌ %v\n", err)
		return
	}

	taskList, err := loadTasks(taskFile)
	if err != nil {
		fmt.Printf("❌ Error loading tasks: %v\n", err)
		return
	}

	clearScreen()
	displayTaskList(taskList, filepath.Base(taskFile))

	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("\n> ")
		if !scanner.Scan() {
			break
		}

		input := strings.TrimSpace(scanner.Text())
		if input == "" {
			continue
		}

		if handleInteractiveCommand(input, taskList, taskFile) {
			break
		}

		clearScreen()
		displayTaskList(taskList, filepath.Base(taskFile))
	}
}

func handleInteractiveCommand(input string, taskList *TaskList, taskFile string) bool {
	switch {
	case input == "q" || input == "quit" || input == "exit":
		return true
	case input == "r" || input == "return":
		runCLI()
		return true
	case strings.HasPrefix(input, "add "):
		handleAddTask(input, taskList, taskFile)
	case strings.HasPrefix(input, "remove "):
		handleRemoveTask(input, taskList, taskFile)
	case strings.HasPrefix(input, "done "):
		handleDoneTask(input, taskList, taskFile)
	default:
		if taskNum, err := strconv.Atoi(input); err == nil {
			handleToggleTimer(taskNum, taskList, taskFile)
		} else {
			fmt.Println("❌ Invalid command. Type a number, 'add <task>', 'remove <number>', 'done <number>', 'r' to return, or 'q' to quit")
		}
	}
	return false
}

func handleAddTask(input string, taskList *TaskList, taskFile string) {
	taskTitle := strings.TrimSpace(input[4:])
	if taskTitle == "" {
		fmt.Println("❌ Task title cannot be empty")
		return
	}

	addTask(taskList, taskTitle)
	if err := saveTasks(taskFile, taskList); err != nil {
		fmt.Printf("❌ Save error: %v\n", err)
	}
}

func handleRemoveTask(input string, taskList *TaskList, taskFile string) {
	taskNumStr := strings.TrimSpace(input[7:])
	taskNum, err := strconv.Atoi(taskNumStr)
	if err != nil {
		fmt.Printf("❌ '%s' is not a valid number\n", taskNumStr)
		return
	}

	if err := removeTask(taskList, taskNum); err != nil {
		fmt.Printf("❌ %v\n", err)
		return
	}

	if err := saveTasks(taskFile, taskList); err != nil {
		fmt.Printf("❌ Save error: %v\n", err)
	}
}

func handleDoneTask(input string, taskList *TaskList, taskFile string) {
	taskNumStr := strings.TrimSpace(input[5:])
	taskNum, err := strconv.Atoi(taskNumStr)
	if err != nil {
		fmt.Printf("❌ '%s' is not a valid number\n", taskNumStr)
		return
	}

	if err := markTaskComplete(taskList, taskNum); err != nil {
		fmt.Printf("❌ %v\n", err)
		return
	}

	if err := saveTasks(taskFile, taskList); err != nil {
		fmt.Printf("❌ Save error: %v\n", err)
	}
}

func handleToggleTimer(taskNum int, taskList *TaskList, taskFile string) {
	if err := toggleTaskTimer(taskList, taskNum); err != nil {
		fmt.Printf("❌ %v\n", err)
		return
	}

	if err := saveTasks(taskFile, taskList); err != nil {
		fmt.Printf("❌ Save error: %v\n", err)
	}
}

func handleSetFolder(config *Config) {
	if len(os.Args) < 3 {
		fmt.Println("❌ Folder path required")
		return
	}

	folder := os.Args[2]
	if strings.HasPrefix(folder, "~/") {
		home, _ := os.UserHomeDir()
		folder = filepath.Join(home, folder[2:])
	}

	absDir, err := filepath.Abs(folder)
	if err != nil {
		fmt.Printf("❌ Invalid path: %v\n", err)
		return
	}

	if _, err := os.Stat(absDir); os.IsNotExist(err) {
		fmt.Printf("❌ Directory not found: %s\n", absDir)
		return
	}

	config.TaskDir = absDir
	if err := saveConfig(config); err != nil {
		fmt.Printf("❌ Save error: %v\n", err)
		return
	}

	fmt.Printf("✅ Task directory set: %s\n", absDir)
	showDirContents(absDir)
}

func handleCreateList(config *Config) {
	if config.TaskDir == "" {
		fmt.Println("❌ No task directory configured")
		fmt.Println("Use: tgo set-folder <path>")
		return
	}

	var listName string
	if len(os.Args) < 3 {
		fmt.Print("Enter list name: ")
		scanner := bufio.NewScanner(os.Stdin)
		if scanner.Scan() {
			listName = strings.TrimSpace(scanner.Text())
		}
		if listName == "" {
			fmt.Println("❌ List name cannot be empty")
			return
		}
	} else {
		listName = strings.Join(os.Args[2:], " ")
	}

	if err := createNewList(config.TaskDir, listName); err != nil {
		fmt.Printf("❌ %v\n", err)
		return
	}

	fmt.Printf("✅ Created list: %s\n", listName)
	showDirContents(config.TaskDir)
}

func handleRemoveList(config *Config) {
	if config.TaskDir == "" {
		fmt.Println("❌ No task directory configured")
		fmt.Println("Use: tgo set-folder <path>")
		return
	}

	taskFiles, err := findTaskFiles(config.TaskDir)
	if err != nil {
		fmt.Printf("❌ %v\n", err)
		showDirContents(config.TaskDir)
		return
	}

	if len(taskFiles) == 1 {
		fmt.Printf("Remove '%s'? (y/N): ", taskFiles[0])
		scanner := bufio.NewScanner(os.Stdin)
		if scanner.Scan() && strings.ToLower(scanner.Text()) == "y" {
			if err := os.Remove(filepath.Join(config.TaskDir, taskFiles[0])); err != nil {
				fmt.Printf("❌ Failed to remove: %v\n", err)
				return
			}
			fmt.Printf("✅ Removed: %s\n", taskFiles[0])
		}
		return
	}

	fmt.Printf("📋 Found %d task lists:\n\n", len(taskFiles))
	for i, file := range taskFiles {
		displayName := strings.TrimSuffix(file, ".json")
		fmt.Printf("%d. %s\n", i+1, displayName)
	}

	fmt.Printf("\nSelect list to remove (1-%d): ", len(taskFiles))
	var choice int
	if _, err := fmt.Scanf("%d", &choice); err != nil || choice < 1 || choice > len(taskFiles) {
		fmt.Println("❌ Invalid selection")
		return
	}

	selectedFile := taskFiles[choice-1]
	fmt.Printf("Remove '%s'? (y/N): ", selectedFile)
	scanner := bufio.NewScanner(os.Stdin)
	if scanner.Scan() && strings.ToLower(scanner.Text()) == "y" {
		if err := os.Remove(filepath.Join(config.TaskDir, selectedFile)); err != nil {
			fmt.Printf("❌ Failed to remove: %v\n", err)
			return
		}
		fmt.Printf("✅ Removed: %s\n", selectedFile)
	}
}

func handleStartTask(config *Config) {
	if len(os.Args) < 3 {
		fmt.Println("❌ Task number required")
		return
	}

	taskNum, err := strconv.Atoi(os.Args[2])
	if err != nil {
		fmt.Printf("❌ '%s' is not a valid number\n", os.Args[2])
		return
	}

	if config.TaskDir == "" {
		fmt.Println("❌ No task directory configured")
		return
	}

	taskFiles, err := findTaskFiles(config.TaskDir)
	if err != nil {
		fmt.Printf("❌ %v\n", err)
		return
	}

	taskFile, err := selectTaskFile(config.TaskDir, taskFiles)
	if err != nil {
		fmt.Printf("❌ %v\n", err)
		return
	}

	taskList, err := loadTasks(taskFile)
	if err != nil {
		fmt.Printf("❌ Load error: %v\n", err)
		return
	}

	if err := toggleTaskTimer(taskList, taskNum); err != nil {
		fmt.Printf("❌ %v\n", err)
		return
	}

	if err := saveTasks(taskFile, taskList); err != nil {
		fmt.Printf("❌ Save error: %v\n", err)
	}
}

func handleMarkDone(config *Config) {
	if len(os.Args) < 3 {
		fmt.Println("❌ Task number required")
		return
	}

	taskNum, err := strconv.Atoi(os.Args[2])
	if err != nil {
		fmt.Printf("❌ '%s' is not a valid number\n", os.Args[2])
		return
	}

	if config.TaskDir == "" {
		fmt.Println("❌ No task directory configured")
		return
	}

	taskFiles, err := findTaskFiles(config.TaskDir)
	if err != nil {
		fmt.Printf("❌ %v\n", err)
		return
	}

	taskFile, err := selectTaskFile(config.TaskDir, taskFiles)
	if err != nil {
		fmt.Printf("❌ %v\n", err)
		return
	}

	taskList, err := loadTasks(taskFile)
	if err != nil {
		fmt.Printf("❌ Load error: %v\n", err)
		return
	}

	if err := markTaskComplete(taskList, taskNum); err != nil {
		fmt.Printf("❌ %v\n", err)
		return
	}

	if err := saveTasks(taskFile, taskList); err != nil {
		fmt.Printf("❌ Save error: %v\n", err)
	}
}

func clearScreen() {
	fmt.Print("\033[2J\033[H")
}