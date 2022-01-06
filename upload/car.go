package upload

import (
	"context"
	"github.com/ipfs/go-blockservice"
	"github.com/ipfs/go-cid"
	"github.com/ipfs/go-datastore"
	"github.com/ipfs/go-datastore/mount"
	syncds "github.com/ipfs/go-datastore/sync"
	flatfs "github.com/ipfs/go-ds-flatfs"
	blockstore "github.com/ipfs/go-ipfs-blockstore"
	chunk "github.com/ipfs/go-ipfs-chunker"
	offline "github.com/ipfs/go-ipfs-exchange-offline"
	ipld "github.com/ipfs/go-ipld-format"
	dag "github.com/ipfs/go-merkledag"
	"github.com/ipfs/go-unixfs/importer"
	"github.com/ipld/go-car"
	"io"
	"io/ioutil"
)

type LocalService struct {
	ipld.DAGService
}

func NewLocalService() (*LocalService, error) {
	storePath, err := ioutil.TempDir("", "w3-upload-ds-*")
	if err != nil {
		return nil, err
	}

	if err := flatfs.Create(storePath, flatfs.IPFS_DEF_SHARD); err != nil {
		return nil, err
	}

	fsStore, err := flatfs.Open(storePath, false)
	if err != nil {
		return nil, err
	}

	mountStore := mount.New([]mount.Mount{{
		Prefix:    datastore.NewKey("/blocks"),
		Datastore: fsStore,
	}})

	blockStore := blockstore.NewBlockstore(syncds.MutexWrap(mountStore))
	blockService := blockservice.New(blockStore, offline.Exchange(blockStore))

	return &LocalService{dag.NewDAGService(blockService)}, nil
}

func (s *LocalService) PackCar(ctx context.Context, content io.ReadCloser) (io.ReadCloser, error) {
	carReader, carWriter := io.Pipe()

	node, err := importer.BuildDagFromReader(s, chunk.DefaultSplitter(content))
	if err != nil {
		return nil, err
	}

	if err := content.Close(); err != nil {
		return nil, err
	}

	go func() {
		err = car.WriteCar(ctx, s, []cid.Cid{node.Cid()}, carWriter)
		if err != nil {
			_ = carWriter.CloseWithError(err)
			return
		}
		_ = carWriter.Close()
	}()

	return carReader, nil
}
