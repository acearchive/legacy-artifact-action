package parse

import (
	"github.com/acearchive/artifact-action/logger"
	"github.com/ipfs/go-cid"
)

// Update this file if the schema of artifact files ever changes in such a way
// that we need to change how we extract the CIDs. These methods operate on a
// `GenericEntry` because they must always support extracting CIDs from all
// past schema versions.

func (a Artifact) listCids() []cid.Cid {
	// We use an anonymous type here because if the artifact schema changes and
	// `ArtifactEntry` changes with it, we still need to support the old schema.
	entry := struct {
		Files []struct {
			Cid string `json:"cid"`
		} `json:"files"`
	}{}

	a.Entry.ToTyped(&entry)

	cidList := make([]cid.Cid, len(entry.Files))

	for fileIndex, artifactFile := range entry.Files {
		artifactCid, err := cid.Parse(artifactFile.Cid)
		if err == nil {
			cidList[fileIndex] = artifactCid
		}
	}

	return cidList
}

type DeduplicatedCIDList []cid.Cid

func ExtractCids(artifacts []Artifact) (DeduplicatedCIDList, error) {
	cidList := make([]cid.Cid, 0, len(artifacts))
	contentSet := make(map[ContentKey]struct{}, len(artifacts))

	for _, artifact := range artifacts {
		for _, currentCid := range artifact.listCids() {
			if _, alreadyExists := contentSet[ContentKeyFromCid(currentCid)]; !alreadyExists {
				contentSet[ContentKeyFromCid(currentCid)] = struct{}{}

				cidList = append(cidList, currentCid)
			}
		}
	}

	logger.Printf("Found %d unique CIDs in artifact files\n", len(cidList))

	return cidList, nil
}
