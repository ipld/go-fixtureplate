package main

import (
	"log"
	"os"

	cli "github.com/urfave/cli/v2"

	_ "github.com/ipld/go-ipld-prime/codec/dagcbor"
)

func main() {
	app := &cli.App{
		Name:  "fixtureplate",
		Usage: "Work with, and inspect IPLD DAGs",
		Commands: []*cli.Command{
			explainCommand,
			generateCommand,
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
