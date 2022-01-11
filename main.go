package main

import (
	"context"
	"fmt"
	"github.com/frawleyskid/w3s-upload/output"
	"github.com/frawleyskid/w3s-upload/parse"
	"github.com/frawleyskid/w3s-upload/upload"
	"os"
)

func run() error {
	ctx := context.Background()

	workspacePath := os.Getenv("GITHUB_WORKSPACE")
	pathGlob := os.Getenv("INPUT_PATH-GLOB")
	w3sToken := os.Getenv("INPUT_W3S-TOKEN")
	ipfsApiAddr := os.Getenv("INPUT_IPFS-API")

	artifacts, err := parse.ArtifactEntries(workspacePath, pathGlob)
	if err != nil {
		return err
	}

	jsonOutput, err := output.Marshal(artifacts)
	if err != nil {
		return err
	}

	fmt.Printf("::set-output name=artifacts::%s\n", jsonOutput)

	if ipfsApiAddr == "" || w3sToken == "" {
		return nil
	}

	cidList, err := parse.ExtractCids(artifacts)
	if err != nil {
		return err
	}

	return upload.Content(ctx, w3sToken, ipfsApiAddr, cidList)
}

func main() {
	if err := run(); err != nil {
		fmt.Printf("::error::%s\n", err.Error())
		os.Exit(1)
	}
}
