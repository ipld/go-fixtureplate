package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ipld/go-ipld-prime"
	fixtureplate "github.com/rvagg/go-trustless-fixtureplate"
)

func main() {
	fileName := os.Args[1]
	var path string
	if len(os.Args) > 2 {
		path = os.Args[2]
	}

	var block fixtureplate.Block
	switch filepath.Ext(fileName) {
	case ".car":
		var err error
		block, err = blockFromCar(fileName)
		if err != nil {
			panic(err)
		}
	case ".json":
		var err error
		block, err = blockFromJson(fileName)
		if err != nil {
			panic(err)
		}
	default:
		panic("unknown file extension")
	}

	var lastDepth int
	if err := block.Navigate(ipld.ParsePath(path), func(reason string, depth int, blk fixtureplate.Block) {
		depthPad := strings.Repeat("  ", depth)
		if depth > lastDepth {
			depthPad = depthPad[:len(depthPad)-2] + "↳ "
		}
		lastDepth = depth
		fmt.Printf("%-10s | %-9s | %s%s\n", blk.Cid, blk.DataType, depthPad, reason)
	}); err != nil {
		panic(err)
	}
}

func blockFromCar(fileName string) (fixtureplate.Block, error) {
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
	return fixtureplate.Iterate(ls, ipld.Path{}, ipld.Path{}, root)
}

func blockFromJson(fileName string) (fixtureplate.Block, error) {
	var block fixtureplate.Block
	filePath, err := filepath.Abs(fileName)
	if err != nil {
		panic(err)
	}
	file, err := os.Open(filePath)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	if err := json.NewDecoder(file).Decode(&block); err != nil {
		panic(err)
	}
	return block, nil
}
