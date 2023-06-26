package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/ipfs/go-cid"
	"github.com/ipld/go-ipld-prime"
	"github.com/ipld/go-ipld-prime/datamodel"
	fixtureplate "github.com/rvagg/go-trustless-fixtureplate"
	cli "github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:  "execute",
		Usage: "Execute a trustless query across a DAG inside a CAR file",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "car",
				Usage:    "CAR file to read from",
				Required: true,
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
		},
		Action: action,
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func action(c *cli.Context) error {
	block, carFile, err := loadCar(c)
	if err != nil {
		return err
	}
	defer carFile.Close()
	path := ipld.ParsePath(c.String("path"))

	scope := fixtureplate.DagScopeAll
	if c.IsSet("scope") {
		scope, err = fixtureplate.AsScope(c.String("scope"))
		if err != nil {
			return err
		}
	}
	byteRange, err := fixtureplate.ParseByteRange(c.String("bytes"))
	if err != nil {
		return err
	}
	printQuery(block.Cid, path, scope, byteRange)

	var lastDepth int
	return block.Navigate(path, scope, byteRange, func(reason string, depth int, blk fixtureplate.Block) {
		depthPad := strings.Repeat("  ", depth)
		if depth > lastDepth {
			depthPad = depthPad[:len(depthPad)-2] + "↳ "
		}
		fo := ""
		if reason == "*" {
			fo = fmt.Sprintf(" (/%s", blk.UnixfsPath.Last().String())
		}
		if blk.ByteSize > 0 {
			fo += fmt.Sprintf("[%d:%d]", blk.ByteOffset, blk.ByteOffset+blk.ByteSize)
		}
		if reason == "*" {
			fo += ")"
		}
		fmt.Printf("%-10s | %-9s | %s%s%s\n", blk.Cid, blk.DataTypeString(), depthPad, reason, fo)
		lastDepth = depth
	})
}

func printQuery(c cid.Cid, path datamodel.Path, scope fixtureplate.DagScope, byteRange fixtureplate.ByteRange) {
	pp := path.String()
	if pp != "" {
		pp = "/" + pp
	}
	br := ""
	if !byteRange.IsDefault() {
		br = fmt.Sprintf("&entity-bytes=%s", byteRange.String())
	}
	fmt.Printf("/ipfs/%s%s?dag-scope=%s%s\n", c, pp, scope, br)
}

func loadCar(c *cli.Context) (fixtureplate.Block, *os.File, error) {
	fileName := c.String("car")
	carPath, err := filepath.Abs(fileName)
	if err != nil {
		return fixtureplate.Block{}, nil, err
	}
	carFile, err := os.Open(carPath)
	if err != nil {
		return fixtureplate.Block{}, nil, err
	}
	ls, root, err := fixtureplate.LinkSystem(carFile)
	if err != nil {
		return fixtureplate.Block{}, nil, err
	}
	block, err := fixtureplate.NewBlock(ls, root)
	if err != nil {
		return fixtureplate.Block{}, nil, err
	}
	return block, carFile, nil
}
