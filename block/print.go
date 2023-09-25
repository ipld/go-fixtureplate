package block

import (
	"fmt"
	"io"
	"strings"

	"github.com/dustin/go-humanize"
	"github.com/ipfs/go-cid"
	"github.com/ipfs/go-unixfsnode/data"
	"github.com/ipld/go-ipld-prime/datamodel"
	trustlessutils "github.com/ipld/go-trustless-utils"
)

func WritingVisitor(w io.Writer, duplicates, fullPath bool) func(p datamodel.Path, depth int, blk Block) {
	var lastDepth int
	seen := make(map[cid.Cid]struct{}, 0)

	return func(p datamodel.Path, depth int, blk Block) {
		if _, ok := seen[blk.Cid]; !duplicates && ok {
			return
		}
		seen[blk.Cid] = struct{}{}

		depthPad := strings.Repeat("  ", depth)
		if depth > lastDepth {
			depthPad = depthPad[:len(depthPad)-2] + "↳ "
		}
		fo := ""
		if fullPath {
			fo = fmt.Sprintf("/%s", blk.UnixfsPath.String())
		} else {
			fo = fmt.Sprintf("/%s", blk.UnixfsPath.Last().String())
		}
		if blk.ByteSize > 0 {
			fo += fmt.Sprintf(" [%d:%d] (%s B)", blk.ByteOffset, blk.ByteOffset+blk.ByteSize-1, humanize.Comma(blk.ByteSize))
		} else if blk.DataType == data.Data_HAMTShard && blk.ShardIndex != "" {
			fo += fmt.Sprintf(" [%s]", blk.ShardIndex)
		}
		fmt.Fprintf(w, "%-10s | %-9s | %s%s\n", blk.Cid, blk.DataTypeString(), depthPad, fo)
		lastDepth = depth
	}
}

func PrintableQuery(c cid.Cid, path datamodel.Path, scope trustlessutils.DagScope, byteRange *trustlessutils.ByteRange, duplicates bool) string {
	pp := path.String()
	if pp != "" {
		pp = "/" + pp
	}
	br := ""
	if byteRange != nil {
		br = fmt.Sprintf("&entity-bytes=%s", byteRange.String())
	}
	dup := ""
	if !duplicates {
		dup = "&dups=n"
	}
	return fmt.Sprintf("/ipfs/%s%s?dag-scope=%s%s%s", c, pp, scope, br, dup)
}
