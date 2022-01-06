package parse

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/ipfs/go-cid"
	"gopkg.in/yaml.v2"
	"os"
	"path/filepath"
	"strings"
)

const frontMatterDelimiter = "---"

var ErrNoFrontMatter = errors.New("this file has no front matter")

func findArtifactFiles(workspacePath, pathGlob string) ([]string, error) {
	return filepath.Glob(filepath.Join(workspacePath, pathGlob))
}

func isFrontMatterDelimiter(line string) bool {
	return strings.HasPrefix(line, frontMatterDelimiter) && strings.TrimSpace(line) == frontMatterDelimiter
}

func isWhitespace(line string) bool {
	return strings.TrimSpace(line) == ""
}

func extractFrontMatter(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}

	defer func() {
		closeErr := file.Close()
		if closeErr != nil && err == nil {
			err = closeErr
		}
	}()

	scanner := bufio.NewScanner(file)

	// Find the start of the front matter block.
findStart:
	for {
		scanner.Scan()
		if scanner.Err() != nil {
			return "", err
		}

		currentLine := scanner.Text()

		switch {
		case isWhitespace(currentLine):
			continue
		case isFrontMatterDelimiter(currentLine):
			break findStart
		default:
			return "", ErrNoFrontMatter
		}
	}

	var frontMatter strings.Builder

	for {
		scanner.Scan()

		if scanner.Err() != nil {
			return "", err
		}

		currentLine := scanner.Text()

		if isFrontMatterDelimiter(currentLine) {
			break
		}

		frontMatter.WriteString(currentLine + "\n")
	}

	return frontMatter.String(), nil
}

type artifactFileEntry struct {
	Cid string `yaml:"cid"`
}

type artifactEntry struct {
	Files []artifactFileEntry `yaml:"files"`
}

func extractArtifactCids(frontMatter string) ([]cid.Cid, error) {
	artifact := artifactEntry{}
	if err := yaml.Unmarshal([]byte(frontMatter), &artifact); err != nil {
		return nil, err
	}

	cidList := make([]cid.Cid, len(artifact.Files))

	for fileIndex, artifactFile := range artifact.Files {
		artifactCid, err := cid.Parse(artifactFile.Cid)
		if err != nil {
			return nil, err
		}

		cidList[fileIndex] = artifactCid
	}

	return cidList, nil
}

func ArtifactFiles(workspacePath, pathGlob string) ([]cid.Cid, error) {
	artifactFilePaths, err := findArtifactFiles(workspacePath, pathGlob)
	if err != nil {
		return nil, err
	}

	fmt.Printf("Found %d artifact files\n", len(artifactFilePaths))

	cidList := make([]cid.Cid, 0, len(artifactFilePaths))
	for _, filePath := range artifactFilePaths {
		frontMatter, err := extractFrontMatter(filePath)
		if err != nil {
			return nil, err
		}

		currentCidList, err := extractArtifactCids(frontMatter)
		if err != nil {
			return nil, err
		}

		cidList = append(cidList, currentCidList...)
	}

	fmt.Printf("Found %d CIDs in artifact files\n", len(cidList))

	return cidList, nil
}
