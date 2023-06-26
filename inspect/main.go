package main

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"

	"github.com/ipld/go-ipld-prime"

	fixtureplate "github.com/rvagg/go-trustless-fixtureplate"
)

var ErrNotDir = errors.New("not a directory")

func main() {
	fileName := os.Args[1]
	carPath, err := filepath.Abs(fileName)
	if err != nil {
		panic(err)
	}
	carFile, err := os.Open(carPath)
	if err != nil {
		panic(err)
	}
	defer carFile.Close()
	ls, root, err := fixtureplate.LinkSystem(carFile)
	if err != nil {
		panic(err)
	}
	block, err := fixtureplate.Iterate(ls, ipld.Path{}, ipld.Path{}, root)
	if err != nil {
		panic(err)
	}
	if err := json.NewEncoder(os.Stdout).Encode(block); err != nil {
		panic(err)
	}
}
