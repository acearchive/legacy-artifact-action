package w3s

import (
	"context"
	"io"

	"github.com/acearchive/artifact-action/parse"
	"github.com/ipfs/go-cid"
	api "github.com/ipfs/go-ipfs-http-client"
	ipld "github.com/ipfs/go-ipld-format"
	"github.com/ipld/go-car"
)

func PackPartialCar(ctx context.Context, client *api.HttpApi, rootCid cid.Cid, cidsToSkip parse.ContentSet) (io.ReadCloser, error) {
	carReader, carWriter := io.Pipe()

	go func() {
		err := car.WriteCarWithWalker(ctx, client.Dag(), []cid.Cid{rootCid}, carWriter, func(node ipld.Node) ([]*ipld.Link, error) {
			linksInNode := node.Links()
			linksToFollow := make([]*ipld.Link, 0, len(linksInNode))

			for _, link := range linksInNode {
				if _, exists := cidsToSkip[parse.ContentKeyFromCid(link.Cid)]; !exists {
					linksToFollow = append(linksToFollow, link)
				}
			}

			return linksToFollow, nil
		})
		if err != nil {
			_ = carWriter.CloseWithError(err)
			return
		}

		_ = carWriter.Close()
	}()

	return carReader, nil
}

func PackCar(ctx context.Context, client *api.HttpApi, id cid.Cid) (io.ReadCloser, error) {
	carReader, carWriter := io.Pipe()

	go func() {
		err := car.WriteCar(ctx, client.Dag(), []cid.Cid{id}, carWriter)
		if err != nil {
			_ = carWriter.CloseWithError(err)
			return
		}

		_ = carWriter.Close()
	}()

	return carReader, nil
}
