package parse

import "github.com/ipfs/go-cid"

type ContentKey string

func newContentKey(id cid.Cid) ContentKey {
	return ContentKey(id.Hash().HexString())
}

// ContentSet is a set of content hashes, which allows us to deduplicate
// between CIDs of different versions.
type ContentSet map[ContentKey]struct{}

func NewContentSet(cap int) ContentSet {
	return make(ContentSet, cap)
}

func (s ContentSet) Insert(c cid.Cid) {
	s[newContentKey(c)] = struct{}{}
}

func (s ContentSet) Contains(c cid.Cid) bool {
	_, exists := s[newContentKey(c)]
	return exists
}

// Difference returns the cids in `cidList` which are not in this set.
func (s ContentSet) Difference(cidList []cid.Cid) []cid.Cid {
	out := make([]cid.Cid, 0, len(cidList))

	for _, currentCid := range cidList {
		if !s.Contains(currentCid) {
			out = append(out, currentCid)
		}
	}

	return out
}
