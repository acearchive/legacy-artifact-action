package parse

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/acearchive/artifact-action/logger"
)

func logArtifactErrors(artifactErrors []error) {
	logger.LogError(ErrInvalidArtifactFiles)
	logger.LogErrorGroup("Artifact file errors:", artifactErrors)
	logger.Exit()
}

func Tree(workspacePath, artifactsPath string) ([]Artifact, error) {
	artifactFilePaths, err := findArtifactFiles(workspacePath, artifactsPath)
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
			Path:   relativePath,
			Slug:   strings.TrimSuffix(filepath.Base(relativePath), ArtifactFileExtension),
			Commit: nil,
			Entry:  entry.ToGeneric(),
		})
	}

	if len(artifactErrors) != 0 {
		logArtifactErrors(artifactErrors)
	} else {
		logger.Println("All artifact files in the tree are valid")
	}

	return artifacts, nil
}
