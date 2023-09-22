package main

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/dustin/go-humanize"
	"github.com/ipfs/go-cid"
	"github.com/ipld/go-fixtureplate/block"
	"github.com/ipld/go-fixtureplate/car"
	"github.com/ipld/go-ipld-prime"
	"github.com/ipld/go-ipld-prime/datamodel"
	trustlessutils "github.com/ipld/go-trustless-utils"
	cli "github.com/urfave/cli/v2"
)

var executeCommand = &cli.Command{
	Name:  "execute",
	Usage: "Execute a trustless query across a DAG inside a CAR file and show the block traversal details",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "car",
			Usage: "CAR file to read from",
		},
		&cli.StringFlag{
			Name:        "path",
			Usage:       "Path to query",
			DefaultText: "",
		},
		&cli.StringFlag{
			Name:        "scope",
			DefaultText: "all",
			Usage:       "Scope of the query, one of: all, entity, block",
		},
		&cli.StringFlag{
			Name:        "bytes",
			DefaultText: "",
			Usage: "Byte range of the terminating entity if that entity is a" +
				" sharded file of the form `from:to`, where * is a valid `to`" +
				" value and negative `to` values are also valid",
		},
		&cli.BoolFlag{
			Name:        "duplicates",
			DefaultText: "false",
			Usage:       "Include duplicate blocks in the output",
		},
		&cli.BoolFlag{
			Name:        "full-path",
			DefaultText: "false",
			Usage:       "Print the full path of each '*' block, not just the last path",
		},
		&cli.StringFlag{
			Name: "query",
			Usage: "Full query (e.g. /ipfs/bafy.../path?dag-scope=all&dups=n&byte-range=0:*)" +
				" (overrides --path, --scope, --bytes)",
		},
	},
	Action: executeAction,
}

func executeAction(c *cli.Context) error {
	blk, carFile, err := loadCar(c)
	if err != nil {
		return err
	}
	defer carFile.Close()

	var path ipld.Path
	scope := trustlessutils.DagScopeAll
	var byteRange *trustlessutils.ByteRange
	var duplicates bool

	if c.IsSet("query") {
		spec := c.String("query")
		specParts := strings.Split(spec, "?")
		if len(specParts) != 2 {
			panic("invalid spec")
		}
		spec = specParts[0]
		query, err := url.ParseQuery(specParts[1])
		if err != nil {
			panic(err)
		}
		specParts = strings.Split(spec, "/")
		if specParts[0] != "" && specParts[1] != "ipfs" {
			panic("invalid spec")
		}
		root, err := cid.Parse(specParts[2])
		if err != nil {
			panic(err)
		}
		if root != blk.Cid {
			// TODO: allow override of root CID?
			panic("root CID does not match CAR file")
		}
		path = ipld.ParsePath(strings.Join(specParts[3:], "/"))
		scope, err = trustlessutils.ParseDagScope(query.Get("dag-scope"))
		if err != nil {
			panic(err)
		}
		duplicates = query.Get("dups") == "y"
		if query.Get("entity-bytes") != "" {
			br, err := trustlessutils.ParseByteRange(query.Get("entity-bytes"))
			if err != nil {
				panic(err)
			}
			byteRange = &br
		}
	} else {
		if !c.IsSet("path") {
			return fmt.Errorf("must specify --path or --query")
		}
		path = ipld.ParsePath(c.String("path"))
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
	}

	if duplicates {
		panic("TODO: support dups=y")
	}

	printQuery(blk.Cid, path, scope, byteRange, duplicates)
	fullPath := c.Bool("full-path")

	if byteRange == nil {
		br, _ := trustlessutils.ParseByteRange("")
		byteRange = &br
	}
	var lastDepth int
	return blk.Navigate(path, scope, *byteRange, func(reason string, depth int, blk block.Block) {
		depthPad := strings.Repeat("  ", depth)
		if depth > lastDepth {
			depthPad = depthPad[:len(depthPad)-2] + "↳ "
		}
		fo := ""
		if reason == "*" {
			if fullPath {
				fo = fmt.Sprintf(" (/%s", blk.UnixfsPath.String())
			} else {
				fo = fmt.Sprintf(" (/%s", blk.UnixfsPath.Last().String())
			}
		}
		if blk.ByteSize > 0 {
			fo += fmt.Sprintf("[%d:%d] (%s B)", blk.ByteOffset, blk.ByteOffset+blk.ByteSize-1, humanize.Comma(blk.ByteSize))
		}
		if reason == "*" {
			fo += ")"
		}
		fmt.Printf("%-10s | %-9s | %s%s%s\n", blk.Cid, blk.DataTypeString(), depthPad, reason, fo)
		lastDepth = depth
	})
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
	if duplicates {
		dup = "&dups=y"
	}
	fmt.Printf("/ipfs/%s%s?dag-scope=%s%s%s\n", c, pp, scope, br, dup)
}

func loadCar(c *cli.Context) (block.Block, *os.File, error) {
	fileName := c.String("car")
	carPath, err := filepath.Abs(fileName)
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
