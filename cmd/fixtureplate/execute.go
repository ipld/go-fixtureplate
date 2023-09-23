package main

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/dustin/go-humanize"
	"github.com/ipfs/go-cid"
	"github.com/ipfs/go-unixfsnode/data"
	"github.com/ipld/go-fixtureplate/block"
	"github.com/ipld/go-fixtureplate/car"
	"github.com/ipld/go-ipld-prime/datamodel"
	trustlessutils "github.com/ipld/go-trustless-utils"
	trustlesshttp "github.com/ipld/go-trustless-utils/http"
	cli "github.com/urfave/cli/v2"
)

var executeCommand = &cli.Command{
	Name: "execute",
	Usage: "Execute a trustless query across a DAG inside a CAR file and show" +
		" the block traversal details",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "car",
			Usage: "CAR file to read from",
		},
		&cli.StringFlag{
			Name:        "path",
			Usage:       "Path to query, required unless --query is specified",
			DefaultText: "/",
		},
		&cli.StringFlag{
			Name:        "scope",
			DefaultText: "all",
			Usage:       "Scope of the query, one of: all, entity, block",
		},
		&cli.StringFlag{
			Name:  "bytes",
			Value: "",
			Usage: "Byte range of the terminating entity if that entity is a" +
				" sharded file of the form `from:to`, where * is a valid `to`" +
				" value and negative `to` values are also valid",
		},
		&cli.BoolFlag{
			Name:        "duplicates",
			DefaultText: "true",
			Usage:       "Include duplicate blocks in the output",
		},
		&cli.BoolFlag{
			Name:  "full-path",
			Value: true,
			Usage: "Print the full path of each block, not just the last path",
		},
		&cli.StringFlag{
			Name: "query",
			Usage: "Full query (e.g. /ipfs/bafy.../path?dag-scope=all&dups=n&byte-range=0:*)" +
				" (will be overridden by --path, --scope, --bytes and --duplicates," +
				" if set; note that this is not strictly a trustless query as it" +
				" incorporates elements, such as 'dups', that are normally included" +
				" in the Accept header)",
		},
	},
	Action: executeAction,
}

func executeAction(c *cli.Context) error {
	if !c.IsSet("car") {
		return fmt.Errorf("no CAR file specified")
	}

	path := datamodel.EmptyPath
	scope := trustlessutils.DagScopeAll
	duplicates := true
	var byteRange *trustlessutils.ByteRange
	var err error
	var root cid.Cid

	if c.IsSet("query") {
		spec := c.String("query")
		specParts := strings.Split(spec, "?")
		spec = specParts[0]

		root, path, err = trustlesshttp.ParseUrlPath(spec)
		if err != nil {
			return err
		}

		switch len(specParts) {
		case 1:
		case 2:
			query, err := url.ParseQuery(specParts[1])
			if err != nil {
				return err
			}
			scope, err = trustlessutils.ParseDagScope(query.Get("dag-scope"))
			if err != nil {
				return err
			}
			duplicates = query.Get("dups") != "n"
			if query.Get("entity-bytes") != "" {
				br, err := trustlessutils.ParseByteRange(query.Get("entity-bytes"))
				if err != nil {
					return err
				}
				byteRange = &br
			}
		default:
			return fmt.Errorf("invalid query: %s", spec)
		}
	}

	if c.IsSet("path") {
		path = datamodel.ParsePath(c.String("path"))
	}

	if c.IsSet("scope") {
		scope, err = trustlessutils.ParseDagScope(c.String("scope"))
		if err != nil {
			return err
		}
	}

	if c.IsSet("bytes") {
		br, err := trustlessutils.ParseByteRange(c.String("bytes"))
		if err != nil {
			return err
		}
		byteRange = &br
	}

	if c.IsSet("duplicates") {
		duplicates = c.Bool("duplicates")
	}

	fullPath := c.Bool("full-path")

	blk, carFile, err := loadCar(c.String("car"))
	if err != nil {
		return err
	}
	defer carFile.Close()

	if root != cid.Undef && root != blk.Cid {
		// TODO: allow override of root CID?
		return fmt.Errorf("root CID [%s] does not match CAR file root [%s]", root, blk.Cid)
	}

	printQuery(blk.Cid, path, scope, byteRange, duplicates)

	if byteRange != nil {
		if scope != trustlessutils.DagScopeEntity {
			fmt.Fprintf(c.App.ErrWriter, "WARNING: byte range specified, but scope is not entity, switching to entity scope\n")
			scope = trustlessutils.DagScopeEntity
		}
	} else {
		br, _ := trustlessutils.ParseByteRange("")
		byteRange = &br
	}

	return blk.Navigate(path, scope, *byteRange, visitor(duplicates, fullPath))
}

func visitor(duplicates, fullPath bool) func(p datamodel.Path, depth int, blk block.Block) {
	var lastDepth int
	seen := make(map[cid.Cid]struct{}, 0)

	return func(p datamodel.Path, depth int, blk block.Block) {
		if _, ok := seen[blk.Cid]; !duplicates && ok {
			return
		}
		seen[blk.Cid] = struct{}{}

		depthPad := strings.Repeat("  ", depth)
		if depth > lastDepth {
			depthPad = depthPad[:len(depthPad)-2] + "↳ "
		}
		fo := ""
		if fullPath {
			fo = fmt.Sprintf("/%s", blk.UnixfsPath.String())
		} else {
			fo = fmt.Sprintf("/%s", blk.UnixfsPath.Last().String())
		}
		if blk.ByteSize > 0 {
			fo += fmt.Sprintf(" [%d:%d] (%s B)", blk.ByteOffset, blk.ByteOffset+blk.ByteSize-1, humanize.Comma(blk.ByteSize))
		} else if blk.DataType == data.Data_HAMTShard && blk.ShardIndex != "" {
			fo += fmt.Sprintf(" [%s]", blk.ShardIndex)
		}
		fmt.Printf("%-10s | %-9s | %s%s\n", blk.Cid, blk.DataTypeString(), depthPad, fo)
		lastDepth = depth
	}
}

func printQuery(c cid.Cid, path datamodel.Path, scope trustlessutils.DagScope, byteRange *trustlessutils.ByteRange, duplicates bool) {
	pp := path.String()
	if pp != "" {
		pp = "/" + pp
	}
	br := ""
	if byteRange != nil {
		br = fmt.Sprintf("&entity-bytes=%s", byteRange.String())
	}
	dup := ""
	if !duplicates {
		dup = "&dups=n"
	}
	fmt.Printf("/ipfs/%s%s?dag-scope=%s%s%s\n", c, pp, scope, br, dup)
}

func loadCar(carPath string) (block.Block, *os.File, error) {
	var err error
	carPath, err = filepath.Abs(carPath)
	if err != nil {
		return block.Block{}, nil, err
	}
	carFile, err := os.Open(carPath)
	if err != nil {
		return block.Block{}, nil, err
	}
	ls, root, err := car.LinkSystem(carFile)
	if err != nil {
		return block.Block{}, nil, err
	}
	blk, err := block.NewBlock(ls, root)
	if err != nil {
		return block.Block{}, nil, err
	}
	return blk, carFile, nil
}
