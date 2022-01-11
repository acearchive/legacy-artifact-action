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

func Marshal(entries []parse.ArtifactEntry) (string, error) {
	artifacts := make([]ArtifactOutput, len(entries))
	for entryIndex, artifactEntry := range entries {
		files := make([]ArtifactFileOutput, len(artifactEntry.Files))
		for fileIndex, fileEntry := range artifactEntry.Files {
			files[fileIndex] = ArtifactFileOutput{
				Name:      fileEntry.Name,
				MediaType: fileEntry.MediaType,
				Filename:  fileEntry.Filename,
				Cid:       fileEntry.Cid,
			}
		}

		artifacts[entryIndex] = ArtifactOutput{
			Title:           artifactEntry.Title,
			Description:     artifactEntry.Description,
			LongDescription: artifactEntry.LongDescription,
			Files:           files,
			People:          artifactEntry.People,
			Identities:      artifactEntry.Identities,
			FromYear:        artifactEntry.FromYear,
			ToYear:          artifactEntry.ToYear,
			Decades:         artifactEntry.Decades,
		}
	}

	output := Output{Artifacts: artifacts}

	marshalledOutput, err := json.Marshal(output)
	if err != nil {
		return "", err
	}

	return string(marshalledOutput), nil
}
