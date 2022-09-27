package parse

import (
	"errors"
	"fmt"
	"mime"
	"net/url"
	"regexp"
	"strings"

	"github.com/ipfs/go-cid"
)

var ErrInvalidArtifactFiles = errors.New("one or more artifact files are invalid")

// This regex must be kept in sync with the one that validates user input on
// the website.
var fileNameRegex = regexp.MustCompile(`^[\w\d][\w\d-]*[\w\d](\.[\w\d]+)*$`)

type InvalidArtifactReason struct {
	Field  EntryField
	Reason string
}

type InvalidArtifactError struct {
	FilePath string
	Reasons  []InvalidArtifactReason
}

func (e InvalidArtifactError) Error() string {
	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("%s:\n", e.FilePath))
	for _, reason := range e.Reasons {
		builder.WriteString(fmt.Sprintf("  %s %s\n", reason.Field.Literal(), reason.Reason))
	}
	return builder.String()
}

func (f EntryField) At(index int) EntryField {
	return EntryField(fmt.Sprintf("%s[%d]", f, index))
}

func (f EntryField) Of(outer EntryField) EntryField {
	return EntryField(fmt.Sprintf("%s.%s", outer, f))
}

func (f EntryField) Literal() string {
	return fmt.Sprintf("`%s`", f)
}

type ErrorCallback func(field EntryField, reason string)

type FieldValidator func(entry ArtifactEntry, reportError ErrorCallback)

func validateIsNotEmpty(field EntryField, value string, reportError ErrorCallback) {
	if value == "" {
		reportError(field, "can not be empty")
	}
}

func validateHasNoDuplicates(field EntryField, values []string, reportError ErrorCallback) {
	deduplicatedValues := make(map[string]struct{})

	for _, value := range values {
		if _, isDuplicate := deduplicatedValues[value]; isDuplicate {
			reportError(field, "contains duplicates")
		}

		deduplicatedValues[value] = struct{}{}
	}
}

func validateVersion(entry ArtifactEntry, reportError ErrorCallback) {
	if entry.Version != CurrentArtifactVersion {
		reportError(FieldVersion, fmt.Sprintf("must be the current version (%d)", CurrentArtifactVersion))
	}
}

func validateTitle(entry ArtifactEntry, reportError ErrorCallback) {
	validateIsNotEmpty(FieldTitle, entry.Title, reportError)
}

func validateDescription(entry ArtifactEntry, reportError ErrorCallback) {
	validateIsNotEmpty(FieldDescription, entry.Description, reportError)
}

func validateFromYear(entry ArtifactEntry, reportError ErrorCallback) {
	if entry.FromYear == 0 {
		reportError(FieldFromYear, "can not be 0")
	}
}

func validateToYear(entry ArtifactEntry, reportError ErrorCallback) {
	if entry.ToYear != nil && *entry.ToYear == 0 {
		reportError(FieldToYear, "can not be 0")
	}

	if entry.ToYear != nil && *entry.ToYear < entry.FromYear {
		reportError(FieldToYear, fmt.Sprintf("can not be come before %s", FieldFromYear.Literal()))
	}
}

func validateDecades(entry ArtifactEntry, reportError ErrorCallback) {
	if len(entry.Decades) == 0 {
		reportError(FieldDecades, "can not be empty")
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
			reportError(FieldDecades.At(decadeIndex), "is not a decade")
			continue
		}

		if entry.FromYear == 0 || (entry.ToYear != nil && *entry.ToYear == 0) {
			continue
		}

		if decade < earliestDecade {
			reportError(FieldDecades.At(decadeIndex), fmt.Sprintf("comes before the decade of %s", FieldFromYear.Literal()))
			continue
		}

		if entry.ToYear != nil {
			if latestDecade := *entry.ToYear - (*entry.ToYear % 10); decade > latestDecade { //nolint:gomnd
				reportError(FieldDecades.At(decadeIndex), fmt.Sprintf("comes after the decade of %s", FieldToYear.Literal()))
				continue
			}
		}

		if _, ok := remainingDecades[decade]; !ok {
			reportError(FieldDecades.At(decadeIndex), "is in the list more than once")
			continue
		}

		delete(remainingDecades, decade)
	}

	for decadeIndex, decade := range entry.Decades {
		if decadeIndex == 0 {
			continue
		}

		if decade < entry.Decades[decadeIndex-1] {
			reportError(FieldDecades, "is not in chronological order")
			break
		}
	}
}

