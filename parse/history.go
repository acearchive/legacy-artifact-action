package parse

import (
	"fmt"
	"github.com/frawleyskid/w3s-upload/logger"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"path/filepath"
	"strings"
	"time"
)

type Revision struct {
	File object.File
	Path string
	Rev  string
	Date time.Time
}

// findRevisions returns all the commits in the git history reachable from
// `HEAD` and returns them in order from most to least recent.
func findRevisions(workspacePath, artifactsPath string) ([]Revision, error) {
	artifactsGlob := filepath.Join(artifactsPath, fmt.Sprintf("*%s", ArtifactFileExtension))

	repo, err := git.PlainOpen(workspacePath)
	if err != nil {
		return nil, err
	}

	pathFilter := func(path string) bool {
		matches, _ := filepath.Match(artifactsGlob, path)
		return matches
	}

	commitIter, err := repo.Log(&git.LogOptions{PathFilter: pathFilter, Order: git.LogOrderCommitterTime})
	if err != nil {
		return nil, err
	}

	var revs []Revision

	commitFunc := func(commit *object.Commit) error {
		stats, err := commit.Stats()
		if err != nil {
			return err
		}

		for _, stat := range stats {
			if matches, _ := filepath.Match(artifactsGlob, stat.Name); matches {
				file, err := commit.File(stat.Name)
				if err != nil {
					return err
				}

				revs = append(revs, Revision{
					File: *file,
					Path: stat.Name,
					Rev:  commit.Hash.String(),
					Date: commit.Committer.When,
				})
			}
		}

		return nil
	}

	if err := commitIter.ForEach(commitFunc); err != nil {
		return nil, err
	}

	return revs, nil
}

func History(workspacePath, artifactsPath string) ([]Artifact, error) {
	artifactRevisions, err := findRevisions(workspacePath, artifactsPath)
	if err != nil {
		return nil, err
	}

	artifacts := make([]Artifact, 0, len(artifactRevisions))

	for revIndex, revision := range artifactRevisions {
		artifactFile, err := revision.File.Reader()
		if err != nil {
			return nil, err
		}

		frontMatter, err := extractFrontMatter(artifactFile)
		if err != nil {
			continue
		}

		entry, err := parseGenericEntry(frontMatter)
		if err != nil {
			continue
		}

		artifacts = append(artifacts, Artifact{
			Path: revision.File.Name,
			Slug: strings.TrimSuffix(filepath.Base(revision.File.Name), ArtifactFileExtension),
			Commit: &ArtifactCommit{
				Rev:  artifactRevisions[revIndex].Rev,
				Date: artifactRevisions[revIndex].Date.UTC(),
			},
			Entry: entry,
		})
	}

	logger.Printf("Found %d artifact files in the history\n", len(artifacts))

	return artifacts, nil
}
