package car

import (
	"os"
	"path/filepath"

	"github.com/ipfs/go-cid"
	"github.com/ipfs/go-unixfsnode"
	storagecar "github.com/ipld/go-car/v2/storage"
	"github.com/ipld/go-ipld-prime"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"
)

func LinkSystem(carFile *os.File) (ipld.LinkSystem, cid.Cid, error) {
	storage, err := storagecar.OpenReadable(carFile)
	if err != nil {
		return ipld.LinkSystem{}, cid.Undef, err
	}

	var root cid.Cid

	if len(storage.Roots()) == 1 {
		root = storage.Roots()[0]
	} else {
		// attempt infer from filename, but not fatal if we don't, caller may be
		// able to get it from another source
		cidStr := filepath.Base(carFile.Name())
		cidStr = cidStr[:len(cidStr)-len(filepath.Ext(cidStr))]
		root, _ = cid.Parse(cidStr)
	}

	lsys := cidlink.DefaultLinkSystem()
	lsys.TrustedStorage = true
	lsys.SetReadStorage(storage)
	unixfsnode.AddUnixFSReificationToLinkSystem(&lsys)

	return lsys, root, nil
}
