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

func runCLI()  {
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
		
	case "create-list":
		if config.TaskFolder == "" {
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
		
		err := createNewList(config.TaskFolder, listName)
		if err != nil {
			fmt.Printf("❌ %v\n", err)
			return
		}
		
		fmt.Printf("✅ Created list: %s\n", listName)
		showFolderContents(config.TaskFolder)
		
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