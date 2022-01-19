package parse

import (
	"errors"
	"fmt"
	"github.com/ipfs/go-cid"
	"mime"
	"path/filepath"
	"strings"
)

const indentString = "    "

var ErrInvalidArtifactFiles = errors.New("one or more artifact files are invalid")

type InvalidArtifactReason struct {
	FieldPath string
	Reason    string
}

type InvalidArtifactError struct {
	FilePath string
	Reasons  []InvalidArtifactReason
}

func (e InvalidArtifactError) Error() string {
	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("'%s':\n", e.FilePath))
	for _, reason := range e.Reasons {
		builder.WriteString(fmt.Sprintf("%s`%s` %s\n", indentString, reason.FieldPath, reason.Reason))
	}
	return builder.String()
}

func validateEntry(entry ArtifactEntry, filePath string) error {
	var reasons []InvalidArtifactReason

	registerError := func(fieldPath, reason string) {
		reasons = append(reasons, InvalidArtifactReason{
			FieldPath: fieldPath,
			Reason:    reason,
		})
	}

	if entry.Version != CurrentArtifactVersion {
		registerError("version", fmt.Sprintf("must be the current version (%d)", CurrentArtifactVersion))
	}

	if entry.Title == "" {
		registerError("title", "can not be empty")
	}

	if entry.Description == "" {
		registerError("description", "can not be empty")
	}

	if entry.FromYear == 0 {
		registerError("fromYear", "can not be 0")
	}

	if entry.ToYear != nil && *entry.ToYear == 0 {
		registerError("toYear", "can not be 0")
	}

	if entry.ToYear != nil && *entry.ToYear < entry.FromYear {
		registerError("toYear", "can not be come before `fromYear`")
	}

	if len(entry.Decades) == 0 {
		registerError("decades", "can not be empty")
	}

	earliestDecade := entry.FromYear - (entry.FromYear % 10) //nolint:gomnd
	remainingDecades := make(map[int]struct{})

	if entry.ToYear == nil {
		remainingDecades[earliestDecade] = struct{}{}
	} else {
		latestDecade := *entry.ToYear - (*entry.ToYear % 10) //nolint:gomnd
		for expectedDecade := earliestDecade; expectedDecade <= latestDecade; expectedDecade += 10 {
			remainingDecades[expectedDecade] = struct{}{}
		}
	}

	for decadeIndex, decade := range entry.Decades {
		if decade%10 != 0 {
			registerError(fmt.Sprintf("decades[%d]", decadeIndex), "is not a decade")
			continue
		}

		if entry.FromYear == 0 || (entry.ToYear != nil && *entry.ToYear == 0) {
			continue
		}

		if decade < earliestDecade {
			registerError(fmt.Sprintf("decades[%d]", decadeIndex), "comes before the decade of `fromYear`")
			continue
		}

		if entry.ToYear != nil {
			if latestDecade := *entry.ToYear - (*entry.ToYear % 10); decade > latestDecade { //nolint:gomnd
				registerError(fmt.Sprintf("decades[%d]", decadeIndex), "comes after the decade of `toYear`")
				continue
			}
		}

		if _, ok := remainingDecades[decade]; !ok {
			registerError(fmt.Sprintf("decades[%d]", decadeIndex), "is in the list more than once")
			continue
		}

		delete(remainingDecades, decade)
	}

	if entry.FromYear != 0 && (entry.ToYear == nil || *entry.ToYear != 0) {
		for expectedButNotFoundDecade := range remainingDecades {
			registerError("decades", fmt.Sprintf("should contain '%d' but doesn't", expectedButNotFoundDecade))
		}
	}

	if len(entry.Files) == 0 {
		registerError("files", "can not be empty")
	}

	for fileIndex, fileEntry := range entry.Files {
		if fileEntry.Name == "" {
			registerError(fmt.Sprintf("files[%d].name", fileIndex), "can not be empty")
		}

		if _, err := cid.Parse(fileEntry.Cid); err != nil {
			registerError(fmt.Sprintf("files[%d].cid", fileIndex), "is not a valid CID")
		}

		if fileEntry.MediaType != nil {
			if _, _, err := mime.ParseMediaType(*fileEntry.MediaType); err != nil {
				registerError(fmt.Sprintf("files[%d].mediaType", fileIndex), "is not a valid media type")
			}
		}

		if fileEntry.Filename != nil && filepath.Ext(*fileEntry.Filename) == "" {
			registerError(fmt.Sprintf("files[%d].filename", fileIndex), "does not have a file extension")
		}
	}

	if len(reasons) == 0 {
		return nil
	} else {
		return InvalidArtifactError{
			FilePath: filePath,
			Reasons:  reasons,
		}
	}
}
