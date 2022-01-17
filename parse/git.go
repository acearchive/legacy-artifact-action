package parse

import (
	"github.com/frawleyskid/w3s-upload/logger"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"path/filepath"
)

type Revision struct {
	File object.File
	Rev  string
}

func findRevisions(workspacePath, pathGlob string) ([]Revision, error) {
	repo, err := git.PlainOpen(workspacePath)
	if err != nil {
		return nil, err
	}

	pathFilter := func(path string) bool {
		matches, _ := filepath.Match(pathGlob, path)
		return matches
	}

	commitIter, err := repo.Log(&git.LogOptions{PathFilter: pathFilter})
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
			if matches, _ := filepath.Match(pathGlob, stat.Name); matches {
				file, err := commit.File(stat.Name)
				if err != nil {
					return err
				}

				revs = append(revs, Revision{
					File: *file,
					Rev:  commit.Hash.String(),
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

func History(workspacePath, pathGlob string) ([]Artifact, error) {
	artifactRevisions, err := findRevisions(workspacePath, pathGlob)
	if err != nil {
		return nil, err
	}

	artifacts := make([]Artifact, 0, len(artifactRevisions))

	for i, revision := range artifactRevisions {
		artifactFile, err := revision.File.Reader()
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
			Slug:  filepath.Base(filepath.Dir(revision.File.Name)),
			Rev:   &artifactRevisions[i].Rev,
		})
	}

	logger.Printf("Found %d valid artifact files in history\n", len(artifacts))

	return artifacts, nil
}
