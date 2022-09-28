package w3s

import (
	"context"
	"io"

	"github.com/ipfs/go-cid"
	api "github.com/ipfs/go-ipfs-http-client"
	"github.com/ipld/go-car"
)

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
