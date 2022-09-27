package dir

import (
	"context"

	"github.com/frawleyskid/w3s-upload/client"
	"github.com/frawleyskid/w3s-upload/parse"
	"github.com/ipfs/go-cid"
	dag "github.com/ipfs/go-merkledag"
	unixfs "github.com/ipfs/go-unixfs/io"
)

func DefaultCidPrefix() cid.Prefix {
	return dag.V1CidPrefix()
}

func Build(ctx context.Context, artifacts []parse.Artifact) (cid.Cid, error) {
	ipfsClientGuard, err := client.New()
	if err != nil {
		return cid.Undef, err
	}

	ipfsClient := ipfsClientGuard.Lock()
	defer ipfsClientGuard.Unlock()

	rootDir := unixfs.NewDirectory(ipfsClient.Dag())
	rootDir.SetCidBuilder(DefaultCidPrefix())

	for _, artifact := range artifacts {
		artifactDir := unixfs.NewDirectory(ipfsClient.Dag())
		artifactDir.SetCidBuilder(DefaultCidPrefix())

		entry := struct {
			Files []struct {
				Filename string `json:"filename"`
				Cid      string `json:"cid"`
			} `json:"files"`
		}{}

		artifact.Entry.ToTyped(&entry)

		for _, file := range entry.Files {
			fileCid, err := cid.Parse(file.Cid)
			if err != nil {
				return cid.Undef, err
			}

			fileNode, err := ipfsClient.Dag().Get(ctx, fileCid)
			if err != nil {
				return cid.Undef, err
			}

			if err := artifactDir.AddChild(ctx, file.Filename, fileNode); err != nil {
				return cid.Undef, err
			}
		}

		artifactNode, err := artifactDir.GetNode()
		if err != nil {
			return cid.Undef, err
		}

		if err := rootDir.AddChild(ctx, artifact.Slug, artifactNode); err != nil {
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
