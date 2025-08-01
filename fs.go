package main

import (
	"fmt"
	"os"
	"strings"
)

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