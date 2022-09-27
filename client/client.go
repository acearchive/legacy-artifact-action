package client

import (
	"errors"
	"sync"

	"github.com/frawleyskid/w3s-upload/cfg"
	api "github.com/ipfs/go-ipfs-http-client"
	"github.com/multiformats/go-multiaddr"
)

var ErrNoIpfsApiAddr = errors.New("the multiaddr of the IPFS API was not provided")

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

	if !cfg.DoIpfs() {
		return nil, ErrNoIpfsApiAddr
	}

	multiaddr, err := multiaddr.NewMultiaddr(cfg.IpfsApi())
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
