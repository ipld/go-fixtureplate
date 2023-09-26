package main

import (
	"fmt"
	"math/rand"
	"os"
	"strings"

	"github.com/ipfs/go-cid"
	"github.com/ipld/go-car/v2"
	storagecar "github.com/ipld/go-car/v2/storage"
	"github.com/ipld/go-fixtureplate/generator"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"
	selectorparse "github.com/ipld/go-ipld-prime/traversal/selector/parse"
	cli "github.com/urfave/cli/v2"
)

var generateCommand = &cli.Command{
	Name:  "generate",
	Usage: "Generate a synthetic UnixFS DAG for use in testing",
	Flags: []cli.Flag{
		&cli.Int64Flag{
			Name:  "seed",
			Usage: "Seed for the random number generator",
		},
	},
	ArgsUsage: "<spec>",
	Action:    generateAction,
}

func generateAction(c *cli.Context) error {
	spec := c.Args().First()
	if spec == "" {
		// "help" becomes a subcommand, clear it to deal with a urfave/cli bug
		// Ref: https://github.com/urfave/cli/blob/v2.25.7/help.go#L253-L255
		c.Command.Subcommands = nil
		cli.ShowCommandHelpAndExit(c, "generate", 0)
		return nil
	}
	entity, err := generator.Parse(spec)
	if err != nil {
		if err, ok := err.(generator.ErrParse); ok {
			// move in enough spaces to point to err.Pos on the line above
			fmt.Printf("Input spec: %s\n", spec)
			fmt.Printf("            %s^\n", strings.Repeat(" ", err.Pos))
		}
		return err
	}
	fmt.Println(entity.Describe(""))

	outf, err := os.CreateTemp("", "fixtureplate-*.car")
	if err != nil {
		return err
	}
	defer func() {
		outf.Close()
		os.Remove(outf.Name())
	}()

	storage, err := storagecar.NewReadableWritable(outf, []cid.Cid{}, car.WriteAsCarV1(true))
	if err != nil {
		return err
	}
	lsys := cidlink.DefaultLinkSystem()
	lsys.TrustedStorage = true
	lsys.SetReadStorage(storage)
	lsys.SetWriteStorage(storage)

	seed := c.Int64("seed")
	rand := rand.New(rand.NewSource(seed))

	rootEnt, err := entity.Generate(lsys, rand)
	if err != nil {
		return err
	}

	outFile := rootEnt.Root.String() + ".car"
	if err := car.TraverseToFile(
		c.Context,
		&lsys,
		rootEnt.Root,
		selectorparse.CommonSelector_ExploreAllRecursively,
		outFile,
		car.WriteAsCarV1(true),
	); err != nil {
		return err
	}

	fmt.Println("Wrote to", outFile)

	return nil
}
