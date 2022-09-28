package dir

import (
	"context"
	"sort"
	"strings"

	"github.com/frawleyskid/w3s-upload/client"
	"github.com/frawleyskid/w3s-upload/parse"
	"github.com/ipfs/go-cid"
	dag "github.com/ipfs/go-merkledag"
	unixfs "github.com/ipfs/go-unixfs/io"
)

func DefaultCidPrefix() cid.Prefix {
	return dag.V1CidPrefix()
}

// fileMapType is a map of file names to their CIDs.
type fileMapType = map[string]cid.Cid

// artifactMapType is a map of artifact slugs to maps of their files.
type artifactMapType = map[string]fileMapType

func getLatestFiles(artifacts []parse.Artifact) artifactMapType {
	artifactMap := make(artifactMapType, len(artifacts))

	sort.Slice(artifacts, func(i, j int) bool {
		return artifacts[j].Commit.Date.Before(artifacts[i].Commit.Date)
	})

	for _, artifact := range artifacts {
		entry := struct {
			Files []struct {
				Filename string `json:"filename"`
				Cid      string `json:"cid"`
			} `json:"files"`
		}{}

		artifact.Entry.ToTyped(&entry)

		if len(entry.Files) == 0 {
			continue
		}

		var (
			fileMap fileMapType
			exists  bool
		)

		if fileMap, exists = artifactMap[artifact.Slug]; !exists {
			fileMap = make(fileMapType)
			artifactMap[artifact.Slug] = fileMap
		}

		for _, file := range entry.Files {
			if len(strings.TrimSpace(file.Filename)) == 0 {
				continue
			}

			parsedCid, err := cid.Parse(file.Cid)
			if err != nil {
				continue
			}

			if _, exists := fileMap[file.Filename]; !exists {
				fileMap[file.Filename] = parsedCid
			}
		}
	}

	return artifactMap
}

func Build(ctx context.Context, artifacts []parse.Artifact) (cid.Cid, error) {
	ipfsClientGuard, err := client.New()
	if err != nil {
		return cid.Undef, err
	}

	ipfsClient := ipfsClientGuard.Lock()
	defer ipfsClientGuard.Unlock()

	artifactMap := getLatestFiles(artifacts)

	// Unsure why this is failing since it doesn't take a context.
	//nolint:contextcheck
	rootDir := unixfs.NewDirectory(ipfsClient.Dag())
	rootDir.SetCidBuilder(DefaultCidPrefix())

	for artifactSlug, artifactFiles := range artifactMap {
		// Unsure why this is failing since it doesn't take a context.
		//nolint:contextcheck
		artifactDir := unixfs.NewDirectory(ipfsClient.Dag())
		artifactDir.SetCidBuilder(DefaultCidPrefix())

		for filename, fileCid := range artifactFiles {
			fileNode, err := ipfsClient.Dag().Get(ctx, fileCid)
			if err != nil {
				return cid.Undef, err
			}

			if err := artifactDir.AddChild(ctx, filename, fileNode); err != nil {
				return cid.Undef, err
			}
		}

		artifactNode, err := artifactDir.GetNode()
		if err != nil {
			return cid.Undef, err
		}

		if err := rootDir.AddChild(ctx, artifactSlug, artifactNode); err != nil {
			return cid.Undef, err
		}

		if err := ipfsClient.Dag().Add(ctx, artifactNode); err != nil {
			return cid.Undef, err
		}
	}

	rootNode, err := rootDir.GetNode()
	if err != nil {
		return cid.Undef, err
	}

	if err := ipfsClient.Dag().Add(ctx, rootNode); err != nil {
		return cid.Undef, err
	}

	return rootNode.Cid(), nil
}
