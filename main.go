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
	w3sToken := os.Getenv("INPUT_W3S-TOKEN")
	pathGlob := os.Getenv("INPUT_PATH-GLOB")
	uploadContent := os.Getenv("INPUT_UPLOAD")
	ipfsApiAddr := os.Getenv("INPUT_IPFS-API")

	artifactEntries, err := parse.ArtifactEntries(workspacePath, pathGlob)
	if err != nil {
		return err
	}

	cidList, err := parse.ExtractCids(artifactEntries)
	if err != nil {
		return err
	}

	if uploadContent != "true" || ipfsApiAddr == "" || w3sToken == "" {
		return nil
	}

	return upload.Content(ctx, w3sToken, ipfsApiAddr, cidList)
}

func main() {
	if err := run(); err != nil {
		fmt.Printf("::error::%s\n", err.Error())
		os.Exit(1)
	}
}
