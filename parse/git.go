package parse

import (
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
)

func FindRevisions(repo *git.Repository, relativePath string) (files []object.File, revs []string, err error) {
	commits, err := repo.Log(&git.LogOptions{FileName: &relativePath})
	if err != nil {
		return nil, nil, err
	}

	commitIter := func(commit *object.Commit) error {
		file, err := commit.File(relativePath)
		if err != nil {
			return err
		}

		files = append(files, *file)
		revs = append(revs, commit.Hash.String())

		return nil
	}

	if err := commits.ForEach(commitIter); err != nil {
		return nil, nil, err
	}

	return files, revs, nil
}
