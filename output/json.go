package output

import (
	"encoding/json"
	"github.com/frawleyskid/w3s-upload/parse"
)

type ArtifactFileOutput struct {
	Name      string  `json:"name"`
	MediaType *string `json:"mediaType"`
	Filename  *string `json:"filename"`
	Cid       string  `json:"cid"`
}

type ArtifactOutput struct {
	Slug            string               `json:"slug"`
	Rev             string               `json:"rev"`
	Title           string               `json:"title"`
	Description     string               `json:"description"`
	LongDescription *string              `json:"longDescription"`
	Files           []ArtifactFileOutput `json:"files"`
	People          []string             `json:"people"`
	Identities      []string             `json:"identities"`
	FromYear        int                  `json:"fromYear"`
	ToYear          *int                 `json:"toYear"`
	Decades         []int                `json:"decades"`
}

type Output struct {
	Artifacts []ArtifactOutput `json:"artifacts"`
}

func Marshal(entries []parse.Artifact) (string, error) {
	artifacts := make([]ArtifactOutput, len(entries))
	for entryIndex, artifact := range entries {
		files := make([]ArtifactFileOutput, len(artifact.Entry.Files))
		for fileIndex, fileEntry := range artifact.Entry.Files {
			files[fileIndex] = ArtifactFileOutput{
				Name:      fileEntry.Name,
				MediaType: fileEntry.MediaType,
				Filename:  fileEntry.Filename,
				Cid:       fileEntry.Cid,
			}
		}

		artifacts[entryIndex] = ArtifactOutput{
			Slug:            artifact.Slug,
			Rev:             artifact.Rev,
			Title:           artifact.Entry.Title,
			Description:     artifact.Entry.Description,
			LongDescription: artifact.Entry.LongDescription,
			Files:           files,
			People:          artifact.Entry.People,
			Identities:      artifact.Entry.Identities,
			FromYear:        artifact.Entry.FromYear,
			ToYear:          artifact.Entry.ToYear,
			Decades:         artifact.Entry.Decades,
		}
	}

	output := Output{Artifacts: artifacts}

	marshalledOutput, err := json.Marshal(output)
	if err != nil {
		return "", err
	}

	return string(marshalledOutput), nil
}
