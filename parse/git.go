package parse

import (
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
)

func FindRevisions(repo *git.Repository, relativePath string) ([]object.File, error) {
	commits, err := repo.Log(&git.LogOptions{FileName: &relativePath})
	if err != nil {
		return nil, err
	}

	var files []object.File

	commitIter := func(commit *object.Commit) error {
		file, err := commit.File(relativePath)
		if err != nil {
			return err
		}

		files = append(files, *file)

		return nil
	}

	if err := commits.ForEach(commitIter); err != nil {
		return nil, err
	}

	return files, nil
}
