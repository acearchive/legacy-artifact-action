package client

import (
	"errors"
	"sync"

	"github.com/acearchive/artifact-action/cfg"
	api "github.com/ipfs/go-ipfs-http-client"
	"github.com/multiformats/go-multiaddr"
)

var ErrNoIpfsAPIAddr = errors.New("the multiaddr of the IPFS API was not provided")

var client *Guard

type Guard struct {
	client *api.HttpApi
	mutex  *sync.Mutex
}

func (g Guard) Lock() *api.HttpApi {
	g.mutex.Lock()
	return g.client
}

func (g Guard) Unlock() {
	g.mutex.Unlock()
}

func New() (*Guard, error) {
	if client != nil {
		return client, nil
	}

	if cfg.IpfsAPI() == "" {
		return nil, ErrNoIpfsAPIAddr
	}

	multiaddr, err := multiaddr.NewMultiaddr(cfg.IpfsAPI())
	if err != nil {
		return nil, err
	}

	apiClient, err := api.NewApi(multiaddr)
	if err != nil {
		return nil, err
	}

	client = &Guard{
		client: apiClient,
		mutex:  &sync.Mutex{},
	}

	return client, nil
}
