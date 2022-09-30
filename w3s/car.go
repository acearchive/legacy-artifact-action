package w3s

import (
	"context"
	"io"

	"github.com/ipfs/go-cid"
	api "github.com/ipfs/go-ipfs-http-client"
	ipld "github.com/ipfs/go-ipld-format"
	"github.com/ipld/go-car"
)

// PackPartialCar creates a car from the given `rootCid`, but only includes the
// first level of links from that root CID. This allows for uploading large
// directories to Web3.Storage as a partial DAG, which avoids the problem of
// uploading large amounts of duplicate data and prevents duplicate data from
// counting against your storage quota.
func PackPartialCar(ctx context.Context, client *api.HttpApi, rootCid cid.Cid) (io.ReadCloser, error) {
	carReader, carWriter := io.Pipe()

	go func() {
		err := car.WriteCarWithWalker(ctx, client.Dag(), []cid.Cid{rootCid}, carWriter, func(node ipld.Node) ([]*ipld.Link, error) {
			if node.Cid() == rootCid {
				return node.Links(), nil
			}

			return nil, nil
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
