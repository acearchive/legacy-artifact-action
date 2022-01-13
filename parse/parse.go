package parse

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/go-git/go-git/v5"
	"github.com/ipfs/go-cid"
	"gopkg.in/yaml.v2"
	"io"
	"path/filepath"
	"strings"
)

const frontMatterDelimiter = "---"

var ErrNoFrontMatter = errors.New("this file has no front matter")

type ArtifactParseError struct {
	Path   string
	Reason string
}

func (e ArtifactParseError) Error() string {
	return fmt.Sprintf("'%s': %s", e.Path, e.Reason)
}

func findArtifactFiles(workspacePath, pathGlob string) ([]string, error) {
	return filepath.Glob(filepath.Join(workspacePath, pathGlob))
}

func isFrontMatterDelimiter(line string) bool {
	return strings.HasPrefix(line, frontMatterDelimiter) && strings.TrimSpace(line) == frontMatterDelimiter
}

func isWhitespace(line string) bool {
	return strings.TrimSpace(line) == ""
}

func extractFrontMatter(file io.ReadCloser) (string, error) {
	var err error

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
	if err := yaml.UnmarshalStrict([]byte(frontMatter), &entry); err != nil {
		return ArtifactEntry{}, err
	}

	return entry, nil
}

func Artifacts(workspacePath, pathGlob string) ([]Artifact, error) {
	artifactFilePaths, err := findArtifactFiles(workspacePath, pathGlob)
	if err != nil {
		return nil, err
	}

	repo, err := git.PlainOpen(workspacePath)
	if err != nil {
		return nil, err
	}

	artifacts := make([]Artifact, 0, len(artifactFilePaths))

	for _, filePath := range artifactFilePaths {
		relativePath, err := filepath.Rel(workspacePath, filePath)
		if err != nil {
			return nil, err
		}

		objects, revs, err := FindRevisions(repo, relativePath)
		if err != nil {
			return nil, err
		}

		for revIndex, object := range objects {
			artifactFile, err := object.Reader()
			if err != nil {
				return nil, err
			}

			frontMatter, err := extractFrontMatter(artifactFile)
			if err != nil {
				continue
			}

			entry, err := parseArtifactEntry(frontMatter)
			if err != nil {
				continue
			}

			artifacts = append(artifacts, Artifact{
				Entry: entry,
				Slug:  filepath.Base(filepath.Dir(relativePath)),
				Rev:   revs[revIndex],
			})
		}
	}

	fmt.Printf("Found %d valid artifact files in history\n", len(artifacts))

	return artifacts, nil
}

func ExtractCids(artifacts []Artifact) ([]cid.Cid, error) {
	cidList := make([]cid.Cid, 0, len(artifacts))
	contentSet := make(map[ContentKey]struct{}, len(artifacts))

	for _, artifact := range artifacts {
		currentCidList, err := artifact.AllCids()
		if err != nil {
			return nil, err
		}

		for _, currentCid := range currentCidList {
			if _, alreadyExists := contentSet[ContentKeyFromCid(currentCid)]; !alreadyExists {
				contentSet[ContentKeyFromCid(currentCid)] = struct{}{}
				cidList = append(cidList, currentCid)
			}
		}
	}

	fmt.Printf("Found %d unique CIDs in artifact files\n", len(cidList))

	return cidList, nil
}
