package fixtureplate

import (
	"fmt"

	"github.com/ipfs/go-cid"
	"github.com/ipfs/go-unixfsnode/data"
	"github.com/ipld/go-ipld-prime/datamodel"
	"github.com/ipld/go-ipld-prime/linking"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"
	"github.com/ipld/go-ipld-prime/node/basicnode"

	_ "github.com/ipld/go-ipld-prime/codec/raw"
)

type Block struct {
	ls linking.LinkSystem

	Cid        cid.Cid
	IpldPath   datamodel.Path
	UnixfsPath datamodel.Path
	DataType   int64 // go-unixfsnode/data.Data_*
	Children   []Child
	ByteOffset int64
	ByteSize   int64
	Arity      int64   // for sharded nodes
	FieldData  []byte  // bitfield data for sharded nodes
	BlockSizes []int64 // for sharded files
}

type Child struct {
	ls    linking.LinkSystem
	block *Block

	Cid        cid.Cid
	IpldPath   datamodel.Path
	UnixfsPath datamodel.Path
	ByteOffset int64
}

func (c *Child) Block() (Block, error) {
	if c.block != nil {
		return *c.block, nil
	}
	block, err := NewBlockWith(c.ls, c.Cid, c.IpldPath, c.UnixfsPath, c.ByteOffset)
	if err != nil {
		return Block{}, err
	}
	c.block = &block
	return block, nil
}

func NewBlock(ls linking.LinkSystem, c cid.Cid) (Block, error) {
	return NewBlockWith(ls, c, datamodel.Path{}, datamodel.Path{}, 0)
}

func NewBlockWith(
	ls linking.LinkSystem,
	c cid.Cid,
	ipldPath,
	unixfsPath datamodel.Path,
	byteOffset int64,
) (Block, error) {

	node, err := ls.Load(linking.LinkContext{}, cidlink.Link{Cid: c}, basicnode.Prototype.Any)
	if err != nil {
		return Block{}, err
	}

	var dt int64 = -1
	var byteSize int64
	var children []Child
	var fieldData []byte
	var arity int64
	var blockSizes []int64

	if node.Kind() == datamodel.Kind_Bytes {
		byt, err := node.AsBytes()
		if err != nil {
			return Block{}, err
		}
		byteSize = int64(len(byt))
	} else {
		pbNode, err := toPbnode(node)
		if err != nil {
			return Block{}, err
		}
		ufsData, err := toData(pbNode)
		if err != nil {
			return Block{}, err
		}
		dt = ufsData.DataType.Int()

		switch dt {
		case data.Data_Raw:
			if !pbNode.Data.Exists() {
				return Block{}, fmt.Errorf("raw block has no data")
			}
			byteSize = int64(len(pbNode.Data.Must().Bytes()))
		case data.Data_Directory:
			children = make([]Child, pbNode.Links.Length())
			for itr := pbNode.Links.Iterator(); !itr.Done(); {
				ii, v := itr.Next()
				if !v.Name.Exists() {
					return Block{}, fmt.Errorf("directory link has no name")
				}
				children[ii] = Child{
					ls:         ls,
					Cid:        v.Hash.Link().(cidlink.Link).Cid,
					IpldPath:   ipldPath.AppendSegmentString("Links").AppendSegmentInt(ii).AppendSegmentString("Hash"),
					UnixfsPath: unixfsPath.AppendSegmentString(v.Name.Must().String()),
				}
			}
		case data.Data_File:
			children = make([]Child, pbNode.Links.Length())
			blockSizes = make([]int64, ufsData.BlockSizes.Length())
			li := ufsData.BlockSizes.Iterator()
			for !li.Done() {
				ii, v := li.Next()
				blockSizes[ii] = v.Int()
			}
			var offset int64
			for itr := pbNode.Links.Iterator(); !itr.Done(); {
				ii, v := itr.Next()
				if !v.Name.Exists() {
					return Block{}, fmt.Errorf("directory link has no name")
				}
				children[ii] = Child{
					ls:         ls,
					Cid:        v.Hash.Link().(cidlink.Link).Cid,
					IpldPath:   ipldPath.AppendSegmentString("Links").AppendSegmentInt(ii).AppendSegmentString("Hash"),
					UnixfsPath: unixfsPath,
					ByteOffset: byteOffset + offset, // TODO: is this correct for nested file structures?
				}
				offset += blockSizes[ii]
				byteSize += blockSizes[ii]
			}
		case data.Data_HAMTShard:
			arity = ufsData.FieldFanout().Must().Int()
			fieldData = ufsData.FieldData().Must().Bytes()
			pfxLen := len(fmt.Sprintf("%X", arity-1))
			children = make([]Child, pbNode.Links.Length())
			for itr := pbNode.Links.Iterator(); !itr.Done(); {
				ii, v := itr.Next()
				if !v.Name.Exists() {
					return Block{}, fmt.Errorf("directory link has no name")
				}
				name := v.Name.Must().String()[pfxLen:]
				unixfsPath := unixfsPath
				if name != "" {
					unixfsPath = unixfsPath.AppendSegmentString(name)
				}
				children[ii] = Child{
					ls:         ls,
					Cid:        v.Hash.Link().(cidlink.Link).Cid,
					IpldPath:   ipldPath.AppendSegmentString("Links").AppendSegmentInt(ii).AppendSegmentString("Hash"),
					UnixfsPath: unixfsPath,
				}
			}
		case data.Data_Metadata:
			return Block{}, fmt.Errorf("metadata block not supported")
		case data.Data_Symlink:
			return Block{}, fmt.Errorf("metadata block not supported")
		default:
			return Block{}, fmt.Errorf("unknown data type: %d", ufsData.Type())
		}
	}

	return Block{
		ls:         ls,
		Cid:        c,
		DataType:   dt,
		IpldPath:   ipldPath,
		UnixfsPath: unixfsPath,
		ByteOffset: byteOffset,
		ByteSize:   byteSize,
		Children:   children,
		Arity:      arity,
		FieldData:  fieldData,
		BlockSizes: blockSizes,
	}, nil
}

func (b Block) DataTypeString() string {
	if dtn, has := data.DataTypeNames[b.DataType]; has {
		return dtn
	}
	return "RawLeaf"
}

func (b Block) Length() int64 {
	var l int64
	for _, c := range b.BlockSizes {
		l += c
	}
	return l
}
