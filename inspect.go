package fixtureplate

import (
	"fmt"

	"github.com/ipfs/go-cid"
	"github.com/ipfs/go-unixfsnode/data"
	dagpb "github.com/ipld/go-codec-dagpb"
	"github.com/ipld/go-ipld-prime"
	"github.com/ipld/go-ipld-prime/linking"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"
	"github.com/ipld/go-ipld-prime/node/basicnode"

	_ "github.com/ipld/go-ipld-prime/codec/raw"
)

func Iterate(ls linking.LinkSystem, ipldPath, unixfsPath ipld.Path, lnk cid.Cid) (Block, error) {
	node, err := ls.Load(ipld.LinkContext{}, cidlink.Link{Cid: lnk}, basicnode.Prototype.Any)
	if err != nil {
		return Block{}, err
	}

	if node.Kind() == ipld.Kind_Bytes {
		byt, err := node.AsBytes()
		if err != nil {
			return Block{}, err
		}
		return Block{
			Cid:        lnk,
			IpldPath:   Path(ipldPath),
			UnixfsPath: Path(unixfsPath),
			DataType:   -1,
			ByteOffset: 0,
			ByteSize:   int64(len(byt)),
		}, nil
	}

	pbNode, err := toPbnode(node)
	if err != nil {
		return Block{}, err
	}
	ufsData, err := toData(pbNode)
	if err != nil {
		return Block{}, err
	}

	switch ufsData.DataType.Int() {
	case data.Data_Raw:
		return IterateRawBlock(ipldPath, unixfsPath, lnk, pbNode)
	case data.Data_Directory:
		return IterateDirectoryBlock(ls, ipldPath, unixfsPath, lnk, pbNode)
	case data.Data_File:
		return IterateFileBlock(ls, ipldPath, unixfsPath, lnk, pbNode)
	case data.Data_Metadata:
		return Block{}, fmt.Errorf("metadata block not supported")
	case data.Data_Symlink:
		return Block{}, fmt.Errorf("metadata block not supported")
	case data.Data_HAMTShard:
		return IterateShardedDirectoryBlock(ls, ipldPath, unixfsPath, lnk, pbNode)
	default:
		return Block{}, fmt.Errorf("unknown data type: %d", ufsData.Type())
	}
}

func IterateRawBlock(ipldPath, unixfsPath ipld.Path, lnk cid.Cid, pbNode dagpb.PBNode) (Block, error) {
	if !pbNode.Data.Exists() {
		return Block{}, fmt.Errorf("raw block has no data")
	}
	return Block{
		Cid:        lnk,
		IpldPath:   Path(ipldPath),
		UnixfsPath: Path(unixfsPath),
		DataType:   DataType(data.Data_Raw),
		ByteOffset: 0,
		ByteSize:   int64(len(pbNode.Data.Must().Bytes())),
	}, nil
}

func IterateDirectoryBlock(ls linking.LinkSystem, ipldPath, unixfsPath ipld.Path, lnk cid.Cid, pbNode dagpb.PBNode) (Block, error) {
	children := make([]Block, pbNode.Links.Length())
	var totalSize int64
	for itr := pbNode.Links.Iterator(); !itr.Done(); {
		ii, v := itr.Next()
		ipldPath := ipldPath.AppendSegmentString("Links").AppendSegmentInt(ii).AppendSegmentString("Hash")
		if !v.Name.Exists() {
			return Block{}, fmt.Errorf("directory link has no name")
		}
		childLnk := v.Hash.Link().(cidlink.Link).Cid
		name := v.Name.Must().String()
		unixfsPath := unixfsPath.AppendSegmentString(name)
		var err error
		children[ii], err = Iterate(ls, ipldPath, unixfsPath, childLnk)
		if err != nil {
			return Block{}, err
		}
		totalSize += children[ii].ByteSize
	}
	return Block{
		Cid:        lnk,
		IpldPath:   Path(ipldPath),
		UnixfsPath: Path(unixfsPath),
		DataType:   DataType(data.Data_Directory),
		Children:   children,
		ByteOffset: 0,
		ByteSize:   totalSize,
	}, nil
}

