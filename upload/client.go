package upload

import (
	api "github.com/ipfs/go-ipfs-http-client"
	"github.com/multiformats/go-multiaddr"
)

func NewClient(addr string) (*api.HttpApi, error) {
	multiAddr, err := multiaddr.NewMultiaddr(addr)
	if err != nil {
		return nil, err
	}

	return api.NewApi(multiAddr)
}
