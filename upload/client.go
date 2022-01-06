package upload

import (
	api "github.com/ipfs/go-ipfs-http-client"
	"github.com/multiformats/go-multiaddr"
)

const defaultClientApiAddr = "/ip4/127.0.0.1/tcp/5001"

func NewClient() (*api.HttpApi, error) {
	addr, err := multiaddr.NewMultiaddr(defaultClientApiAddr)
	if err != nil {
		return nil, err
	}

	return api.NewApi(addr)
}
