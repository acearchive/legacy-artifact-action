package upload

import (
	api "github.com/ipfs/go-ipfs-http-client"
)

func NewClient() (*api.HttpApi, error) {
	addr, err := api.ApiAddr(api.DefaultPathRoot)
	if err != nil {
		return nil, err
	}

	return api.NewApi(addr)
}
