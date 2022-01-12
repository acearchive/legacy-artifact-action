package w3s

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/frawleyskid/w3s-upload/parse"
	"github.com/ipfs/go-cid"
	"net/http"
	"time"
)

const iso8601 = "2006-01-02T15:04:05.999Z07:00"

var ErrHttpStatus = errors.New("unexpected response status")

// In v0.0.5 of github.com/web3-storage/go-w3s-client, I've been experiencing an
// issue where `Client.List` doesn't return the last page, so for the time-being
// I'm reimplementing it using the HTTP API.

type w3sStatus struct {
	Cid     string    `json:"cid"`
	Created time.Time `json:"created"`
}

func requestUploads(ctx context.Context, token string, before time.Time) ([]w3sStatus, error) {
	requestUrl := fmt.Sprintf("https://api.web3.storage/user/uploads/?before=%s", before.Format(iso8601))

	req, err := http.NewRequestWithContext(ctx, "GET", requestUrl, http.NoBody)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: %d", ErrHttpStatus, res.StatusCode)
	}

	var page []w3sStatus

	decoder := json.NewDecoder(res.Body)

	if err := decoder.Decode(&page); err != nil {
		return nil, err
	}

	if err := res.Body.Close(); err != nil {
		return nil, err
	}

	return page, nil
}

func listExistingCids(ctx context.Context, token string) (map[parse.ContentKey]struct{}, error) {
	cidSet := make(map[parse.ContentKey]struct{})
	pagingParameter := time.Now()

	for {
		page, err := requestUploads(ctx, token, pagingParameter)
		if err != nil {
			return nil, err
		}

		if len(page) == 0 {
			break
		}

		for _, status := range page {
			currentCid, err := cid.Parse(status.Cid)
			if err != nil {
				return nil, err
			}

			pagingParameter = status.Created
			cidSet[parse.ContentKeyFromCid(currentCid)] = struct{}{}
		}
	}

	return cidSet, nil
}
