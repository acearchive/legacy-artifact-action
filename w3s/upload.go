package w3s

import (
	"context"

	"github.com/acearchive/artifact-action/cfg"
	"github.com/acearchive/artifact-action/client"
	"github.com/acearchive/artifact-action/logger"
	"github.com/acearchive/artifact-action/parse"
	"github.com/acearchive/artifact-action/pin"
	"github.com/ipfs/go-cid"
	"github.com/web3-storage/go-w3s-client"
)

const w3sPinApiEndpoint = "https://api.web3.storage"

func uploadWithoutPinning(ctx context.Context, token string, w3sClient w3s.Client, ipfsClientGuard *client.Guard, cidList []cid.Cid) error {
	ipfsClient := ipfsClientGuard.Lock()
	defer ipfsClientGuard.Unlock()

	existingCids, err := listExistingCids(ctx, token)
	if err != nil {
		return err
	}

	logger.Printf("Found %d files in Web3.Storage\n", len(existingCids))

	cidsToPin := make([]cid.Cid, 0, len(cidList))

	// We need to check if content already exists in Web3.Storage before we
	// store it again, but we should compare CIDs by their multihash rather
	// than directly so that a CIDv0 and a CIDv1 aren't detected as different
	// contents.
	for _, id := range cidList {
		if _, alreadyUploaded := existingCids[parse.ContentKeyFromCid(id)]; !alreadyUploaded {
			cidsToPin = append(cidsToPin, id)
		}
	}

	logger.Printf("Skipping %d files that are already in Web3.Storage\n", len(cidList)-len(cidsToPin))
	logger.Printf("Uploading %d files\n", len(cidsToPin))

	for currentIndex, currentCid := range cidsToPin {
		logger.Printf("Uploading (%d/%d): %s\n", currentIndex+1, len(cidsToPin), currentCid.String())

		car, err := PackCar(ctx, ipfsClient, currentCid)
		if err != nil {
			return err
		}

		if cfg.DryRun() {
			continue
		}

		if _, err := w3sClient.PutCar(ctx, car); err != nil {
			return err
		}
	}

	return nil
}

func Upload(ctx context.Context, token string, cidList []cid.Cid, directoryCid cid.Cid) error {
	w3sClient, err := w3s.NewClient(w3s.WithToken(token))
	if err != nil {
		return err
	}

	ipfsClientGuard, err := client.New()
	if err != nil {
		return err
	}

	if !cfg.W3SPin() {
		cidsToUpload := make([]cid.Cid, len(cidList), len(cidList)+1)
		copy(cidsToUpload, cidList)
		cidsToUpload = append(cidsToUpload, directoryCid)

		return uploadWithoutPinning(ctx, token, w3sClient, ipfsClientGuard, cidsToUpload)
	}

	existingContent, err := listExistingCids(ctx, token)
	if err != nil {
		return err
	}

	if err := pin.Pin(ctx, w3sPinApiEndpoint, token, cidList, existingContent); err != nil {
		return err
	}

	ipfsClient := ipfsClientGuard.Lock()
	defer ipfsClientGuard.Unlock()

	car, err := PackPartialCar(ctx, ipfsClient, directoryCid, existingContent)
	if err != nil {
		return err
	}

	if !cfg.DryRun() {
		if _, err := w3sClient.PutCar(ctx, car); err != nil {
			return err
		}
	}

	return nil
}
