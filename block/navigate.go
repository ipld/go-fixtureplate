package block

import (
	"errors"
	"fmt"
	"math"
	"math/bits"
	"net/url"
	"strings"

	"github.com/ipfs/go-bitfield"
	"github.com/ipfs/go-cid"
	"github.com/ipfs/go-unixfsnode/data"
	"github.com/ipld/go-fixtureplate/unixfs"
	"github.com/ipld/go-ipld-prime/datamodel"
	trustlessutils "github.com/ipld/go-trustless-utils"
	trustlesshttp "github.com/ipld/go-trustless-utils/http"
)

func (b Block) Navigate(
	path datamodel.Path,
	scope trustlessutils.DagScope,
	bytes trustlessutils.ByteRange,
	visitFn func(p datamodel.Path, depth int, b Block),
) error {
	visitFn(datamodel.Path{}, 0, b)

	progress := datamodel.Path{}
	curr := b
	depth := 0

outer:
	for path.Len() > 0 {
		var nextSeg datamodel.PathSegment
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
					visitFn(progress, depth, blk)
					curr = blk
					if scope == trustlessutils.DagScopeEntity && path.Len() == 0 {
						break outer
					}
					continue outer
				}
			}
		case data.Data_HAMTShard:
			child, _depth, found, err := curr.findInHamt(progress, depth+1, visitFn)
			if err != nil {
				return err
			}
			if !found {
				return errors.New("not found in HAMT")
			}
			depth = _depth
			visitFn(progress, depth, child)
			curr = child
			if scope == trustlessutils.DagScopeEntity && path.Len() == 0 {
				break outer
			}
			continue outer
		default:
			return errors.New("unsupported " + data.DataTypeNames[int64(curr.DataType)])
		}

		return fmt.Errorf("segment not found in %s: %s / %s", data.DataTypeNames[curr.DataType], nextSeg.String(), path.String())
	}

	switch scope {
	case trustlessutils.DagScopeBlock:
		return nil
	case trustlessutils.DagScopeEntity:
		return curr.visitAllEntity(progress, bytes, depth+1, visitFn)
	}

	return curr.visitAll(progress, depth+1, visitFn)
}

func (b Block) visitAll(p datamodel.Path, depth int, visitFn func(p datamodel.Path, depth int, b Block)) error {
	for _, child := range b.Children {
		blk, err := child.Block()
		if err != nil {
			return err
		}
		visitFn(p, depth, blk)
		if err := blk.visitAll(p, depth+1, visitFn); err != nil {
			return err
		}
	}
	return nil
}

func (b Block) visitAllFile(p datamodel.Path, bytes trustlessutils.ByteRange, depth int, visitFn func(p datamodel.Path, depth int, b Block)) error {
	from := bytes.From
	var to int64 = math.MaxInt64
	if bytes.To != nil {
		to = *bytes.To
		if to > 0 {
			to++ // selector is exclusive, so increment the end
		}
	}
	if from < 0 {
		from = b.Length() + from
		if from < 0 {
			from = 0
		}
	}
	if to < 0 {
		to = b.Length() + to
		if to < 0 {
			to = 0
		}
	}
	if from > to {
		br := &trustlessutils.ByteRange{From: from, To: &to}
		return fmt.Errorf("invalid range (len=%d) %s (orig=%s)", b.Length(), br.String(), bytes.String())
	}

	var visit func(b Block, depth int) error
	visit = func(b Block, depth int) error {
		if len(b.Children) > 0 && b.DataType != data.Data_File {
			return errors.New("expected file")
		}
		for ii, child := range b.Children {
			if child.ByteOffset+b.BlockSizes[ii]-1 < from {
				continue
			}
			if child.ByteOffset >= to {
				continue
			}
			blk, err := child.Block()
			if err != nil {
				return err
			}
			visitFn(p, depth, blk)
			if err := visit(blk, depth+1); err != nil {
				return err
			}
		}
		return nil
	}
	return visit(b, depth)
}

func (b Block) visitAllEntity(p datamodel.Path, bytes trustlessutils.ByteRange, depth int, visitFn func(p datamodel.Path, depth int, b Block)) error {
	if b.DataType == data.Data_File {
		return b.visitAllFile(p, bytes, depth, visitFn)
	}

	for _, child := range b.Children {
		blk, err := child.Block()
		if err != nil {
			return err
		}
		if blk.UnixfsPath.Last() != p.Last() {
			continue
		}
		visitFn(p, depth, blk)
		blk.visitAllEntity(p, bytes, depth+1, visitFn)
	}
	return nil
}

func (b Block) findInHamt(p datamodel.Path, depth int, visitFn func(p datamodel.Path, depth int, b Block)) (Block, int, bool, error) {
	if b.Arity <= 0 {
		return Block{}, 0, false, errors.New("no fanout (arity) for hamt node")
	}
	key := p.Last().String()
	hv := &unixfs.HashBits{Bits: unixfs.Hash([]byte(key))}
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
		if blk.UnixfsPath.String() == b.UnixfsPath.String() && blk.DataType == data.Data_HAMTShard {
			visitFn(p.Pop(), depth, blk)
			node = blk
			depth++
		} else if child.UnixfsPath.Last().String() == key {
			return blk, depth, true, nil
		} else {
			return Block{}, 0, false, fmt.Errorf("unexpected hamt child, %s != %s", child.UnixfsPath.Last().String(), key)
		}
	}
}

func ParseQuery(spec string) (
	root cid.Cid,
	path datamodel.Path,
	scope trustlessutils.DagScope,
	duplicates bool,
	byteRange *trustlessutils.ByteRange,
	err error,
) {
	specParts := strings.Split(spec, "?")
	spec = specParts[0]

	root, path, err = trustlesshttp.ParseUrlPath(spec)
	if err != nil {
		return root, path, scope, duplicates, byteRange, err
	}

	switch len(specParts) {
	case 1:
	case 2:
		query, err := url.ParseQuery(specParts[1])
		if err != nil {
			return root, path, scope, duplicates, byteRange, err
		}
		scope, err = trustlessutils.ParseDagScope(query.Get("dag-scope"))
		if err != nil {
			return root, path, scope, duplicates, byteRange, err
		}
		duplicates = query.Get("dups") != "n"
		if query.Get("entity-bytes") != "" {
			br, err := trustlessutils.ParseByteRange(query.Get("entity-bytes"))
			if err != nil {
				return root, path, scope, duplicates, byteRange, err
			}
			byteRange = &br
		}
	default:
		return root, path, scope, duplicates, byteRange, fmt.Errorf("invalid query: %s", spec)
	}

	return root, path, scope, duplicates, byteRange, nil
}
