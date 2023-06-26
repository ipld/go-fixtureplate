package fixtureplate

import (
	"errors"
	"fmt"
	"math/bits"

	"github.com/ipfs/go-bitfield"
	"github.com/ipfs/go-unixfsnode/data"
	"github.com/ipld/go-ipld-prime/datamodel"
)

func (b Block) Navigate(path datamodel.Path, visitFn func(reason string, depth int, b Block)) error {
	visitFn("/", 0, b)

	progress := datamodel.Path{}
	var nextSeg datamodel.PathSegment
	curr := b
	depth := 0

outer:
	for path.Len() > 0 {
		nextSeg, path = path.Shift()
		progress = progress.AppendSegment(nextSeg)

		switch int64(curr.DataType) {
		case data.Data_Directory:
			for _, child := range curr.Children {
				if child.UnixfsPath.String() == progress.String() {
					blk, err := child.Block()
					if err != nil {
						return err
					}
					depth++
					visitFn("/"+nextSeg.String(), depth, blk)
					curr = blk
					continue outer
				}
			}
		case data.Data_HAMTShard:
			child, _depth, found, err := curr.findInHamt(nextSeg.String(), depth+1, visitFn)
			if err != nil {
				return err
			}
			if !found {
				return errors.New("not found in HAMT")
			}
			depth = _depth
			visitFn("/"+nextSeg.String(), depth, child)
			curr = child
			continue outer
		default:
			return errors.New("unsupported " + data.DataTypeNames[int64(curr.DataType)])
		}

		return fmt.Errorf("segment not found in %s: %s / %s", data.DataTypeNames[curr.DataType], nextSeg.String(), path.String())
	}
	if curr.DataType == data.Data_File {
		fmt.Println("Visit all file")
		return curr.visitAll("/"+nextSeg.String(), depth+1, visitFn)
	}
	return curr.visitAll("*", depth+1, visitFn)
}

func (b Block) visitAll(reason string, depth int, visitFn func(reason string, depth int, b Block)) error {
	for _, child := range b.Children {
		blk, err := child.Block()
		if err != nil {
			return err
		}
		visitFn(reason, depth, blk)
		if err := blk.visitAll(reason, depth+1, visitFn); err != nil {
			return err
		}
	}
	return nil
}

func (b Block) findInHamt(key string, depth int, visitFn func(reason string, depth int, b Block)) (Block, int, bool, error) {
	if b.Arity <= 0 {
		return Block{}, 0, false, errors.New("no fanout (arity) for hamt node")
	}
	hv := &hashBits{b: hash([]byte(key))}
	log2 := bits.TrailingZeros(uint(b.Arity))
	node := b
	for { // descend into hamt
		childIndex, err := hv.Next(log2)
		if err != nil {
			return Block{}, 0, false, err
		}
		if len(node.FieldData) == 0 {
			return Block{}, 0, false, errors.New("no field data for hamt node")
		}
		if node.Arity != b.Arity {
			return Block{}, 0, false, errors.New("inconsistent arity")
		}
		bf, err := bitfield.NewBitfield(int(node.Arity))
		if err != nil {
			return Block{}, 0, false, err
		}
		bf.SetBytes(node.FieldData)
		if !bf.Bit(childIndex) {
			return Block{}, depth, false, nil // not found in this hamt
		}
		linkIndex := bf.OnesBefore(childIndex)
		if linkIndex >= len(node.Children) || linkIndex < 0 {
			return Block{}, 0, false, errors.New("bad shard indexing")
		}
		child := node.Children[linkIndex]
		blk, err := child.Block()
		if err != nil {
			return Block{}, 0, false, err
		}
		if blk.DataType == data.Data_HAMTShard {
			visitFn("<hamt>", depth, blk)
			node = blk
			depth++
		} else if child.UnixfsPath.Last().String() == key {
			return blk, depth, true, nil
		} else {
			return Block{}, 0, false, fmt.Errorf("unexpected hamt child, %s != %s", child.UnixfsPath.Last().String(), key)
		}
	}
}
