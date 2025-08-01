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

const configFile = ".task-cli-config.json"

func getConfigPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, configFile)
}

func loadConfig() (*Config, error) {
	configPath := getConfigPath()
	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			config := &Config{}
			saveConfig(config)
			return config, nil
		}
		return nil, err
	}
	
	var config Config
	err = json.Unmarshal(data, &config)
	return &config, err
}

func saveConfig(config *Config) error {
	configPath := getConfigPath()
	data, err := json.Marshal(config)
	if err != nil {
		return err
	}
	return os.WriteFile(configPath, data, 0644)
}

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
		fmt.Printf("%d. %s\n", i+1, displayName)
	}
	
	fmt.Print("\nSelect a list (1-", len(taskFiles), "): ")
	var choice int
	_, err := fmt.Scanf("%d", &choice)
	if err != nil || choice < 1 || choice > len(taskFiles) {
		return "", fmt.Errorf("invalid selection")
	}
	
	return filepath.Join(folder, taskFiles[choice-1]), nil
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

func listTasks(taskList *TaskList) {
	fmt.Printf("\n📋 %s\n", taskList.Title)
	fmt.Println(strings.Repeat("=", len(taskList.Title)+4))
	
	for i, task := range taskList.Items {
		status := "❌"
		if task.Done {
			status = "✅"
		}
		fmt.Printf("%d. %s %s\n", i+1, status, task.Title)
	}
	fmt.Println()
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

func showFolderContents(folder string) {
	entries, err := os.ReadDir(folder)
	if err != nil {
		return
	}
	
	fmt.Printf("\nFiles in %s:\n", folder)
	var taskFiles []string
	var otherFiles []string
	
	for _, entry := range entries {
		if !entry.IsDir() {
			if strings.HasSuffix(entry.Name(), ".1list") {
				taskFiles = append(taskFiles, entry.Name())
			} else {
				otherFiles = append(otherFiles, entry.Name())
			}
		}
	}
	
	if len(taskFiles) > 0 {
		fmt.Println("  📋 Task lists:")
		for _, file := range taskFiles {
			fmt.Printf("    - %s\n", file)
		}
	}
	
	if len(otherFiles) > 0 {
		fmt.Println("  📄 Other files:")
		for _, file := range otherFiles {
			fmt.Printf("    - %s\n", file)
		}
	}
	
	if len(taskFiles) == 0 && len(otherFiles) == 0 {
		fmt.Println("  (no files found)")
	}
	fmt.Println()
}

func printUsage() {
	fmt.Print(`
📝 Task CLI Manager

Usage:
  ./1list                     - List tasks
  ./1list done <number>       - Toggle task
  ./1list set-folder <path>   - Set folder
  ./1list help               - Show help

Examples:
  ./1list set-folder ~/Tasks
  ./1list
  ./1list done 3
`)
}

func main() {
	config, err := loadConfig()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	
	if len(os.Args) < 2 {
		if config.TaskFolder == "" {
			fmt.Println("🔧 No folder configured")
			fmt.Println("Use: ./1list set-folder <path>")
			return
		}
		
		taskFiles, err := findTaskFiles(config.TaskFolder)
		if err != nil {
			fmt.Printf("❌ %v\n", err)
			showFolderContents(config.TaskFolder)
			return
		}
		
		taskFile, err := selectTaskFile(config.TaskFolder, taskFiles)
		if err != nil {
			fmt.Printf("❌ %v\n", err)
			return
		}
		
		taskList, err := loadTasks(taskFile)
		if err != nil {
			fmt.Printf("❌ Error loading: %v\n", err)
			return
		}
		
		fmt.Printf("📁 %s\n", filepath.Base(taskFile))
		listTasks(taskList)
		
		fmt.Println("💡 Commands: <number>, 'add <task>', 'q':")
		scanner := bufio.NewScanner(os.Stdin)
		for {
			fmt.Print("> ")
			if !scanner.Scan() {
				break
			}
			input := strings.TrimSpace(scanner.Text())
			
			if input == "q" || input == "quit" || input == "exit" {
				break
			}
			
			if strings.HasPrefix(input, "add ") {
				taskTitle := strings.TrimSpace(input[4:])
				if taskTitle == "" {
					fmt.Println("❌ Need task title")
					continue
				}
				addTask(taskList, taskTitle)
				
				err = saveTasks(taskFile, taskList)
				if err != nil {
					fmt.Printf("❌ Save error: %v\n", err)
					continue
				}
				
				fmt.Println()
				listTasks(taskList)
				fmt.Println("💡 Commands: <number>, 'add <task>', 'q':")
				continue
			}
			
			taskNum, err := strconv.Atoi(input)
			if err != nil {
				fmt.Println("❌ Enter number, 'add <task>', or 'q'")
				continue
			}
			
			err = markTaskDone(taskList, taskNum)
			if err != nil {
				fmt.Printf("❌ %v\n", err)
				continue
			}
			
			err = saveTasks(taskFile, taskList)
			if err != nil {
				fmt.Printf("❌ Save error: %v\n", err)
				continue
			}
			
			fmt.Println()
			listTasks(taskList)
			fmt.Println("💡 Commands: <number>, 'add <task>', 'q':")
		}
		return
	}
	
	command := os.Args[1]
	
	switch command {
	case "set-folder":
		if len(os.Args) < 3 {
			fmt.Println("❌ Need folder path")
			return
		}
		
		folder := os.Args[2]
		if strings.HasPrefix(folder, "~/") {
			home, _ := os.UserHomeDir()
			folder = filepath.Join(home, folder[2:])
		}
		
		absFolder, err := filepath.Abs(folder)
		if err != nil {
			fmt.Printf("❌ Invalid path: %v\n", err)
			return
		}
		
		if _, err := os.Stat(absFolder); os.IsNotExist(err) {
			fmt.Printf("❌ Folder not found: %s\n", absFolder)
			return
		}
		
		config.TaskFolder = absFolder
		err = saveConfig(config)
		if err != nil {
			fmt.Printf("❌ Save error: %v\n", err)
			return
		}
		
		fmt.Printf("✅ Folder set: %s\n", absFolder)
		showFolderContents(absFolder)
		
	case "done":
		if len(os.Args) < 3 {
			fmt.Println("❌ Need task number")
			return
		}
		
		taskNum, err := strconv.Atoi(os.Args[2])
		if err != nil {
			fmt.Printf("❌ '%s' not a number\n", os.Args[2])
			return
		}
		
		if config.TaskFolder == "" {
			fmt.Println("❌ No folder configured")
			return
		}
		
		taskFiles, err := findTaskFiles(config.TaskFolder)
		if err != nil {
			fmt.Printf("❌ %v\n", err)
			return
		}
		
		taskFile, err := selectTaskFile(config.TaskFolder, taskFiles)
		if err != nil {
			fmt.Printf("❌ %v\n", err)
			return
		}
		
		taskList, err := loadTasks(taskFile)
		if err != nil {
			fmt.Printf("❌ Load error: %v\n", err)
			return
		}
		
		err = markTaskDone(taskList, taskNum)
		if err != nil {
			fmt.Printf("❌ %v\n", err)
			return
		}
		
		err = saveTasks(taskFile, taskList)
		if err != nil {
			fmt.Printf("❌ Save error: %v\n", err)
			return
		}
		
	case "help", "-h", "--help":
		printUsage()
		
	default:
		fmt.Printf("❌ Unknown: %s\n", command)
		printUsage()
	}
}