package pin

import (
	"context"

	"github.com/acearchive/artifact-action/cfg"
	"github.com/acearchive/artifact-action/logger"
	"github.com/acearchive/artifact-action/parse"
	"github.com/ipfs/go-cid"
	pinning "github.com/ipfs/go-pinning-service-http-client"
)

// We add a metadata entry to the pin for the root directory. This enables us
// to programmatically unpin old roots. While theoretically old roots should be
// deduplicated and take up no additional space, many IPFS pinning services do
// not deduplicate data like this, at least when it comes to calculating your
// usage towards your quota.
//
// This metadata key is intended to be globally unique so that you can query
// for all the roots in a pinning service, even if you share it with other
// services.
const rootMetaKey = "lgbt.acearchive.artifact-action.root"

func rootMeta() map[string]string {
	return map[string]string{rootMetaKey: "true"}
}

func listPins(ctx context.Context, client *pinning.Client) (parse.ContentSet, error) {
	existingPins, errChan := client.Ls(ctx, pinning.PinOpts.FilterStatus(pinning.StatusPinned))

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

func pinDeduplicated(ctx context.Context, client *pinning.Client, fileCids []cid.Cid, rootCid cid.Cid, alreadyPinned parse.ContentSet) error {

	filesToUpload := make([]cid.Cid, 0, len(fileCids))

	for _, id := range fileCids {
		if _, alreadyUploaded := alreadyPinned[parse.ContentKeyFromCid(id)]; !alreadyUploaded {
			filesToUpload = append(filesToUpload, id)
		}
	}

	logger.Printf("Skipping %d files that are already pinned\n", len(fileCids)-len(filesToUpload))
	logger.Printf("Pinning %d files\n", len(filesToUpload))

	for currentIndex, currentCid := range filesToUpload {
		logger.Printf("Pinning (%d/%d): %s\n", currentIndex+1, len(filesToUpload), currentCid.String())

		if !cfg.DryRun() {
			if _, err := client.Add(ctx, currentCid); err != nil {
				return err
			}
		}
	}

	logger.Printf("\nPinning the root directory: /ipfs/%s\n", rootCid.String())

	if !cfg.DryRun() {
		if _, err := client.Add(ctx, rootCid, pinning.PinOpts.AddMeta(rootMeta())); err != nil {
			return err
		}
	}

	return nil
}

func Pin(ctx context.Context, endpoint, token string, fileCids []cid.Cid, rootCid cid.Cid) error {
	client := pinning.NewClient(endpoint, token)

	existingContent, err := listPins(ctx, client)
	if err != nil {
		return err
	}

	return pinDeduplicated(ctx, client, fileCids, rootCid, existingContent)
}
