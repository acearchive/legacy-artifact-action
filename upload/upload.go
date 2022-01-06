package upload

import (
	"context"
	"github.com/ipfs/go-cid"
	"github.com/web3-storage/go-w3s-client"
)

func listExistingCids(ctx context.Context, client w3s.Client) (map[cid.Cid]struct{}, error) {
	iter, err := client.List(ctx)
	if err != nil {
		return nil, err
	}

	var cidSet = make(map[cid.Cid]struct{})

	for {
		status, err := iter.Next()
		if err != nil {
			return nil, err
		} else if status == nil {
			break
		}

		cidSet[status.Cid] = struct{}{}
	}

	return cidSet, nil
}

func Content(ctx context.Context, token string, cidList []cid.Cid) error {
	service, err := NewLocalService()
	if err != nil {
		return err
	}

	client, err := w3s.NewClient(w3s.WithToken(token))
	if err != nil {
		return err
	}

	existingCids, err := listExistingCids(ctx, client)
	if err != nil {
		return err
	}

	for _, id := range cidList {
		if _, exists := existingCids[id]; exists {
			continue
		}

		content, err := DownloadContent(ctx, id)
		if err != nil {
			return err
		}

		car, err := service.PackCar(ctx, content)
		if err != nil {
			return err
		}

		if _, err := client.PutCar(ctx, car); err != nil {
			return err
		}
	}

	return nil
}
