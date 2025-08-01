package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestConfig(t *testing.T) {
	tempDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	t.Run("LoadConfig_NewConfig", func(t *testing.T) {
		config, err := loadConfig()
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if config.TaskDir != "" {
			t.Fatalf("Expected empty task folder, got %s", config.TaskDir)
		}
	})

	t.Run("SaveAndLoadConfig", func(t *testing.T) {
		config := &Config{TaskDir: "/test/path"}
		err := saveConfig(config)
		if err != nil {
			t.Fatalf("Expected no error saving config, got %v", err)
		}

		loadedConfig, err := loadConfig()
		if err != nil {
			t.Fatalf("Expected no error loading config, got %v", err)
		}
		if loadedConfig.TaskDir != "/test/path" {
			t.Fatalf("Expected '/test/path', got %s", loadedConfig.TaskDir)
		}
	})
}

func TestTaskOperations(t *testing.T) {
	t.Run("AddTask", func(t *testing.T) {
		taskList := &TaskList{
			Title: "Test List",
			Items: []Task{},
		}

		addTask(taskList, "Test Task")

		if len(taskList.Items) != 1 {
			t.Fatalf("Expected 1 task, got %d", len(taskList.Items))
		}
		if taskList.Items[0].Title != "Test Task" {
			t.Fatalf("Expected 'Test Task', got %s", taskList.Items[0].Title)
		}
		if taskList.Items[0].Done {
			t.Fatalf("Expected task to be not done")
		}
		if taskList.Items[0].ID == 0 {
			t.Fatalf("Expected task to have an ID")
		}
	})

	t.Run("MarkTaskDone", func(t *testing.T) {
		taskList := &TaskList{
			Title: "Test List",
			Items: []Task{
				{ID: 1, Title: "Task 1", Done: false},
				{ID: 2, Title: "Task 2", Done: false},
			},
		}

		err := markTaskDone(taskList, 1)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if !taskList.Items[0].Done {
			t.Fatalf("Expected task 1 to be done")
		}

		err = markTaskDone(taskList, 1)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if taskList.Items[0].Done {
			t.Fatalf("Expected task 1 to be not done after toggle")
		}
	})

	t.Run("MarkTaskDone_InvalidIndex", func(t *testing.T) {
		taskList := &TaskList{
			Title: "Test List",
			Items: []Task{
				{ID: 1, Title: "Task 1", Done: false},
			},
		}

		err := markTaskDone(taskList, 0)
		if err == nil {
			t.Fatalf("Expected error for index 0")
		}

		err = markTaskDone(taskList, 2)
		if err == nil {
			t.Fatalf("Expected error for index 2")
		}

		err = markTaskDone(taskList, -1)
		if err == nil {
			t.Fatalf("Expected error for negative index")
		}
	})
}

func TestFileOperations(t *testing.T) {
	tempDir := t.TempDir()

	t.Run("SaveAndLoadTasks", func(t *testing.T) {
		taskList := &TaskList{
			Title: "Test List",
			Items: []Task{
				{ID: 1, Title: "Task 1", Done: false, Comment: "Test comment"},
				{ID: 2, Title: "Task 2", Done: true},
			},
		}

		filePath := filepath.Join(tempDir, "test.1list")
		err := saveTasks(filePath, taskList)
		if err != nil {
			t.Fatalf("Expected no error saving tasks, got %v", err)
		}

		loadedList, err := loadTasks(filePath)
		if err != nil {
			t.Fatalf("Expected no error loading tasks, got %v", err)
		}

		if loadedList.Title != "Test List" {
			t.Fatalf("Expected title 'Test List', got %s", loadedList.Title)
		}
		if len(loadedList.Items) != 2 {
			t.Fatalf("Expected 2 tasks, got %d", len(loadedList.Items))
		}
		if loadedList.Items[0].Title != "Task 1" {
			t.Fatalf("Expected 'Task 1', got %s", loadedList.Items[0].Title)
		}
		if loadedList.Items[1].Done != true {
			t.Fatalf("Expected task 2 to be done")
		}
	})

	t.Run("LoadTasks_FileNotExists", func(t *testing.T) {
		_, err := loadTasks(filepath.Join(tempDir, "nonexistent.1list"))
		if err == nil {
			t.Fatalf("Expected error loading nonexistent file")
		}
	})

	t.Run("LoadTasks_InvalidJSON", func(t *testing.T) {
		invalidFile := filepath.Join(tempDir, "invalid.1list")
		err := os.WriteFile(invalidFile, []byte("invalid json"), 0644)
		if err != nil {
			t.Fatalf("Failed to create invalid file: %v", err)
		}

		_, err = loadTasks(invalidFile)
		if err == nil {
			t.Fatalf("Expected error loading invalid JSON")
		}
	})
}

