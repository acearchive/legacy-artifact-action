package main

import (
	"context"
	"fmt"
	"github.com/frawleyskid/w3s-upload/parse"
	"github.com/frawleyskid/w3s-upload/upload"
	"os"
)

func run() error {
	ctx := context.Background()

	workspacePath := os.Getenv("GITHUB_WORKSPACE")
	w3sToken := os.Getenv("INPUT_W3S_TOKEN")
	pathGlob := os.Getenv("INPUT_PATH_GLOB")
	uploadContent := os.Getenv("INPUT_UPLOAD")

	cidList, err := parse.ParseArtifactFiles(workspacePath, pathGlob)
	if err != nil {
		return err
	}

	if uploadContent != "true" {
		return nil
	}

	return upload.UploadContent(ctx, w3sToken, cidList)
}

func main() {
	err := run()
	if err != nil {
		fmt.Printf("::error::%s", err.Error())
		os.Exit(1)
	}
}
