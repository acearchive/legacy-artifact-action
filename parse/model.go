package parse

const CurrentArtifactVersion = 1

type Artifact struct {
	Path  string
	Slug  string
	Rev   *string
	Entry ArtifactEntry
}

type ArtifactEntryFile struct {
	Name      string  `yaml:"name"`
	MediaType *string `yaml:"mediaType"`
	Filename  *string `yaml:"filename"`
	Cid       string  `yaml:"cid"`
}

type ArtifactEntryLink struct {
	Name string `yaml:"name"`
	Url  string `yaml:"url"`
}

type ArtifactEntry struct {
	Version         int                 `yaml:"version"`
	Title           string              `yaml:"title"`
	Description     string              `yaml:"description"`
	LongDescription *string             `yaml:"longDescription"`
	Files           []ArtifactEntryFile `yaml:"files"`
	Links           []ArtifactEntryLink `yaml:"links"`
	People          []string            `yaml:"people"`
	Identities      []string            `yaml:"identities"`
	FromYear        int                 `yaml:"fromYear"`
	ToYear          *int                `yaml:"toYear"`
	Decades         []int               `yaml:"decades"`
}
