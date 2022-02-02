package output

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/frawleyskid/w3s-upload/parse"
	"github.com/ipfs/go-cid"
	"github.com/spf13/viper"
)

const prettyJsonIndent = "  "

var ErrInvalidOutput = errors.New("invalid output parameter")

type ArtifactFileOutput struct {
	Name      string  `json:"name"`
	MediaType *string `json:"mediaType"`
	Filename  *string `json:"filename"`
	Cid       string  `json:"cid"`
}

type ArtifactLinkOutput struct {
	Name string `json:"name"`
	Url  string `json:"url"`
}

type ArtifactOutput struct {
	Slug            string               `json:"slug"`
	Rev             *string              `json:"rev"`
	Version         int                  `json:"version"`
	Title           string               `json:"title"`
	Description     string               `json:"description"`
	LongDescription *string              `json:"longDescription"`
	Files           []ArtifactFileOutput `json:"files"`
	Links           []ArtifactLinkOutput `json:"links"`
	People          []string             `json:"people"`
	Identities      []string             `json:"identities"`
	FromYear        int                  `json:"fromYear"`
	ToYear          *int                 `json:"toYear"`
	Decades         []int                `json:"decades"`
}

type Output struct {
	Artifacts []ArtifactOutput `json:"artifacts"`
}

func marshalArtifact(entries []parse.Artifact, pretty bool) (string, error) {
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

		links := make([]ArtifactLinkOutput, len(artifact.Entry.Links))
		for linkIndex, linkEntry := range artifact.Entry.Links {
			links[linkIndex] = ArtifactLinkOutput{
				Name: linkEntry.Name,
				Url:  linkEntry.Url,
			}
		}

		artifacts[entryIndex] = ArtifactOutput{
			Slug:            artifact.Slug,
			Rev:             artifact.Rev,
			Version:         artifact.Entry.Version,
			Title:           artifact.Entry.Title,
			Description:     artifact.Entry.Description,
			LongDescription: artifact.Entry.LongDescription,
			Files:           files,
			Links:           links,
			People:          artifact.Entry.People,
			Identities:      artifact.Entry.Identities,
			FromYear:        artifact.Entry.FromYear,
			ToYear:          artifact.Entry.ToYear,
			Decades:         artifact.Entry.Decades,
		}
	}

	output := Output{Artifacts: artifacts}

	var (
		marshalledOutput []byte
		err              error
	)

	if pretty {
		marshalledOutput, err = json.MarshalIndent(output, "", prettyJsonIndent)
	} else {
		marshalledOutput, err = json.Marshal(output)
	}

	if err != nil {
		return "", err
	}

	return string(marshalledOutput), nil
}

func marshalCid(cids []cid.Cid, pretty bool) (string, error) {
	var (
		marshalledOutput []byte
		err              error
	)

	marshalledCids := make([]string, len(cids))

	for i, id := range cids {
		marshalledCids[i] = id.String()
	}

	if pretty {
		marshalledOutput, err = json.MarshalIndent(marshalledCids, "", prettyJsonIndent)
	} else {
		marshalledOutput, err = json.Marshal(marshalledCids)
	}

	if err != nil {
		return "", err
	}

	return string(marshalledOutput), nil
}

func Print(entries []parse.Artifact, cidList []cid.Cid) error {
	if viper.GetBool("action") {
		artifactOutput, err := marshalArtifact(entries, false)
		if err != nil {
			return err
		}

		fmt.Printf("::set-output name=artifacts::%s\n", artifactOutput)

		cidOutput, err := marshalCid(cidList, false)
		if err != nil {
			return err
		}

		fmt.Printf("::set-output name=cids::%s\n", cidOutput)

		return nil
	}

	switch outputMode := viper.GetString("output"); outputMode {
	case "artifacts":
		artifactOutput, err := marshalArtifact(entries, true)
		if err != nil {
			return err
		}

		fmt.Println(artifactOutput)
	case "cids":
		cidOutput, err := marshalCid(cidList, true)
		if err != nil {
			return err
		}

		fmt.Println(cidOutput)
	case "summary":
	default:
		return fmt.Errorf("%w: %s", ErrInvalidOutput, outputMode)
	}

	return nil
}
