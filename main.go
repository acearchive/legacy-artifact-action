package main

import (
	"fmt"
	"os"
)

func main() {
	workspaceDir := os.Getenv("GITHUB_WORKSPACE")

	entries, err := os.ReadDir(workspaceDir)
	if err != nil {
		panic(err)
	}

	for _, entry := range entries {
		fmt.Println(entry.Name())
	}
}
