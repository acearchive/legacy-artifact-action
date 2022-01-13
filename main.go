package main

import (
	"context"
	"fmt"
	"github.com/frawleyskid/w3s-upload/output"
	"github.com/frawleyskid/w3s-upload/parse"
	"github.com/frawleyskid/w3s-upload/pin"
	"github.com/frawleyskid/w3s-upload/w3s"
	"os"
)

func run() error {
	ctx := context.Background()

	workspacePath := os.Getenv("GITHUB_WORKSPACE")
	pathGlob := os.Getenv("INPUT_PATH-GLOB")
	w3sToken := os.Getenv("INPUT_W3S-TOKEN")
	ipfsApiAddr := os.Getenv("INPUT_IPFS-API")
	pinEndpoint := os.Getenv("INPUT_PIN-ENDPOINT")
	pinToken := os.Getenv("INPUT_PIN-TOKEN")

	if err := parse.Validate(workspacePath, pathGlob); err != nil {
		return err
	}

	artifacts, err := parse.Artifacts(workspacePath, pathGlob)
	if err != nil {
		return err
	}

	jsonOutput, err := output.Marshal(artifacts)
	if err != nil {
		return err
	}

	fmt.Printf("::set-output name=artifacts::%s\n", jsonOutput)

	cidList, err := parse.ExtractCids(artifacts)
	if err != nil {
		return err
	}

	if ipfsApiAddr != "" && w3sToken != "" {
		if err := w3s.Upload(ctx, w3sToken, ipfsApiAddr, cidList); err != nil {
			return err
		}
	}

	if pinEndpoint != "" && pinToken != "" {
		if err := pin.Pin(ctx, pinEndpoint, pinToken, cidList); err != nil {
			return err
		}
	}

	return nil
}

func main() {
	if err := run(); err != nil {
		fmt.Printf("::error::%s\n", err.Error())
		os.Exit(1)
	}
}
