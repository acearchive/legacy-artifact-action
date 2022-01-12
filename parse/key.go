package parse

import "github.com/ipfs/go-cid"

type ContentKey string

func ContentKeyFromCid(id cid.Cid) ContentKey {
	return ContentKey(id.Hash().HexString())
}
