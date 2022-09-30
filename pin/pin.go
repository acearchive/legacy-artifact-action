package pin

import (
	"context"

	"github.com/acearchive/artifact-action/cfg"
	"github.com/acearchive/artifact-action/logger"
	"github.com/acearchive/artifact-action/parse"
	"github.com/ipfs/go-cid"
	pinclient "github.com/ipfs/go-pinning-service-http-client"
)

func listPins(ctx context.Context, endpoint, token string) (parse.ContentSet, error) {
	client := pinclient.NewClient(endpoint, token)

	existingPins, errChan := client.Ls(ctx, pinclient.PinOpts.FilterStatus(pinclient.StatusPinned))

	existingContent := make(parse.ContentSet)

	for pinStatus := range existingPins {
		existingContent[parse.ContentKeyFromCid(pinStatus.GetPin().GetCid())] = struct{}{}
	}

	if err := <-errChan; err != nil {
		return nil, err
	}

	logger.Printf("Found %d pins\n", len(existingContent))

	return existingContent, nil
}

func pinDeduplicated(ctx context.Context, endpoint, token string, allCids []cid.Cid, alreadyPinned parse.ContentSet) error {
	client := pinclient.NewClient(endpoint, token)

	filesToUpload := make([]cid.Cid, 0, len(allCids))

	for _, id := range allCids {
		if _, alreadyUploaded := alreadyPinned[parse.ContentKeyFromCid(id)]; !alreadyUploaded {
			filesToUpload = append(filesToUpload, id)
		}
	}

	logger.Printf("Skipping %d files that are already pinned\n", len(allCids)-len(filesToUpload))
	logger.Printf("Pinning %d files\n", len(filesToUpload))

	for currentIndex, currentCid := range filesToUpload {
		logger.Printf("Pinning (%d/%d): %s\n", currentIndex+1, len(filesToUpload), currentCid.String())

		if !cfg.DryRun() {
			if _, err := client.Add(ctx, currentCid); err != nil {
				return err
			}
		}
	}

	return nil
}

func Pin(ctx context.Context, endpoint, token string, cidList []cid.Cid) error {
	existingContent, err := listPins(ctx, endpoint, token)
	if err != nil {
		return err
	}

	return pinDeduplicated(ctx, endpoint, token, cidList, existingContent)
}