func TestCreateNewList(t *testing.T) {
	tempDir := t.TempDir()

	t.Run("CreateNewList_Success", func(t *testing.T) {
		err := createNewList(tempDir, "Shopping List")
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		files, err := os.ReadDir(tempDir)
		if err != nil {
			t.Fatalf("Failed to read temp dir: %v", err)
		}

		found := false
		var createdFile string
		for _, file := range files {
			if file.Name() == "Shopping-List-1.1list" {
				found = true
				createdFile = file.Name()
				break
			}
		}

		if !found {
			t.Fatalf("Expected to find Shopping-List-1.1list file")
		}

		taskList, err := loadTasks(filepath.Join(tempDir, createdFile))
		if err != nil {
			t.Fatalf("Failed to load created list: %v", err)
		}

		if taskList.Title != "Shopping List" {
			t.Fatalf("Expected title 'Shopping List', got %s", taskList.Title)
		}
		if len(taskList.Items) != 0 {
			t.Fatalf("Expected empty task list, got %d items", len(taskList.Items))
		}
	})

	t.Run("CreateNewList_EmptyName", func(t *testing.T) {
		err := createNewList(tempDir, "")
		if err == nil {
			t.Fatalf("Expected error for empty name")
		}

		err = createNewList(tempDir, "   ")
		if err == nil {
			t.Fatalf("Expected error for whitespace-only name")
		}
	})

	t.Run("CreateNewList_SpecialCharacters", func(t *testing.T) {
		err := createNewList(tempDir, "Work/Personal Tasks")
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		files, err := os.ReadDir(tempDir)
		if err != nil {
			t.Fatalf("Failed to read temp dir: %v", err)
		}

		found := false
		for _, file := range files {
			if file.Name() == "Work-Personal-Tasks-1.1list" {
				found = true
				break
			}
		}

		if !found {
			t.Fatalf("Expected to find Work-Personal-Tasks-1.1list file")
		}
	})

	t.Run("CreateNewList_WindowsPath", func(t *testing.T) {
		err := createNewList(tempDir, "Documents\\Important")
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		files, err := os.ReadDir(tempDir)
		if err != nil {
			t.Fatalf("Failed to read temp dir: %v", err)
		}

		found := false
		for _, file := range files {
			if file.Name() == "Documents-Important-1.1list" {
				found = true
				break
			}
		}

		if !found {
			t.Fatalf("Expected to find Documents-Important-1.1list file")
		}
	})
}

