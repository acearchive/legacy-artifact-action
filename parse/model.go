package parse

import "github.com/ipfs/go-cid"

const CurrentArtifactVersion = 1

type Artifact struct {
	Slug  string
	Rev   *string
	Entry ArtifactEntry
}

type ArtifactFileEntry struct {
	Name      string  `yaml:"name"`
	MediaType *string `yaml:"mediaType"`
	Filename  *string `yaml:"filename"`
	Cid       string  `yaml:"cid"`
}

type ArtifactLinkEntry struct {
	Name string `yaml:"name"`
	Url  string `yaml:"url"`
}

type ArtifactEntry struct {
	Version         int                 `yaml:"version"`
	Title           string              `yaml:"title"`
	Description     string              `yaml:"description"`
	LongDescription *string             `yaml:"longDescription"`
	Files           []ArtifactFileEntry `yaml:"files"`
	Links           []ArtifactLinkEntry `yaml:"links"`
	People          []string            `yaml:"people"`
	Identities      []string            `yaml:"identities"`
	FromYear        int                 `yaml:"fromYear"`
	ToYear          *int                `yaml:"toYear"`
	Decades         []int               `yaml:"decades"`
}

func (e Artifact) AllCids() ([]cid.Cid, error) {
	cidList := make([]cid.Cid, len(e.Entry.Files))

	for fileIndex, artifactFile := range e.Entry.Files {
		artifactCid, err := cid.Parse(artifactFile.Cid)
		if err != nil {
			return nil, err
		}

		cidList[fileIndex] = artifactCid
	}

	return cidList, nil
}
