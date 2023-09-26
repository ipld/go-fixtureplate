package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/ipfs/go-cid"
	"github.com/ipld/go-fixtureplate/block"
	"github.com/ipld/go-fixtureplate/car"
	"github.com/ipld/go-ipld-prime/datamodel"
	trustlessutils "github.com/ipld/go-trustless-utils"
	cli "github.com/urfave/cli/v2"
)

var explainCommand = &cli.Command{
	Name: "explain",
	Usage: "Execute a trustless query across a DAG inside a CAR file and show" +
		" the block traversal details",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "car",
			Usage: "CAR file to read from, if not supplied, the first unnamed argument will be used",
		},
		&cli.StringFlag{
			Name:  "root",
			Usage: "Override the root CID of the CAR file or query",
		},
		&cli.StringFlag{
			Name:        "path",
			Usage:       "Path to query, required unless --query is specified",
			DefaultText: "/",
		},
		&cli.StringFlag{
			Name:        "scope",
			Aliases:     []string{"dag-scope"},
			DefaultText: "all",
			Usage:       "Scope of the query, one of: all, entity, block",
		},
		&cli.StringFlag{
			Name:    "bytes",
			Aliases: []string{"entity-bytes"},
			Value:   "",
			Usage: "Byte range of the terminating entity if that entity is a" +
				" sharded file of the form `from:to`, where * is a valid `to`" +
				" value and negative `to` values are also valid",
		},
		&cli.BoolFlag{
			Name:        "duplicates",
			Aliases:     []string{"dups"},
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
		&cli.BoolFlag{
			Name:  "ignore-missing",
			Value: false,
			Usage: "Ignore missing blocks in the CAR. Useful for when you have a" +
				" partial CAR and want to do a full (path=/) listing to see what's in" +
				" it.",
		},
	},
	Action: explainAction,
}

func explainAction(c *cli.Context) error {
	carPath := c.String("car")
	if carPath == "" {
		if c.Args().Len() > 0 {
			carPath = c.Args().First()
		} else {
			return fmt.Errorf("no CAR file specified")
		}
	}

	path := datamodel.Path{}
	scope := trustlessutils.DagScopeAll
	duplicates := true
	var byteRange *trustlessutils.ByteRange
	var err error
	var root cid.Cid

	if c.IsSet("query") {
		if root, path, scope, duplicates, byteRange, err = block.ParseQuery(c.String("query")); err != nil {
			return err
		}
	}

	if c.IsSet("path") {
		path = datamodel.ParsePath(c.String("path"))
	}

	if c.IsSet("root") {
		root, err = cid.Parse(c.String("root"))
		if err != nil {
			return err
		}
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

	blk, carFile, err := loadCar(c.App.ErrWriter, root, carPath)
	if err != nil {
		return err
	}
	defer carFile.Close()

	fmt.Println(block.PrintableQuery(blk.Cid, path, scope, byteRange, duplicates))

	if byteRange != nil {
		if scope != trustlessutils.DagScopeEntity {
			fmt.Fprintf(c.App.ErrWriter, "WARNING: byte range specified, but scope is not entity, switching to entity scope\n")
			scope = trustlessutils.DagScopeEntity
		}
	} else {
		br, _ := trustlessutils.ParseByteRange("")
		byteRange = &br
	}

	return blk.Navigate(path, scope, *byteRange, c.Bool("ignore-missing"), block.WritingVisitor(c.App.Writer, duplicates, fullPath))
}

func loadCar(printWriter io.Writer, requestedRoot cid.Cid, carPath string) (block.Block, *os.File, error) {
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
	if requestedRoot != cid.Undef && requestedRoot != root {
		cr := root.String()
		if root == cid.Undef {
			cr = "none"
		}
		fmt.Fprintf(printWriter, "Requested root CID [%s] does not match CAR file root [%s], proceeding with request\n", requestedRoot.String(), cr)
		root = requestedRoot
	}
	if root == cid.Undef {
		return block.Block{}, nil, fmt.Errorf("no root CID specified and CAR file has no root CID")
	}
	blk, err := block.NewBlock(ls, root)
	if err != nil {
		return block.Block{}, nil, err
	}
	return blk, carFile, nil
}
