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

// uploadWithoutPinning uses the W3S client library to upload files by wrapping
// them in a CAR archive. Since files submitted to Ace Archive via CID will
// already be on the IPFS network, this is less efficient than using the
// pinning API because we need to download each file over the IPFS network,
// pack it into a CAR locally, and then reupload it.
func uploadWithoutPinning(ctx context.Context, token string, w3sClient w3s.Client, ipfsClientGuard *client.Guard, cidList []cid.Cid, existingContent parse.ContentSet) error {
	ipfsClient := ipfsClientGuard.Lock()
	defer ipfsClientGuard.Unlock()

	cidsToPin := make([]cid.Cid, 0, len(cidList))

	// We need to check if content already exists in Web3.Storage before we
	// store it again, but we should compare CIDs by their multihash rather
	// than directly so that a CIDv0 and a CIDv1 aren't detected as different
	// contents.
	for _, id := range cidList {
		if _, alreadyUploaded := existingContent[parse.ContentKeyFromCid(id)]; !alreadyUploaded {
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

		if !cfg.DryRun() {
			if _, err := w3sClient.PutCar(ctx, car); err != nil {
				return err
			}
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

	existingContent, err := listExistingCids(ctx, token)
	if err != nil {
		return err
	}

	if cfg.W3SPin() {
		if err := pin.Pin(ctx, w3sPinApiEndpoint, token, cidList, existingContent); err != nil {
			return err
		}
	} else {
		if err := uploadWithoutPinning(ctx, token, w3sClient, ipfsClientGuard, cidList, existingContent); err != nil {
			return err
		}
	}

	ipfsClient := ipfsClientGuard.Lock()
	defer ipfsClientGuard.Unlock()

	logger.Println("Uploading root directory as a partial DAG.")

	car, err := PackPartialCar(ctx, ipfsClient, directoryCid)
	if err != nil {
		return err
	}

	if !cfg.DryRun() {
		if err := uploadPartialDAG(ctx, token, car); err != nil {
			return err
		}
	}

	return nil
}
