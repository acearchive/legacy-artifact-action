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

func parseArtifactEntry(frontMatter string) (ArtifactEntry, error) {
	entry := ArtifactEntry{}
	if err := yaml.Unmarshal([]byte(frontMatter), &entry); err != nil {
		return ArtifactEntry{}, err
	}

	return entry, nil
}

func ArtifactEntries(workspacePath, pathGlob string) ([]ArtifactEntry, error) {
	artifactFilePaths, err := findArtifactFiles(workspacePath, pathGlob)
	if err != nil {
		return nil, err
	}

	fmt.Printf("Found %d artifact files\n", len(artifactFilePaths))

	entries := make([]ArtifactEntry, len(artifactFilePaths))

	for entryIndex, filePath := range artifactFilePaths {
		frontMatter, err := extractFrontMatter(filePath)
		if err != nil {
			return nil, err
		}

		entry, err := parseArtifactEntry(frontMatter)
		if err != nil {
			return nil, err
		}

		entries[entryIndex] = entry
	}

	return entries, nil
}

func ExtractCids(entries []ArtifactEntry) ([]cid.Cid, error) {
	cidList := make([]cid.Cid, 0, len(entries))

	for _, entry := range entries {
		currentCidList, err := entry.AllCids()
		if err != nil {
			return nil, err
		}

		cidList = append(cidList, currentCidList...)
	}

	fmt.Printf("Found %d CIDs in artifact files\n", len(cidList))

	return cidList, nil
}
