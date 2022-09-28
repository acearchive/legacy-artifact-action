package pin

import (
	"context"

	"github.com/frawleyskid/w3s-upload/cfg"
	"github.com/frawleyskid/w3s-upload/logger"
	"github.com/frawleyskid/w3s-upload/parse"
	"github.com/ipfs/go-cid"
	pinclient "github.com/ipfs/go-pinning-service-http-client"
)

func Pin(ctx context.Context, endpoint, token string, cidList []cid.Cid) error {
	client := pinclient.NewClient(endpoint, token)

	existingPins, errChan := client.Ls(ctx)

	existingContent := make(map[parse.ContentKey]struct{})

	for pinStatus := range existingPins {
		existingContent[parse.ContentKeyFromCid(pinStatus.GetPin().GetCid())] = struct{}{}
	}

	if err := <-errChan; err != nil {
		return err
	}

	logger.Printf("Found %d pins\n", len(existingContent))

	filesToUpload := make([]cid.Cid, 0, len(cidList))

	for _, id := range cidList {
		if _, alreadyUploaded := existingContent[parse.ContentKeyFromCid(id)]; !alreadyUploaded {
			filesToUpload = append(filesToUpload, id)
		}
	}

	logger.Printf("Skipping %d files that are already pinned\n", len(cidList)-len(filesToUpload))
	logger.Printf("Pinning %d files\n", len(filesToUpload))

	for currentIndex, currentCid := range filesToUpload {
		logger.Printf("Pinning (%d/%d): %s\n", currentIndex+1, len(filesToUpload), currentCid.String())

		if cfg.DryRun() {
			continue
		}

		if _, err := client.Add(ctx, currentCid); err != nil {
			return err
		}
	}

	return nil
}
