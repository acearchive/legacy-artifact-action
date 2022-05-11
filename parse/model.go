package parse

const CurrentArtifactVersion = 1

type Artifact struct {
	Path  string        `json:"path"`
	Slug  string        `json:"slug"`
	Rev   *string       `json:"rev"`
	Entry ArtifactEntry `json:"entry"`
}

type ArtifactEntryFile struct {
	Name      string  `yaml:"name" json:"name"`
	MediaType *string `yaml:"mediaType" json:"mediaType"`
	Filename  *string `yaml:"filename" json:"filename"`
	Cid       string  `yaml:"cid" json:"cid"`
}

type ArtifactEntryLink struct {
	Name string `yaml:"name" json:"name"`
	Url  string `yaml:"url" json:"url"`
}

type ArtifactEntry struct {
	Version         int                 `yaml:"version" json:"version"`
	Title           string              `yaml:"title" json:"title"`
	Description     string              `yaml:"description" json:"description"`
	LongDescription *string             `yaml:"longDescription" json:"longDescription"`
	Files           []ArtifactEntryFile `yaml:"files" json:"files"`
	Links           []ArtifactEntryLink `yaml:"links" json:"links"`
	People          []string            `yaml:"people" json:"people"`
	Identities      []string            `yaml:"identities" json:"identities"`
	FromYear        int                 `yaml:"fromYear" json:"fromYear"`
	ToYear          *int                `yaml:"toYear" json:"toYear"`
	Decades         []int               `yaml:"decades" json:"decades"`
	Aliases         []string            `yaml:"aliases" json:"aliases"`
}
