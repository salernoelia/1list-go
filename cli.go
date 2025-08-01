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
📝 Task CLI Manager

Usage:
  ./1list                     - List tasks
  ./1list done <number>       - Toggle task
  ./1list set-folder <path>   - Set folder
  ./1list create-list <name>  - Create new list
  ./1list help               - Show help

Examples:
  ./1list set-folder ~/Tasks
  ./1list create-list "Shopping"
  ./1list
  ./1list done 3
`)
}

func setDir(config *Config)  {
			if len(os.Args) < 3 {
			fmt.Println("❌ Need folder path")
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
			fmt.Printf("❌ Dir not found: %s\n", absDir)
			return
		}
		
		config.TaskDir = absDir
		err = saveConfig(config)
		if err != nil {
			fmt.Printf("❌ Save error: %v\n", err)
			return
		}
		
		fmt.Printf("✅ Dir set: %s\n", absDir)
		showDirContents(absDir)
}

func removeList(config *Config)  {
	if config.TaskDir == "" {
			fmt.Println("❌ No folder configured")
			fmt.Println("Use: ./1list set-folder <path>")
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
				err := os.Remove(filepath.Join(config.TaskDir, taskFiles[0]))
				if err != nil {
					fmt.Printf("❌ Failed to remove: %v\n", err)
					return
				}
				fmt.Printf("✅ Removed: %s\n", taskFiles[0])
			}
			return
		}
		
		fmt.Printf("📋 Found %d task lists:\n\n", len(taskFiles))
		for i, file := range taskFiles {
			displayName := strings.TrimSuffix(file, ".1list")
			if idx := strings.Index(displayName, "-"); idx != -1 {
				displayName = strings.TrimSpace(displayName[:idx])
			}
			fmt.Printf("%d. %s\n", i+1, displayName)
		}
		
		fmt.Print("\nSelect list to remove (1-", len(taskFiles), "): ")
		var choice int
		_, err = fmt.Scanf("%d", &choice)
		if err != nil || choice < 1 || choice > len(taskFiles) {
			fmt.Println("❌ Invalid selection")
			return
		}
		
		selectedFile := taskFiles[choice-1]
		fmt.Printf("Remove '%s'? (y/N): ", selectedFile)
		scanner := bufio.NewScanner(os.Stdin)
		if scanner.Scan() && strings.ToLower(scanner.Text()) == "y" {
			err := os.Remove(filepath.Join(config.TaskDir, selectedFile))
			if err != nil {
				fmt.Printf("❌ Failed to remove: %v\n", err)
				return
			}
			fmt.Printf("✅ Removed: %s\n", selectedFile)
		}
}

func createList(config *Config)  {
	if config.TaskDir == "" {
			fmt.Println("❌ No folder configured")
			fmt.Println("Use: ./1list set-folder <path>")
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
		
		err := createNewList(config.TaskDir, listName)
		if err != nil {
			fmt.Printf("❌ %v\n", err)
			return
		}
		
		fmt.Printf("✅ Created list: %s\n", listName)
		showDirContents(config.TaskDir)
}

func runCLI()  {
		config, err := loadConfig()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	
	if len(os.Args) < 2 {
		if config.TaskDir == "" {
			fmt.Println("🔧 No folder configured")
			fmt.Println("Use: ./1list set-folder <path>")
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
			fmt.Printf("❌ Error loading: %v\n", err)
			return
		}
		
		fmt.Printf("📁 %s\n", filepath.Base(taskFile))
		listTasks(taskList)
		
		fmt.Println("💡 Commands: <number>, 'add <task>', 'remove <number>', return: 'r', quit: 'q':")
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

			if input == "r" || input == "return"  {
				runCLI()
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
				fmt.Println("💡 Commands: <number>, 'add <task>', return: 'r', 'remove <number>', quit: 'q':")
				continue
			}
			
	
			if strings.HasPrefix(input, "remove ") {
				taskNumStr := strings.TrimSpace(input[7:])
				taskNum, err := strconv.Atoi(taskNumStr)
				if err != nil {
					fmt.Printf("❌ '%s' not a number\n", taskNumStr)
					continue
				}
				
				if taskNum < 1 || taskNum > len(taskList.Items) {
					fmt.Printf("❌ Invalid task number. Use 1-%d\n", len(taskList.Items))
					continue
				}
				
				removedTask := taskList.Items[taskNum-1]
				taskList.Items = append(taskList.Items[:taskNum-1], taskList.Items[taskNum:]...)
				
				err = saveTasks(taskFile, taskList)
				if err != nil {
					fmt.Printf("❌ Save error: %v\n", err)
					continue
				}
				
				fmt.Printf("🗑️ Removed: %s\n", removedTask.Title)
				fmt.Println()
				listTasks(taskList)
				fmt.Println("💡 Commands: <number>, 'add <task>', 'remove <number>', return: 'r', quit: 'q':")
				continue
			}
			
			taskNum, err := strconv.Atoi(input)
			if err != nil {
				fmt.Println("❌ Enter number, 'add <task>', 'remove <number>', or 'q'")
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
			fmt.Println("💡 Commands: <number>, 'add <task>', 'remove <number>', 'q':")
		}
		return
	}
	
	command := os.Args[1]
	
	switch command {
	case "set-folder":
		setDir(config)
		
	case "create-list":
		createList(config)
		
	case "remove-list":
		removeList(config)
		
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
		
		if config.TaskDir == "" {
			fmt.Println("❌ No folder configured")
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