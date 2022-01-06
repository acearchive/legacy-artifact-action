package upload

import (
	"context"
	"errors"
	"fmt"
	"github.com/ipfs/go-cid"
	"github.com/web3-storage/go-w3s-client"
	"io"
)

func listExistingCids(ctx context.Context, client w3s.Client) (map[cid.Cid]struct{}, error) {
	iter, err := client.List(ctx)
	if err != nil {
		return nil, err
	}

	var cidSet = make(map[cid.Cid]struct{})

	for {
		status, err := iter.Next()
		if err != nil && !errors.Is(err, io.EOF) {
			return nil, err
		} else if errors.Is(err, io.EOF) || status == nil {
			break
		}

		cidSet[status.Cid] = struct{}{}
	}

	return cidSet, nil
}

func Content(ctx context.Context, token string, cidList []cid.Cid) error {
	w3sClient, err := w3s.NewClient(w3s.WithToken(token))
	if err != nil {
		return err
	}

	ipfsClient, err := NewClient()
	if err != nil {
		return err
	}

	existingCids, err := listExistingCids(ctx, w3sClient)
	if err != nil {
		return err
	}

	for _, id := range cidList {
		if _, exists := existingCids[id]; exists {
			fmt.Printf("Skipping previously archived content: %s\n", id.String())
			continue
		}

		car, err := PackCar(ctx, ipfsClient, id)
		if err != nil {
			return err
		}

		if _, err := w3sClient.PutCar(ctx, car); err != nil {
			return err
		}
	}

	return nil
}
