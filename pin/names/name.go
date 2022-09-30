package names

import (
	"strings"

	"github.com/ipfs/go-cid"
)

const (
	filesName = "files"
	rootsName = "roots"
	prefix    = "lgbt.acearchive.artifact-action"
)

const (
	FilesPrefix = prefix + "/" + filesName
	RootsPrefix = prefix + "/" + rootsName
)

func newPinName(pinCid cid.Cid, parents ...string) string {
	var name strings.Builder

	name.WriteString(prefix)

	for _, segment := range parents {
		name.WriteRune('/')
		name.WriteString(segment)
	}

	name.WriteRune('/')
	name.WriteString(pinCid.String())

	return name.String()
}

func ForFile(pinCid cid.Cid) string {
	return newPinName(pinCid, filesName)
}

func ForRoot(pinCid cid.Cid) string {
	return newPinName(pinCid, rootsName)
}
