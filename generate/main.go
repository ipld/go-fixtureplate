package main

import (
	"fmt"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"testing"

	"github.com/ipfs/go-cid"
	unixfs "github.com/ipfs/go-unixfsnode/testutil"
	"github.com/ipfs/go-unixfsnode/testutil/namegen"
	car "github.com/ipld/go-car/v2"
	storagecar "github.com/ipld/go-car/v2/storage"
	"github.com/ipld/go-ipld-prime/linking"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"
	"github.com/stretchr/testify/require"
)

func main() {
	t := &testing.T{}

	// absolute path to carPath relative to this file
	outf, err := os.Create(".tmp.car")
	if err != nil {
		panic(err)
	}
	defer outf.Close()
	storage, err := storagecar.NewReadableWritable(outf, []cid.Cid{}, car.WriteAsCarV1(true))
	if err != nil {
		panic(err)
	}
	lsys := cidlink.DefaultLinkSystem()
	lsys.TrustedStorage = true
	lsys.SetReadStorage(storage)
	lsys.SetWriteStorage(storage)
	var seed int64 = 0
	rand := rand.New(rand.NewSource(seed))
	t.Log("Seed", seed)
	/*
		- file
		- file
		- dir <-- test entity
			- sm file
			- lg file <-- test entity
			...
		- sharded dir <-- test entity, all
			- sm file <-- test entity, all, block
			- lg file <-- test entity, all, block
			...
		- dir
			- file
			- file
			- dir <-- test entity
				-file
				-file
				...
			...
		- sharded dir
			- file
			- file
			- dir <-- test entity
				-file
				-file <-- test entity
				...
			...
	*/
	toFileName := func(name string) string {
		ext, err := namegen.RandomFileExtension(rand)
		require.NoError(t, err)
		return name + ext
	}
	i := 0
	dir := unixfs.GenerateDirectoryFor(t, &lsys, rand, "", false, func(name string) *unixfs.DirEntry {
		i++
		switch {
		case i <= 10:
			// 10 small files
			return mkFile(t, &lsys, rand, toFileName(name), 1000)
		case i <= 20:
			// 10 large files
			return mkFile(t, &lsys, rand, toFileName(name), 2000000)
		case i == 21 || i == 22:
			// 1 plain and 1 sharded dir of 40 files
			j := 0
			sharded := i == 22
			de := unixfs.GenerateDirectoryFor(t, &lsys, rand, name, sharded, func(_name string) *unixfs.DirEntry {
				j++
				_name = toFileName(_name)
				sz := 1000
				if j > 40 {
					return nil
				} else if j == 40 {
					sz = 2000000
					fmt.Printf("Large file, sharded=%v, name=%s\n", sharded, _name)
				}
				return mkFile(t, &lsys, rand, _name, sz)
			})
			fmt.Printf("Dir of files, sharded=%v, path=%s: %s\n", sharded, de.Path, de.Root.String())
			return &de
		case i == 23 || i == 24:
			// 1 plain and one sharded dir of 40 dirs of 10 files
			j := 0
			sharded := i == 24
			de := unixfs.GenerateDirectoryFor(t, &lsys, rand, name, sharded, func(_name string) *unixfs.DirEntry {
				j++
				if j > 40 {
					return nil
				}
				k := 0
				de := unixfs.GenerateDirectoryFor(t, &lsys, rand, _name, false, func(__name string) *unixfs.DirEntry {
					k++
					__name = toFileName(__name)
					sz := 1000
					if k > 10 {
						return nil
					} else if j == 40 && k == 10 {
						fmt.Printf("Large file nested, sharded=%v, name=%s\n", sharded, __name)
						sz = 2000000
					}
					return mkFile(t, &lsys, rand, __name, sz)
				})
				return &de
			})
			fmt.Printf("Dir of dirs, sharded=%v, path=%s: %s\n", sharded, de.Path, de.Root.String())
			return &de
		default:
			return nil
		}
	})

	/*
		fmt.Println()
		var ls func(de DirEntry)
		ls = func(de DirEntry) {
			fmt.Println(de.Path, de.TSize, len(de.Children), len(de.SelfCids), de.Root.String())
			for _, c := range de.Children {
				ls(c)
			}
		}
		ls(dir)
	*/
	fmt.Println("Root", dir.Root.String())
	outf.Close()
	// move .tmp.car to cwd/{root}.car
	carPath := filepath.Join(".", dir.Root.String()+".car")
	if err = os.Rename(".tmp.car", carPath); err != nil {
		panic(err)
	}
	fmt.Println("Wrote to", carPath)
}

func mkFile(t *testing.T, lsys *linking.LinkSystem, rand *rand.Rand, name string, szRange int) *unixfs.DirEntry {
	var sz int
	for sz <= 0 {
		sz = int(math.Abs(rand.NormFloat64() * float64(szRange)))
	}
	dirEnt := unixfs.GenerateFile(t, lsys, rand, sz)
	dirEnt.Path = name
	return &dirEnt
}
