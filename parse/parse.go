package parse

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/acearchive/artifact-action/logger"
	"github.com/icza/dyno"
	"gopkg.in/yaml.v2"
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

func findArtifactFiles(workspacePath, artifactsPath string) ([]string, error) {
	return filepath.Glob(filepath.Join(workspacePath, artifactsPath, fmt.Sprintf("*%s", ArtifactFileExtension)))
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

func parseGenericEntry(frontMatter string) (GenericEntry, error) {
	entry := GenericEntry{}
	if err := yaml.Unmarshal([]byte(frontMatter), &entry); err != nil {
		return GenericEntry{}, err
	}

	return entry.Sanitize(), nil
}

func (e ArtifactEntry) ToGeneric() GenericEntry {
	rawJSON, err := json.Marshal(e)
	if err != nil {
		logger.LogError(fmt.Errorf("%w\n%#v", err, e))
		logger.Exit()
	}

	entry := GenericEntry{}

	if err := json.Unmarshal(rawJSON, &entry); err != nil {
		logger.LogError(fmt.Errorf("%w\n%#v", err, e))
		logger.Exit()
	}

	return entry.Sanitize()
}

// Sanitize replaces `map[any]any` values that cannot be serialized to JSON
// with `map[string]any`.
func (e GenericEntry) Sanitize() GenericEntry {
	if generic, ok := dyno.ConvertMapI2MapS(map[string]any(e)).(map[string]any); ok {
		return generic
	}

	panic("failed type assertion, this is a bug")
}

func (e GenericEntry) ToTyped(value any) {
	rawJSON, err := json.Marshal(e)
	if err != nil {
		logger.LogError(fmt.Errorf("%w\n%#v", err, e))
		logger.Exit()
	}

	if err := json.Unmarshal(rawJSON, value); err != nil {
		logger.LogError(fmt.Errorf("%w\n%#v", err, e))
		logger.Exit()
	}
}

func (a Artifact) Version() int {
	entry := struct {
		Version int `json:"version"`
	}{}

	a.Entry.ToTyped(&entry)

	return entry.Version
}
