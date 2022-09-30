package pin

import (
	"context"
	"fmt"
	"time"

	"github.com/acearchive/artifact-action/cfg"
	"github.com/acearchive/artifact-action/logger"
	"github.com/acearchive/artifact-action/parse"
	"github.com/acearchive/artifact-action/pin/names"
	"github.com/ipfs/go-cid"
	pinning "github.com/ipfs/go-pinning-service-http-client"
)

// This is deliberately larger than necessary in order to mitigate problems
// associated with paging. See `filterAlreadyPinned`.
const pageLimit = 100

// These are the statuses we filter on when checking if a CID is already pinned
// by the service.
func filterExistingPinStatus() pinning.LsOption {
	return pinning.PinOpts.FilterStatus(
		pinning.StatusQueued,
		pinning.StatusPinning,
		pinning.StatusPinned,
	)
}

func oneshotChan[E any](c chan E) <-chan bool {
	outChan := make(chan bool, 1)

	go func() {
		_, isOpen := <-c
		outChan <- isOpen
		close(outChan)
	}()

	return outChan
}

type NotPinnedCIDList []cid.Cid

type pinInfo struct {
	Cid     cid.Cid
	Created time.Time
}

// filterAlreadyPinned returns the subset of the given CIDs have not already been
// pinned and need to be pinned.
//
// The way that we have to go about this is wonky. As of time of writing,
// pagination is fundamentally broken in the pinning services API, both in
// terms of the spec and in terms of common implementations. See this issue for
// details:
//
// https://github.com/ipfs/kubo/issues/8762
//
// We can work around the problem by trying our best to get as many CIDs from
// the remote via pagination as we can, and then from there we can individually
// check any that slipped through.
func filterAlreadyPinned(ctx context.Context, client *pinning.Client, fileCids parse.DeduplicatedCIDList) (NotPinnedCIDList, error) {
	pinnedContent := parse.NewContentSet(len(fileCids))

	// Technically a new pin could be created between "now" and when this
	// request actually hits the server, but that's not much of a problem since
	// pagination is broken and we're probably not going to catch all the CIDs
	// by paging anyways.
	cursor := time.Now().UTC()

	// Attempt to page over as many pins as possible.
	for {
		pins, err := client.LsSync(
			ctx,
			pinning.PinOpts.FilterStatus(pinning.StatusPinned),
			pinning.PinOpts.Limit(pageLimit),
			// pinning.PinOpts.FilterBefore(cursor),
		)
		if err != nil {
			return nil, err
		}

		nextCursor := cursor

		fmt.Printf("pins = %d\n", len(pins))
		for _, pinStatus := range pins {
			pinnedContent.Insert(pinStatus.GetPin().GetCid())
			nextCursor = pinStatus.GetCreated()
		}

		nextCursor = nextCursor.UTC()

		if !nextCursor.Before(cursor) {
			break
		}

		cursor = nextCursor
	}

	fmt.Printf("len(pinnedContent) = %d\n", len(pinnedContent))

	// The CIDs of files in the repository which are not already pinned as far
	// as we can tell based on the CIDs we got via paging. We're going to
	// double-check each of these individually in case they were missed due to
	// pagination bugs.
	cidsToRecheck := pinnedContent.Difference(fileCids)

	fmt.Printf("len(cidsToRecheck) = %d\n", len(cidsToRecheck))

	cidsToPin := make(NotPinnedCIDList, 0, len(cidsToRecheck))

	for _, currentCid := range cidsToRecheck {
		// Check if this CID is already pinned. We also filter on the name
		// because the naming scheme namespaces these CIDs as files in
		// artifacts which were pinned by this tool. If other tools/services
		// are also pinning to this remote, then we don't want to step on each
		// others' toes. We can't assume that a CID is pinned forever if
		// another service may decide to unpin it later.
		pins, err := client.LsSync(
			ctx,
			filterExistingPinStatus(),
			pinning.PinOpts.FilterCIDs(currentCid),
			pinning.PinOpts.FilterName(names.New(currentCid, names.FileKind)),
			pinning.PinOpts.Limit(1), // It's possible to have multiple pins for the same CID.
		)
		if err != nil {
			return nil, err
		}

		if len(pins) > 0 {
			fmt.Println("IT'S PINNED")
			cidsToPin = append(cidsToPin, currentCid)
		} else {
			fmt.Println("IT'S NOT PINNED")
		}
	}

	numAlreadyPinned := len(fileCids) - len(cidsToPin)
	logger.Printf("Skipping %d files that are already pinned\n", numAlreadyPinned)

	return cidsToPin, nil
}

func pinCid(ctx context.Context, client *pinning.Client, cidToPin cid.Cid, kind names.Kind) error {
	if !cfg.DryRun() {
		if _, err := client.Add(ctx, cidToPin, pinning.PinOpts.WithName(names.New(cidToPin, kind))); err != nil {
			return err
		}
	}

	return nil
}

func pinFiles(ctx context.Context, client *pinning.Client, cidsToPin NotPinnedCIDList, rootCid cid.Cid) error {
	logger.Printf("Pinning %d files\n", len(cidsToPin))

	for currentIndex, currentCid := range cidsToPin {
		logger.Printf("Pinning (%d/%d): %s\n", currentIndex+1, len(cidsToPin), currentCid.String())

		if err := pinCid(ctx, client, currentCid, names.FileKind); err != nil {
			return err
		}
	}

	logger.Printf("\nPinning the root directory: /ipfs/%s\n", rootCid.String())

	return pinCid(ctx, client, rootCid, names.RootKind)
}

func Pin(ctx context.Context, endpoint, token string, fileCids parse.DeduplicatedCIDList, rootCid cid.Cid) error {
	client := pinning.NewClient(endpoint, token)

	cidsToPin, err := filterAlreadyPinned(ctx, client, fileCids)
	if err != nil {
		return err
	}

	return pinFiles(ctx, client, cidsToPin, rootCid)
}
