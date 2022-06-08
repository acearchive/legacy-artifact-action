package parse

import (
	"time"
)

const CurrentArtifactVersion = 2

const ArtifactFileExtension = ".md"

type GenericEntry map[string]interface{}

type Artifact struct {
	Path   string          `json:"path"`
	Slug   string          `json:"slug"`
	Commit *ArtifactCommit `json:"commit"`
	Entry  GenericEntry    `json:"entry"`
}

type ArtifactCommit struct {
	Rev  string    `json:"rev"`
	Date time.Time `json:"date"`
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

type EntryField string

const (
	FieldVersion         EntryField = "version"
	FieldTitle           EntryField = "title"
	FieldDescription     EntryField = "description"
	FieldLongDescription EntryField = "longDescription"
	FieldFiles           EntryField = "files"
	FieldFileName        EntryField = "name"
	FieldFileMediaType   EntryField = "mediaType"
	FieldFileFilename    EntryField = "filename"
	FieldFileCid         EntryField = "cid"
	FieldLinks           EntryField = "links"
	FieldLinkName        EntryField = "name"
	FieldLinkUrl         EntryField = "url"
	FieldPeople          EntryField = "people"
	FieldIdentities      EntryField = "identities"
	FieldFromYear        EntryField = "fromYear"
	FieldToYear          EntryField = "toYear"
	FieldDecades         EntryField = "decades"
	FieldAliases         EntryField = "aliases"
)
