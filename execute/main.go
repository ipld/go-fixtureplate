package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ipld/go-ipld-prime"
	fixtureplate "github.com/rvagg/go-trustless-fixtureplate"
)

func main() {
	block, carFile := loadCar()
	defer carFile.Close()
	var path string
	if len(os.Args) > 2 {
		path = os.Args[2]
	}

	var lastDepth int
	if err := block.Navigate(ipld.ParsePath(path), func(reason string, depth int, blk fixtureplate.Block) {
		depthPad := strings.Repeat("  ", depth)
		if depth > lastDepth {
			depthPad = depthPad[:len(depthPad)-2] + "↳ "
		}
		fo := ""
		if reason == "*" {
			fo = fmt.Sprintf(" (/%s", blk.UnixfsPath.Last().String())
		}
		if blk.ByteSize > 0 {
			fo += fmt.Sprintf("[%d:%d]", blk.ByteOffset, blk.ByteSize)
		}
		if reason == "*" {
			fo += ")"
		}
		fmt.Printf("%-10s | %-9s | %s%s%s\n", blk.Cid, blk.DataTypeString(), depthPad, reason, fo)
		lastDepth = depth
	}); err != nil {
		panic(err)
	}
}

func loadCar() (fixtureplate.Block, *os.File) {
	fileName := os.Args[1]
	carPath, err := filepath.Abs(fileName)
	if err != nil {
		panic(err)
	}
	carFile, err := os.Open(carPath)
	if err != nil {
		panic(err)
	}
	ls, root, err := fixtureplate.LinkSystem(carFile)
	if err != nil {
		panic(err)
	}
	block, err := fixtureplate.NewBlock(ls, root)
	if err != nil {
		panic(err)
	}
	return block, carFile
}
