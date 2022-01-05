package main

import (
	"fmt"
	"os"
)

func main() {
	workspaceDir := os.Getenv("GITHUB_WORKSPACE")
	fmt.Printf("::notice::%s\n", workspaceDir)
}