func IterateFileBlock(ls linking.LinkSystem, ipldPath, unixfsPath ipld.Path, lnk cid.Cid, pbNode dagpb.PBNode) (Block, error) {
	ufsData, _ := toData(pbNode)
	li := ufsData.BlockSizes.Iterator()
	lengths := make([]int64, ufsData.BlockSizes.Length())
	for !li.Done() {
		ii, v := li.Next()
		lengths[ii] = v.Int()
	}
	children := make([]Block, pbNode.Links.Length())
	var offset int64
	var totalSize int64
	for itr := pbNode.Links.Iterator(); !itr.Done(); {
		ii, v := itr.Next()
		ipldPath := ipldPath.AppendSegmentString("Links").AppendSegmentInt(ii).AppendSegmentString("Hash")
		if !v.Name.Exists() {
			return Block{}, fmt.Errorf("directory link has no name")
		}
		childLnk := v.Hash.Link().(cidlink.Link).Cid
		var err error
		children[ii], err = Iterate(ls, ipldPath, unixfsPath, childLnk)
		if err != nil {
			return Block{}, err
		}
		if children[ii].DataType != -1 {
			return Block{}, fmt.Errorf("file link has non-raw child")
		}
		children[ii].ByteOffset = offset
		if children[ii].ByteSize != lengths[ii] {
			return Block{}, fmt.Errorf("file link has invalid length %d != %d", children[ii].ByteSize, lengths[ii])
		}
		offset += lengths[ii]
		totalSize += lengths[ii]
	}
	return Block{
		Cid:        lnk,
		IpldPath:   Path(ipldPath),
		UnixfsPath: Path(unixfsPath),
		DataType:   DataType(data.Data_File),
		Children:   children,
		ByteOffset: 0,
		ByteSize:   totalSize,
	}, nil
}

func IterateShardedDirectoryBlock(ls linking.LinkSystem, ipldPath, unixfsPath ipld.Path, lnk cid.Cid, pbNode dagpb.PBNode) (Block, error) {
	ufsData, _ := toData(pbNode)
	fanout := ufsData.FieldFanout().Must().Int()
	fieldData := ufsData.FieldData().Must().Bytes()
	pfxLen := len(fmt.Sprintf("%X", fanout-1))
	children := make([]Block, pbNode.Links.Length())
	var totalSize int64
	for itr := pbNode.Links.Iterator(); !itr.Done(); {
		ii, v := itr.Next()
		ipldPath := ipldPath.AppendSegmentString("Links").AppendSegmentInt(ii).AppendSegmentString("Hash")
		if !v.Name.Exists() {
			return Block{}, fmt.Errorf("directory link has no name")
		}
		childLnk := v.Hash.Link().(cidlink.Link).Cid
		name := v.Name.Must().String()[pfxLen:]
		unixfsPath := unixfsPath
		if name != "" {
			unixfsPath = unixfsPath.AppendSegmentString(name)
		}
		var err error
		children[ii], err = Iterate(ls, ipldPath, unixfsPath, childLnk)
		if err != nil {
			return Block{}, err
		}
		totalSize += children[ii].ByteSize
	}
	return Block{
		Cid:        lnk,
		IpldPath:   Path(ipldPath),
		UnixfsPath: Path(unixfsPath),
		DataType:   DataType(data.Data_HAMTShard),
		Children:   children,
		ByteOffset: 0,
		ByteSize:   totalSize,
		Arity:      fanout,
		Data:       fieldData,
	}, nil
}

func toPbnode(node ipld.Node) (dagpb.PBNode, error) {
	pbb := dagpb.Type.PBNode.NewBuilder()
	if err := pbb.AssignNode(node); err != nil {
		return nil, err
	}
	return pbb.Build().(dagpb.PBNode), nil
}

func toData(node dagpb.PBNode) (data.UnixFSData, error) {
	if !node.Data.Exists() {
		return nil, fmt.Errorf("no data")
	}
	ufsBytes := node.Data.Must().Bytes()
	ufsNode, err := data.DecodeUnixFSData(ufsBytes)
	if err != nil {
		return nil, err
	}
	return ufsNode, nil
}
