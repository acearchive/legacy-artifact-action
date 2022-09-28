package w3s

import (
	"context"

	"github.com/acearchive/artifact-action/cfg"
	"github.com/acearchive/artifact-action/client"
	"github.com/acearchive/artifact-action/logger"
	"github.com/acearchive/artifact-action/parse"
	"github.com/ipfs/go-cid"
	"github.com/web3-storage/go-w3s-client"
)

func Upload(ctx context.Context, token string, cidList []cid.Cid) error {
	w3sClient, err := w3s.NewClient(w3s.WithToken(token))
	if err != nil {
		return err
	}

	ipfsClientGuard, err := client.New()
	if err != nil {
		return err
	}

	ipfsClient := ipfsClientGuard.Lock()
	defer ipfsClientGuard.Unlock()

	existingCids, err := listExistingCids(ctx, token)
	if err != nil {
		return err
	}

	logger.Printf("Found %d files in Web3.Storage\n", len(existingCids))

	filesToUpload := make([]cid.Cid, 0, len(cidList))

	// We need to check if content already exists in Web3.Storage before we
	// store it again, but we should compare CIDs by their multihash rather
	// than directly so that a CIDv0 and a CIDv1 aren't detected as different
	// contents.
	for _, id := range cidList {
		if _, alreadyUploaded := existingCids[parse.ContentKeyFromCid(id)]; !alreadyUploaded {
			filesToUpload = append(filesToUpload, id)
		}
	}

	logger.Printf("Skipping %d files that are already in Web3.Storage\n", len(cidList)-len(filesToUpload))
	logger.Printf("Uploading %d files\n", len(filesToUpload))

	for currentIndex, currentCid := range filesToUpload {
		logger.Printf("Uploading (%d/%d): %s\n", currentIndex+1, len(filesToUpload), currentCid.String())

		if cfg.DryRun() {
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
