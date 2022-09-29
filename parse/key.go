package parse

import "github.com/ipfs/go-cid"

type ContentKey string

type ContentSet = map[ContentKey]struct{}

func ContentKeyFromCid(id cid.Cid) ContentKey {
	return ContentKey(id.Hash().HexString())
}
