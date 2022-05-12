package parse

import (
	"github.com/frawleyskid/w3s-upload/logger"
	"os"
	"path/filepath"
)

func logArtifactErrors(artifactErrors []error) {
	logger.LogError(ErrInvalidArtifactFiles)
	logger.LogErrorGroup("Artifact file errors:", artifactErrors)
	os.Exit(1)
}

func Tree(workspacePath, pathGlob string) ([]Artifact, error) {
	artifactFilePaths, err := findArtifactFiles(workspacePath, pathGlob)
	if err != nil {
		return nil, err
	}

	logger.Printf("Found %d artifact files in the tree\n", len(artifactFilePaths))

	var artifactErrors []error

	artifacts := make([]Artifact, 0, len(artifactFilePaths))

	for _, filePath := range artifactFilePaths {
		relativePath, err := filepath.Rel(workspacePath, filePath)
		if err != nil {
			return nil, err
		}

		registerErr := func(reason error) {
			artifactErrors = append(artifactErrors, ArtifactParseError{
				Path:   relativePath,
				Reason: reason.Error(),
			})
		}

		artifactFile, err := os.Open(filePath)
		if err != nil {
			return nil, err
		}

		frontMatter, err := extractFrontMatter(artifactFile)
		if err != nil {
			registerErr(err)
			continue
		}

		entry, err := parseArtifactEntry(frontMatter)
		if err != nil {
			registerErr(err)
			continue
		}

		if validateErr := ValidateEntry(entry, relativePath); validateErr != nil {
			artifactErrors = append(artifactErrors, validateErr)
		}

		artifacts = append(artifacts, Artifact{
			Path:  relativePath,
			Slug:  filepath.Base(filepath.Dir(relativePath)),
			Rev:   nil,
			Entry: entry,
		})
	}

	if len(artifactErrors) != 0 {
		logArtifactErrors(artifactErrors)
	} else {
		logger.Println("All artifact files in the tree are valid")
	}

	return artifacts, nil
}
