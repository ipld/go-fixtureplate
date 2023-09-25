package generator

import (
	"fmt"
	"io"
	"math/rand"
	"strings"

	"github.com/dustin/go-humanize"
	unixfstestutil "github.com/ipfs/go-unixfsnode/testutil"
	"github.com/ipld/go-ipld-prime/linking"
	trustlesstestutil "github.com/ipld/go-trustless-utils/testutil"
)

type Entity interface {
	Generate(lsys linking.LinkSystem, rnd *rand.Rand) (unixfstestutil.DirEntry, error)
	String() string
	Describe(indent string) string
}

var _ Entity = File{}
var _ Entity = Directory{}

type File struct {
	Name             string
	Size             uint64
	RandomSize       bool
	ZeroContent      bool
	Multiplier       int
	RandomMultiplier bool
}

func (f File) String() string {
	var sb strings.Builder
	if f.RandomMultiplier {
		sb.WriteRune('~')
	}
	if f.RandomMultiplier || f.Multiplier > 1 {
		sb.WriteString(fmt.Sprintf("%d*", f.Multiplier))
	}
	sb.WriteString("file:")
	if f.RandomSize {
		sb.WriteRune('~')
	}
	sb.WriteString(strings.ReplaceAll(humanize.Bytes(uint64(f.Size)), " ", ""))
	if f.ZeroContent {
		sb.WriteString("{zero}")
	}
	return sb.String()
}

func (f File) Describe(indent string) string {
	var sb strings.Builder
	if indent != "" {
		sb.WriteString(indent)
		sb.WriteString("→ ")
	}
	if f.RandomMultiplier {
		sb.WriteString("Approximately ")
		sb.WriteString(fmt.Sprintf("%d", f.Multiplier))
	} else {
		if f.Multiplier > 1 {
			sb.WriteString(fmt.Sprintf("%d", f.Multiplier))
		} else {
			sb.WriteString("A")
		}
	}
	sb.WriteString(" file")
	if f.Multiplier > 1 {
		sb.WriteRune('s')
	}
	if f.Name != "" {
		sb.WriteString(` named "`)
		sb.WriteString(f.Name)
		sb.WriteRune('"')
	}
	sb.WriteString(" of ")
	if f.RandomSize {
		sb.WriteString("approximately ")
	}
	if f.Size%1024 == 0 {
		sb.WriteString(humanize.IBytes(uint64(f.Size)))
	} else {
		sb.WriteString(humanize.Bytes(uint64(f.Size)))
	}
	if f.ZeroContent {
		sb.WriteString(" containing just zeros")
	}
	return sb.String()
}

// Generate _one_ of the files described by this descriptor. If there are
// multiple files described by this descriptor, call this function multiple
// times.
func (f File) Generate(lsys linking.LinkSystem, rand *rand.Rand) (unixfstestutil.DirEntry, error) {
	var rndReader io.Reader = rand
	if f.ZeroContent {
		rndReader = trustlesstestutil.ZeroReader{}
	}
	targetFileSize := int(f.Size)
	if f.RandomSize {
		for {
			targetFileSize = int(rand.NormFloat64()*float64(targetFileSize)/10.0 + float64(targetFileSize))
			if targetFileSize > 0 {
				break
			}
		}
	}
	return unixfstestutil.UnixFSFile(lsys, targetFileSize, unixfstestutil.WithRandReader(rndReader))
}

type DirType string

const (
	DirType_Sharded DirType = "sharded"
	DirType_Plain   DirType = "plain"
)

type Directory struct {
	Type             DirType
	ShardBitwidth    int
	Name             string
	Multiplier       int
	RandomMultiplier bool
	Children         []Entity
}

func (d Directory) String() string {
	var sb strings.Builder
	if d.RandomMultiplier {
		sb.WriteRune('~')
	}
	if d.RandomMultiplier || d.Multiplier > 1 {
		sb.WriteString(fmt.Sprintf("%d*", d.Multiplier))
	}
	sb.WriteString("dir")
	switch d.Type {
	case DirType_Sharded:
		sb.WriteString(fmt.Sprintf("{sharded:%d}", d.ShardBitwidth))
	}
	sb.WriteRune('(')
	for i, c := range d.Children {
		if i > 0 {
			sb.WriteString(",")
		}
		sb.WriteString(c.String())
	}
	sb.WriteString(")")
	return sb.String()
}

func (d Directory) Describe(indent string) string {
	var sb strings.Builder
	if indent != "" {
		sb.WriteString(indent)
		sb.WriteString("→ ")
	}
	if d.RandomMultiplier {
		sb.WriteString("Approximately ")
		sb.WriteString(fmt.Sprintf("%d", d.Multiplier))
	} else {
		if d.Multiplier > 1 {
			sb.WriteString(fmt.Sprintf("%d", d.Multiplier))
		} else {
			sb.WriteString("A")
		}
	}
	if d.Multiplier > 1 {
		sb.WriteString(" directories")
	} else {
		sb.WriteString(" directory")
	}
	if d.Name != "" {
		sb.WriteString(` named "`)
		sb.WriteString(d.Name)
		sb.WriteRune('"')
	}
	switch d.Type {
	case DirType_Sharded:
		sb.WriteString(fmt.Sprintf(" sharded with bitwidth %d", d.ShardBitwidth))
	}
	sb.WriteString(" containing:")
	for _, c := range d.Children {
		sb.WriteString("\n")
		sb.WriteString(c.Describe(indent + "  "))
	}
	return sb.String()
}

func (d Directory) Generate(lsys linking.LinkSystem, rand *rand.Rand) (unixfstestutil.DirEntry, error) {
	var sbw int
	if d.Type == DirType_Sharded {
		sbw = d.ShardBitwidth
	}
	children := make([]Entity, 0)
	for _, child := range d.Children {
		var multiplier int
		var rndMultiplier bool
		switch et := child.(type) {
		case File:
			multiplier = et.Multiplier
			rndMultiplier = et.RandomMultiplier
		case Directory:
			multiplier = et.Multiplier
			rndMultiplier = et.RandomMultiplier
		default:
			return unixfstestutil.DirEntry{}, fmt.Errorf("unknown entity type: %T", et)
		}
		if rndMultiplier {
			for {
				multiplier = int(rand.NormFloat64()*float64(multiplier)/10.0 + float64(multiplier))
				if multiplier >= 0 { // could be zero!
					break
				}
			}
		}
		for i := 0; i < multiplier; i++ {
			children = append(children, child)
		}
	}
	chidx := 0
	return unixfstestutil.UnixFSDirectory(
		lsys,
		0,
		unixfstestutil.WithRandReader(rand),
		unixfstestutil.WithShardBitwidth(sbw),
		unixfstestutil.WithChildGenerator(func(name string) (*unixfstestutil.DirEntry, error) {
			if chidx >= len(children) {
				return nil, nil
			}
			ch := children[chidx]
			chidx++
			de, err := ch.Generate(lsys, rand)
			if err != nil {
				return nil, err
			}
			de.Path = name
			return &de, nil
		}))
}
