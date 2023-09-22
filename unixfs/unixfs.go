package unixfs

import (
	"errors"
	"fmt"

	"github.com/ipfs/go-unixfsnode/data"
	dagpb "github.com/ipld/go-codec-dagpb"
	"github.com/ipld/go-ipld-prime"
	"github.com/spaolacci/murmur3"
)

func ToPbnode(node ipld.Node) (dagpb.PBNode, error) {
	pbb := dagpb.Type.PBNode.NewBuilder()
	if err := pbb.AssignNode(node); err != nil {
		return nil, err
	}
	return pbb.Build().(dagpb.PBNode), nil
}

func ToData(node dagpb.PBNode) (data.UnixFSData, error) {
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

/* --------- HAMT --------- */

func Hash(val []byte) []byte {
	h := murmur3.New64()
	h.Write(val)
	return h.Sum(nil)
}

// HashBits is a helper that allows the reading of the 'next n bits' as an integer.
type HashBits struct {
	Bits     []byte
	Consumed int
}

func mkmask(n int) byte {
	return (1 << uint(n)) - 1
}

// Next returns the next 'i' bits of the hashBits value as an integer, or an
// error if there aren't enough bits.
func (hb *HashBits) Next(i int) (int, error) {
	if hb.Consumed+i > len(hb.Bits)*8 {
		return 0, errors.New("hamt is too deep")
	}
	return hb.next(i), nil
}

func (hb *HashBits) next(i int) int {
	curbi := hb.Consumed / 8
	leftb := 8 - (hb.Consumed % 8)

	curb := hb.Bits[curbi]
	if i == leftb {
		out := int(mkmask(i) & curb)
		hb.Consumed += i
		return out
	}
	if i < leftb {
		a := curb & mkmask(leftb) // mask out the high bits we don't want
		b := a & ^mkmask(leftb-i) // mask out the low bits we don't want
		c := b >> uint(leftb-i)   // shift whats left down
		hb.Consumed += i
		return int(c)
	}
	out := int(mkmask(leftb) & curb)
	out <<= uint(i - leftb)
	hb.Consumed += leftb
	out += hb.next(i - leftb)
	return out

}
