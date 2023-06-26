package fixtureplate

import (
	"errors"
	"fmt"
	"math/bits"

	"github.com/ipfs/go-unixfsnode/data"
	"github.com/spaolacci/murmur3"
)

func hash(val []byte) []byte {
	h := murmur3.New64()
	h.Write(val)
	return h.Sum(nil)
}

// hashBits is a helper that allows the reading of the 'next n bits' as an integer.
type hashBits struct {
	b        []byte
	consumed int
}

func mkmask(n int) byte {
	return (1 << uint(n)) - 1
}

// Next returns the next 'i' bits of the hashBits value as an integer, or an
// error if there aren't enough bits.
func (hb *hashBits) Next(i int) (int, error) {
	if hb.consumed+i > len(hb.b)*8 {
		return 0, errors.New("hamt is too deep")
	}
	return hb.next(i), nil
}

func (hb *hashBits) next(i int) int {
	curbi := hb.consumed / 8
	leftb := 8 - (hb.consumed % 8)

	curb := hb.b[curbi]
	if i == leftb {
		out := int(mkmask(i) & curb)
		hb.consumed += i
		return out
	}
	if i < leftb {
		a := curb & mkmask(leftb) // mask out the high bits we don't want
		b := a & ^mkmask(leftb-i) // mask out the low bits we don't want
		c := b >> uint(leftb-i)   // shift whats left down
		hb.consumed += i
		return int(c)
	}
	out := int(mkmask(leftb) & curb)
	out <<= uint(i - leftb)
	hb.consumed += leftb
	out += hb.next(i - leftb)
	return out

}

func log2Size(nd data.UnixFSData) int {
	return bits.TrailingZeros(uint(nd.FieldFanout().Must().Int()))
}

func maxPadLength(nd data.UnixFSData) int {
	return len(fmt.Sprintf("%X", nd.FieldFanout().Must().Int()-1))
}