func validateAliases(entry ArtifactEntry, reportError ErrorCallback) {
	validateHasNoDuplicates(FieldAliases, entry.Aliases, reportError)

	for i, alias := range entry.Aliases {
		if strings.Contains(alias, "/") {
			reportError(FieldAliases.At(i), "can not contain '/'")
		}
	}
}

func validatePeople(entry ArtifactEntry, reportError ErrorCallback) {
	validateHasNoDuplicates(FieldPeople, entry.People, reportError)
}

func validateIdentities(entry ArtifactEntry, reportError ErrorCallback) {
	validateHasNoDuplicates(FieldIdentities, entry.Identities, reportError)
}

func validateFiles(entry ArtifactEntry, reportError ErrorCallback) {
	if len(entry.Files) == 0 && len(entry.Links) == 0 {
		reportError(FieldFiles, fmt.Sprintf("and %s can not both be empty", FieldLinks.Literal()))
	}

	uniqueFiles := make(map[string]struct{}, len(entry.Files))

	for fileIndex, fileEntry := range entry.Files {
		uniqueFiles[fileEntry.Filename] = struct{}{}

		validateIsNotEmpty(FieldFileName.Of(FieldFiles.At(fileIndex)), fileEntry.Name, reportError)

		if _, err := cid.Parse(fileEntry.Cid); err != nil {
			reportError(FieldFileCid.Of(FieldFiles.At(fileIndex)), "is not a valid CID")
		}

		if fileEntry.MediaType != nil {
			if _, _, err := mime.ParseMediaType(*fileEntry.MediaType); err != nil {
				reportError(FieldFileMediaType.Of(FieldFiles.At(fileIndex)), "is not a valid media type")
			}
		}

		if len(fileEntry.Filename) == 0 {
			reportError(FieldFileFilename.Of(FieldFiles.At(fileIndex)), "does not have a file name")
		} else if len(strings.TrimSpace(fileEntry.Filename)) == 0 {
			reportError(FieldFileFilename.Of(FieldFiles.At(fileIndex)), "has a file name that is entirely whitespace")
		} else if !fileNameRegex.MatchString(fileEntry.Filename) {
			reportError(FieldFileFilename.Of(FieldFiles.At(fileIndex)), "contains illegal characters")
		}
	}

	if len(uniqueFiles) < len(entry.Files) {
		reportError(FieldFiles, "contains duplicate file names")
	}
}

func validateLinks(entry ArtifactEntry, reportError ErrorCallback) {
	for linkIndex, linkEntry := range entry.Links {
		validateIsNotEmpty(FieldLinkName.Of(FieldLinks.At(linkIndex)), linkEntry.Name, reportError)

		if linkUrl, err := url.Parse(linkEntry.Url); err != nil {
			reportError(FieldLinkUrl.Of(FieldLinks.At(linkIndex)), "is not a valid URL")
		} else if linkUrl.Scheme != "https" {
			reportError(FieldLinkUrl.Of(FieldLinks.At(linkIndex)), "is not an HTTPS URL")
		}
	}
}

var allValidators = []FieldValidator{
	validateVersion,
	validateTitle,
	validateDescription,
	validateFromYear,
	validateToYear,
	validateDecades,
	validateAliases,
	validatePeople,
	validateIdentities,
	validateFiles,
	validateLinks,
}

func ValidateEntry(entry ArtifactEntry, filePath string) error {
	var reasons []InvalidArtifactReason

	for _, validator := range allValidators {
		validator(entry, func(field EntryField, reason string) {
			reasons = append(reasons, InvalidArtifactReason{
				Field:  field,
				Reason: reason,
			})
		})
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
