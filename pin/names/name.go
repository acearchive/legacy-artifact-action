package names

import (
	"strings"

	"github.com/ipfs/go-cid"
)

const Namespace = "lgbt.acearchive.artifact-action"

type Kind string

const (
	// FileKind is the CID of a file in an artifact.
	FileKind = "files"

	// RootKind is the CID of the root directory containing every file in every
	// artifact.
	RootKind = "roots"
)

func newPinName(pinCid cid.Cid, parents ...string) string {
	var name strings.Builder

	name.WriteString(Namespace)

	for _, segment := range parents {
		name.WriteRune('/')
		name.WriteString(segment)
	}

	name.WriteRune('/')
	name.WriteString(pinCid.String())

	return name.String()
}

func New(pinCid cid.Cid, kind Kind) string {
	return newPinName(pinCid, string(kind))
}
