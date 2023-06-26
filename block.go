package fixtureplate

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/bits"

	"github.com/ipfs/go-bitfield"
	"github.com/ipfs/go-cid"
	"github.com/ipfs/go-unixfsnode/data"
	"github.com/ipld/go-ipld-prime"
	"github.com/ipld/go-ipld-prime/codec/dagjson"
	"github.com/ipld/go-ipld-prime/node/basicnode"
)

type Block struct {
	Cid        cid.Cid   `json:"cid"`
	IpldPath   Path      `json:"ipldPath"`
	UnixfsPath Path      `json:"unixfsPath"`
	DataType   DataType  `json:"dataType"` // go-unixfsnode/data.Data_*
	Children   []Block   `json:"children,omitempty"`
	ByteOffset int64     `json:"byteOffset"`
	ByteSize   int64     `json:"byteSize"`
	Arity      int64     `json:"arity,omitempty"` // for sharded nodes
	Data       FieldData `json:"data,omitempty"`  // bitfield data for sharded nodes
}

type DataType int64

func (d DataType) MarshalJSON() ([]byte, error) {
	dataType := "RawLeaf"
	if dtn, has := data.DataTypeNames[int64(d)]; has {
		dataType = dtn
	}
	return json.Marshal(dataType)
}

func (d *DataType) UnmarshalJSON(byts []byte) error {
	var dataType string
	if err := json.Unmarshal(byts, &dataType); err != nil {
		return err
	}
	if dataType == "RawLeaf" {
		*d = DataType(-1)
	} else {
		dt, ok := data.DataTypeValues[dataType]
		if !ok {
			return fmt.Errorf("unknown data type: %s", dataType)
		}
		*d = DataType(dt)
	}
	return nil
}

func (d DataType) String() string {
	if dtn, has := data.DataTypeNames[int64(d)]; has {
		return dtn
	}
	return "RawLeaf"
}

func (d DataType) Int() int64 {
	return int64(d)
}

type FieldData []byte

func (fd FieldData) MarshalJSON() ([]byte, error) {
	if len(fd) == 0 {
		return json.Marshal(nil)
	}
	var buf bytes.Buffer
	dagjson.Encode(basicnode.NewBytes(fd), &buf)
	return buf.Bytes(), nil
}

func (fd *FieldData) UnmarshalJSON(byts []byte) error {
	if bytes.Equal(byts, []byte("null")) {
		return nil
	}
	var buf bytes.Buffer
	buf.Write(byts)
	nb := basicnode.Prototype.Bytes.NewBuilder()
	err := dagjson.Decode(nb, &buf)
	if err != nil {
		return err
	}
	*fd, err = nb.Build().AsBytes()
	return err
}

type Path ipld.Path

func (p Path) MarshalJSON() ([]byte, error) {
	return json.Marshal(p.String())
}

func (p *Path) UnmarshalJSON(data []byte) error {
	var pathStr string
	if err := json.Unmarshal(data, &pathStr); err != nil {
		return err
	}
	*p = Path(ipld.ParsePath(pathStr))
	return nil
}

func (p Path) String() string {
	return "/" + p.Path().String()
}

func (p Path) Path() ipld.Path {
	return ipld.Path(p)
}

func (b Block) Print(w io.Writer) {
	b.print("", w)
}

func (b Block) print(indent string, w io.Writer) {
	_b := b
	_b.Children = nil
	j, err := json.Marshal(_b)
	if err != nil {
		panic(err)
	}
	fmt.Fprintf(w, "%s%s\n", indent, string(j))
	for _, child := range b.Children {
		child.print(indent+"  ", w)
	}
}

func (b Block) Navigate(path ipld.Path, visitFn func(reason string, depth int, b Block)) error {
	visitFn("/", 0, b)

	progress := ipld.Path{}
	curr := b
	depth := 0

outer:
	for path.Len() > 0 {
		var nextSeg ipld.PathSegment
		nextSeg, path = path.Shift()
		progress = progress.AppendSegment(nextSeg)

		switch int64(curr.DataType) {
		case data.Data_Directory:
			for _, child := range curr.Children {
				if child.UnixfsPath.Path().String() == progress.String() {
					depth++
					visitFn("/"+nextSeg.String(), depth, child)
					curr = child
					continue outer
				}
			}
		case data.Data_HAMTShard:
			child, _depth, found, err := curr.findInHamt(nextSeg.String(), depth+1, visitFn)
			if err != nil {
				return err
			}
			if !found {
				errors.New("not found in HAMT")
			}
			depth = _depth
			visitFn("/"+nextSeg.String(), depth, child)
			curr = child
			continue outer
		default:
			return errors.New("unsupported")
		}

		return fmt.Errorf("segment not found: %s / %s", nextSeg.String(), path.String())
	}
	curr.visitAll("*", depth+1, visitFn)
	return nil
}

func (b Block) visitAll(reason string, depth int, visitFn func(reason string, depth int, b Block)) error {
	for _, child := range b.Children {
		visitFn(reason, depth, child)
		if err := child.visitAll(reason, depth+1, visitFn); err != nil {
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
	for {
		childIndex, err := hv.Next(log2)
		if err != nil {
			return Block{}, 0, false, err
		}
		if len(node.Data) == 0 {
			return Block{}, 0, false, errors.New("no field data for hamt node")
		}
		if node.Arity != b.Arity {
			return Block{}, 0, false, errors.New("inconsistent arity")
		}
		bf, err := bitfield.NewBitfield(int(node.Arity))
		if err != nil {
			return Block{}, 0, false, err
		}
		bf.SetBytes(node.Data)
		if !bf.Bit(childIndex) {
			return Block{}, depth, false, nil // not found in this hamt
		}
		linkIndex := bf.OnesBefore(childIndex)
		if linkIndex >= len(node.Children) || linkIndex < 0 {
			return Block{}, 0, false, errors.New("bad shard indexing")
		}
		child := node.Children[linkIndex]
		if child.DataType == DataType(data.Data_HAMTShard) {
			visitFn("<hamt>", depth, child)
			node = child
			depth++
		} else if child.UnixfsPath.Path().Last().String() == key {
			return child, depth, true, nil
		} else {
			return Block{}, 0, false, fmt.Errorf("unexpected hamt child, %s != %s", child.UnixfsPath.Path().Last().String(), key)
		}
	}
}
