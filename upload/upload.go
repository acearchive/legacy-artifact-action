package upload

import (
	"context"
	"errors"
	"fmt"
	"github.com/ipfs/go-cid"
	"github.com/web3-storage/go-w3s-client"
	"io"
)

type contentKey string

func contentKeyFromCid(id cid.Cid) contentKey {
	return contentKey(id.Hash().HexString())
}

func listExistingCids(ctx context.Context, client w3s.Client) (map[contentKey]struct{}, error) {
	iter, err := client.List(ctx)
	if err != nil {
		return nil, err
	}

	var cidSet = make(map[contentKey]struct{})

	for {
		status, err := iter.Next()
		if err != nil && !errors.Is(err, io.EOF) {
			return nil, err
		} else if errors.Is(err, io.EOF) || status == nil {
			break
		}

		cidSet[contentKeyFromCid(status.Cid)] = struct{}{}
	}

	return cidSet, nil
}

func Content(ctx context.Context, token, apiAddr string, cidList []cid.Cid) error {
	w3sClient, err := w3s.NewClient(w3s.WithToken(token))
	if err != nil {
		return err
	}

	ipfsClient, err := NewClient(apiAddr)
	if err != nil {
		return err
	}

	existingCids, err := listExistingCids(ctx, w3sClient)
	if err != nil {
		return err
	}

	for _, currentCid := range cidList {
		// We need to check if content already exists in Web3.Storage before we
		// store it again, but we should compare CIDs by their multihash rather
		// than directly so that a CIDv0 and a CIDv1 aren't detected as
		// different contents.
		if _, exists := existingCids[contentKeyFromCid(currentCid)]; exists {
			fmt.Printf("Skipping previously archived content: %s\n", currentCid.String())
			continue
		}

		car, err := PackCar(ctx, ipfsClient, currentCid)
		if err != nil {
			return err
		}

		if _, err := w3sClient.PutCar(ctx, car); err != nil {
			return err
		}
	}

	return nil
}