func TestFindTaskFiles(t *testing.T) {
	tempDir := t.TempDir()

	t.Run("FindTaskFiles_NoFiles", func(t *testing.T) {
		_, err := findTaskFiles(tempDir)
		if err == nil {
			t.Fatalf("Expected error when no .1list files found")
		}
	})

	t.Run("FindTaskFiles_WithFiles", func(t *testing.T) {
		err := os.WriteFile(filepath.Join(tempDir, "list1.1list"), []byte("{}"), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
		err = os.WriteFile(filepath.Join(tempDir, "list2.1list"), []byte("{}"), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
		err = os.WriteFile(filepath.Join(tempDir, "other.txt"), []byte("text"), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		files, err := findTaskFiles(tempDir)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if len(files) != 2 {
			t.Fatalf("Expected 2 files, got %d", len(files))
		}

		expectedFiles := map[string]bool{
			"list1.1list": false,
			"list2.1list": false,
		}

		for _, file := range files {
			if _, exists := expectedFiles[file]; exists {
				expectedFiles[file] = true
			} else {
				t.Fatalf("Unexpected file: %s", file)
			}
		}

		for file, found := range expectedFiles {
			if !found {
				t.Fatalf("Expected file not found: %s", file)
			}
		}
	})

	t.Run("FindTaskFiles_EmptyDir", func(t *testing.T) {
		_, err := findTaskFiles("")
		if err == nil {
			t.Fatalf("Expected error for empty folder path")
		}
	})

	t.Run("FindTaskFiles_NonexistentDir", func(t *testing.T) {
		_, err := findTaskFiles("/nonexistent/path")
		if err == nil {
			t.Fatalf("Expected error for nonexistent folder")
		}
	})
}

func TestTaskModel(t *testing.T) {
	t.Run("TaskSerialization", func(t *testing.T) {
		task := Task{
			ID:               12345,
			Title:            "Test Task",
			Done:             true,
			Comment:          "Test comment",
			CommentDisplayed: true,
		}

		data, err := json.Marshal(task)
		if err != nil {
			t.Fatalf("Failed to marshal task: %v", err)
		}

		var unmarshaled Task
		err = json.Unmarshal(data, &unmarshaled)
		if err != nil {
			t.Fatalf("Failed to unmarshal task: %v", err)
		}

		if task.ID != unmarshaled.ID {
			t.Fatalf("ID mismatch: expected %d, got %d", task.ID, unmarshaled.ID)
		}
		if task.Title != unmarshaled.Title {
			t.Fatalf("Title mismatch: expected %s, got %s", task.Title, unmarshaled.Title)
		}
		if task.Done != unmarshaled.Done {
			t.Fatalf("Done mismatch: expected %t, got %t", task.Done, unmarshaled.Done)
		}
		if task.Comment != unmarshaled.Comment {
			t.Fatalf("Comment mismatch: expected %s, got %s", task.Comment, unmarshaled.Comment)
		}
		if task.CommentDisplayed != unmarshaled.CommentDisplayed {
			t.Fatalf("CommentDisplayed mismatch: expected %t, got %t", task.CommentDisplayed, unmarshaled.CommentDisplayed)
		}
	})
}

func TestEdgeCases(t *testing.T) {
	tempDir := t.TempDir()

	t.Run("VeryLongTaskTitle", func(t *testing.T) {
		longTitle := strings.Repeat("a", 1000)
		taskList := &TaskList{Title: "Test", Items: []Task{}}
		
		addTask(taskList, longTitle)
		
		if len(taskList.Items) != 1 {
			t.Fatalf("Expected 1 task")
		}
		if taskList.Items[0].Title != longTitle {
			t.Fatalf("Title was truncated or modified")
		}

		filePath := filepath.Join(tempDir, "long.1list")
		err := saveTasks(filePath, taskList)
		if err != nil {
			t.Fatalf("Failed to save tasks with long title: %v", err)
		}

		loadedList, err := loadTasks(filePath)
		if err != nil {
			t.Fatalf("Failed to load tasks with long title: %v", err)
		}
		if loadedList.Items[0].Title != longTitle {
			t.Fatalf("Long title was not preserved")
		}
	})



	t.Run("ConcurrentTaskIDs", func(t *testing.T) {
		t.Skip("Skipping concurrent test - CLI is single-threaded")
	})

	t.Run("FilePermissions", func(t *testing.T) {
		readOnlyDir := filepath.Join(tempDir, "readonly")
		err := os.Mkdir(readOnlyDir, 0555)
		if err != nil {
			t.Fatalf("Failed to create readonly dir: %v", err)
		}
		defer os.Chmod(readOnlyDir, 0755)

		err = createNewList(readOnlyDir, "Test List")
		if err == nil {
			t.Fatalf("Expected error creating list in readonly directory")
		}
	})
}

func BenchmarkAddTask(b *testing.B) {
	taskList := &TaskList{Title: "Benchmark", Items: []Task{}}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		addTask(taskList, "Benchmark Task")
	}
}

func BenchmarkMarkTaskDone(b *testing.B) {
	taskList := &TaskList{Title: "Benchmark", Items: make([]Task, 1000)}
	for i := range taskList.Items {
		taskList.Items[i] = Task{
			ID:    int64(i),
			Title: "Task",
			Done:  false,
		}
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		markTaskDone(taskList, (i%1000)+1)
	}
}

func BenchmarkSaveTasks(b *testing.B) {
	tempDir := b.TempDir()
	taskList := &TaskList{
		Title: "Benchmark",
		Items: make([]Task, 100),
	}
	
	for i := range taskList.Items {
		taskList.Items[i] = Task{
			ID:    int64(i),
			Title: "Benchmark Task",
			Done:  i%2 == 0,
		}
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		filePath := filepath.Join(tempDir, "bench.1list")
		saveTasks(filePath, taskList)
	}
}