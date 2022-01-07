package upload

import (
	"context"
	"errors"
	"fmt"
	"github.com/ipfs/go-cid"
	"github.com/web3-storage/go-w3s-client"
	"io"
	"time"
)

const maxPageSize = 25

type contentKey string

func contentKeyFromCid(id cid.Cid) contentKey {
	return contentKey(id.Hash().HexString())
}

func listExistingCids(ctx context.Context, client w3s.Client) (map[contentKey]struct{}, error) {
	paginationParams := []w3s.ListOption{w3s.WithMaxResults(maxPageSize)}
	cidSet := make(map[contentKey]struct{})

	for {
		iter, err := client.List(ctx, paginationParams...)
		if err != nil {
			return nil, err
		}

		currentPageSize := 0

		// The endpoint returns items in order from newest to oldest.
		var oldestTimestamp time.Time

		for {
			status, err := iter.Next()
			if err != nil && !errors.Is(err, io.EOF) {
				return nil, err
			} else if errors.Is(err, io.EOF) {
				break
			}

			currentPageSize++
			oldestTimestamp = status.Created
			cidSet[contentKeyFromCid(status.Cid)] = struct{}{}
		}

		if currentPageSize < maxPageSize {
			break
		}

		paginationParams = []w3s.ListOption{w3s.WithMaxResults(maxPageSize), w3s.WithBefore(oldestTimestamp)}
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

	fmt.Printf("Found %d files in Web3.Storage\n", len(existingCids))

	filesToUpload := make([]cid.Cid, 0, len(cidList))

	// We need to check if content already exists in Web3.Storage before we
	// store it again, but we should compare CIDs by their multihash rather
	// than directly so that a CIDv0 and a CIDv1 aren't detected as different
	// contents.
	for _, id := range cidList {
		if _, alreadyUploaded := existingCids[contentKeyFromCid(id)]; !alreadyUploaded {
			filesToUpload = append(filesToUpload, id)
		}
	}

	fmt.Printf("Skipping %d files that are already in Web3.Storage\n", len(cidList)-len(filesToUpload))
	fmt.Printf("Uploading %d files\n", len(filesToUpload))

	for currentIndex, currentCid := range filesToUpload {
		fmt.Printf("Uploading (%d/%d): %s\n", currentIndex+1, len(filesToUpload), currentCid.String())

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
