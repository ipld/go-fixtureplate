package generator

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"unicode"

	"github.com/dustin/go-humanize"
)

type ErrParse struct {
	Pos int
	Err error
}

func (e ErrParse) Error() string {
	return fmt.Sprintf("parse error at position %d: %s", e.Pos, e.Err)
}

// directory that contains ~5 files of alternating sizes, 2 directories, each of
// which contain ~10 files of the same size, and 1 more file with a large size
// but all zeros
//
// dir((~5*(file:1KB,file:100KB)),(2*dir(~10*file:50KB)),(file:1GB{zero:true}))
//
// broken down as:
//
// (~5*(file:1KB,file:100KB)): Around 5 files alternating between 1KB and 100KB sizes.
// (2*dir(~10*file:50KB)): 2 directories, each containing around 10 files of 50KB.
// (file:1GB{zero:true}): A single file with a large size (1GB) and all zeros.

func Parse(str string) (Entity, error) {
	p := &parser{str: str}
	e, err := p.parseEntity()
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(p.str[p.pos:]) != "" {
		return nil, errors.New("unexpected trailing characters")
	}
	if e.GetMultiplier() != 1 || e.IsRandomMultiplier() {
		return nil, errors.New("root entity must be strictly signular")
	}
	if e.GetName() != "" {
		return nil, errors.New("root entity can't be named")
	}
	return e, nil
}

type parser struct {
	str string
	pos int
}

func (p *parser) newParseError(msg string, a ...any) error {
	return ErrParse{
		Pos: p.pos,
		Err: fmt.Errorf(msg, a...),
	}
}

func (p *parser) nextChar(ch rune) (bool, error) {
	if !p.hasMore() {
		return false, p.newParseError("unexpected end")
	}
	return rune(p.str[p.pos]) == ch, nil
}

func (p *parser) hasMore() bool {
	return p.pos < len(p.str)
}

func (p *parser) parseEntity() (Entity, error) {
	rnd, err := p.slurpRandom()
	if err != nil {
		return nil, err
	}
	multiplier, err := p.slurpMultiplier()
	if err != nil {
		return nil, err
	}
	typ, err := p.slurpType()
	if err != nil {
		return nil, err
	}
	var entity Entity
	switch typ {
	case "file":
		entity, err = p.parseFile(multiplier, rnd)
	case "dir":
		entity, err = p.parseDir(multiplier, rnd)
	}
	if err != nil {
		return nil, err
	}
	return entity, nil
}

func (p *parser) parseFile(multiplier int, rnd bool) (Entity, error) {
	// must be followed by human readable size
	size, rndSize, err := p.slurpSize()
	if err != nil {
		return nil, err
	}
	name, zero, err := p.slurpFileOptions()
	if err != nil {
		return nil, err
	}
	if name != "" && (multiplier > 1 || rnd) {
		return nil, p.newParseError("file with a multiplier can't be named")
	}
	return File{
		Name:             name,
		Multiplier:       multiplier,
		RandomMultiplier: rnd,
		Size:             size,
		RandomSize:       rndSize,
		ZeroContent:      zero,
	}, nil
}

func (p *parser) parseDir(multiplier int, rnd bool) (Entity, error) {
	name, shardBitwidth, err := p.slurpDirOptions()
	if err != nil {
		return nil, err
	}
	if name != "" && (multiplier > 1 || rnd) {
		return nil, p.newParseError("directory with a multiplier can't be named")
	}
	if err := p.slurpOpen(); err != nil {
		return nil, err
	}
	typ := DirType_Plain
	if shardBitwidth > 0 {
		typ = DirType_Sharded
	}
	dir := Directory{
		Type:             typ,
		Name:             name,
		ShardBitwidth:    shardBitwidth,
		Multiplier:       multiplier,
		RandomMultiplier: rnd,
		Children:         []Entity{},
	}
	for {
		entity, err := p.parseEntity()
		if err != nil {
			return nil, err
		}
		dir.Children = append(dir.Children, entity)
		if comma, err := p.slurpComma(); err != nil {
			return nil, err
		} else if !comma {
			break
		}
	}
	if err := p.slurpClose(); err != nil {
		return nil, err
	}
	return dir, nil
}

// slurpFileOptions looks for an optional {} block which may optionally contain
// `zero` and `name:"foo"`, comma separated. Returns name adn zero.
func (p *parser) slurpFileOptions() (name string, zero bool, err error) {
	if !p.hasMore() {
		return "", false, nil
	}
	if ok, err := p.nextChar('{'); err != nil {
		return "", false, err
	} else if !ok {
		return "", false, nil
	}
	p.pos++
	if !p.hasMore() {
		return "", false, p.newParseError("unexpected end")
	}
	var vc int
	for p.hasMore() {
		if ok, err := p.nextChar('}'); err != nil {
			return "", false, err
		} else if ok {
			p.pos++
			break
		}
		if vc > 0 {
			if ok, err := p.nextChar(','); err != nil {
				return "", false, err
			} else if !ok {
				return "", false, p.newParseError("expected ','")
			}
			p.pos++
		}
		if strings.HasPrefix(p.str[p.pos:], "zero") {
			p.pos += 4
			zero = true
			vc++
			continue
		}
		// look for name:"foobar"
		if strings.HasPrefix(p.str[p.pos:], "name") {
			p.pos += 4
			if ok, err := p.nextChar(':'); err != nil {
				return "", false, err
			} else if !ok {
				return "", false, p.newParseError("expected ':'")
			}
			p.pos++
			if name, err = p.slurpQuotedString(); err != nil {
				return "", false, err
			}
			vc++
			continue
		}
		return "", false, p.newParseError("expected 'zero' or 'name'")
	}
	return name, zero, nil
}

// slurpQuotedString looks for a quoted string, which is always required
func (p *parser) slurpQuotedString() (string, error) {
	if !p.hasMore() {
		return "", p.newParseError("unexpected end")
	}
	if ok, err := p.nextChar('"'); err != nil {
		return "", err
	} else if !ok {
		return "", p.newParseError("expected '\"'")
	}
	p.pos++
	iend := p.pos
	for _, r := range p.str[p.pos:] {
		if r == '"' {
			break
		}
		iend++
	}
	if iend == p.pos {
		return "", p.newParseError("expected name")
	}
	name := p.str[p.pos:iend]
	p.pos = iend
	if ok, err := p.nextChar('"'); err != nil {
		return "", err
	} else if !ok {
		return "", p.newParseError("expected '\"'")
	}
	p.pos++
	return name, nil
}

// slurpDirOptions looks for an optional {} block which may optionally contain
// `name:"foo"`, or `sharded:X`, or just `sharded`, comma separated. Returns
// name and shardBitwidth. If neither are supplied, zero values are substituted.
// If `sharded` is supplied without bitwidth, the default of `4` is used.
func (p *parser) slurpDirOptions() (name string, shardBitwidth int, err error) {
	if !p.hasMore() {
		return "", 0, nil
	}
	if ok, err := p.nextChar('{'); err != nil {
		return "", 0, err
	} else if !ok {
		return "", 0, nil
	}
	p.pos++
	if !p.hasMore() {
		return "", 0, p.newParseError("unexpected end")
	}
	var vc int
	for p.hasMore() {
		if ok, err := p.nextChar('}'); err != nil {
			return "", 0, err
		} else if ok {
			p.pos++
			break
		}
		if vc > 0 {
			if ok, err := p.nextChar(','); err != nil {
				return "", 0, err
			} else if !ok {
				return "", 0, p.newParseError("expected ','")
			}
			p.pos++
		}
		if strings.HasPrefix(p.str[p.pos:], "sharded") {
			p.pos += 7
			shardBitwidth = 4
			if ok, err := p.nextChar(':'); err != nil {
				return "", 0, err
			} else if ok { // optional bitwidth specified
				p.pos++
				// extract the number
				var ok bool
				if shardBitwidth, ok, err = p.slurpInteger(); err != nil {
					return "", 0, err
				} else if !ok {
					return "", 0, p.newParseError("expected integer")
				} else if shardBitwidth <= 0 {
					return "", 0, p.newParseError("expected integer > 0")
				}
			}
			vc++
			continue
		}
		if strings.HasPrefix(p.str[p.pos:], "name") {
			p.pos += 4
			if ok, err := p.nextChar(':'); err != nil {
				return "", 0, err
			} else if !ok {
				return "", 0, p.newParseError("expected ':'")
			}
			p.pos++
			if name, err = p.slurpQuotedString(); err != nil {
				return "", 0, err
			}
			vc++
			continue
		}
		return "", 0, p.newParseError("expected 'sharded' or 'name'")
	}
	return name, shardBitwidth, nil
}

// slurpInteger parses an integer, if one exists, return the integer and true
// if one exists, false otherwise
func (p *parser) slurpInteger() (int, bool, error) {
	// figure out where the integers end, then parse the integer
	iend := p.pos
	for _, r := range p.str[p.pos:] {
		if r < '0' || r > '9' {
			break
		}
		iend++
	}
	if iend == p.pos {
		return 0, false, nil
	}
	ii, err := strconv.Atoi(p.str[p.pos:iend])
	if err != nil {
		return 0, false, p.newParseError("expected integer")
	}
	p.pos = iend
	return ii, true, nil
}

// slurpSize looks for a ':' followed by a human readable byte size, we'll use
// humanize.ParseBytes() for that but we should first collect [0-9.a-zA-Z] and
// parse that
func (p *parser) slurpSize() (uint64, bool, error) {
	if ok, err := p.nextChar(':'); err != nil {
		return 0, false, err
	} else if !ok {
		return 0, false, p.newParseError("expected ':'")
	}
	p.pos++
	rnd, err := p.slurpRandom()
	if err != nil {
		return 0, false, err
	}
	if !p.hasMore() {
		return 0, false, p.newParseError("unexpected end")
	}
	iend := p.pos
	// find the number portion
	for _, r := range p.str[p.pos:] {
		if !(unicode.IsDigit(r) || r == '.') {
			break
		}
		iend++
	}
	if iend == p.pos {
		return 0, false, p.newParseError("expected size")
	}
	// skip over spaces
	if iend < len(p.str) {
		for _, r := range p.str[iend:] {
			if !unicode.IsSpace(r) {
				break
			}
			iend++
		}
	}
	// find the units portion after the number
	if iend < len(p.str) {
		for _, r := range p.str[iend:] {
			if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z')) {
				break
			}
			iend++
		}
	}
	if iend == p.pos {
		return 0, false, p.newParseError("expected size")
	}
	int, err := humanize.ParseBytes(p.str[p.pos:iend])
	if err != nil {
		return 0, false, p.newParseError("expected human readable size: %w", err)
	}
	p.pos = iend
	return int, rnd, nil
}

// slurpComma looks for a ',', which is optional and used to indicate further
// list items
func (p *parser) slurpComma() (bool, error) {
	if ok, err := p.nextChar(','); err != nil {
		return false, err
	} else if ok {
		p.pos++
		return true, nil
	}
	return false, nil
}

// slurpOpen looks for a '(', which is strictly required
func (p *parser) slurpOpen() error {
	if ok, err := p.nextChar('('); err != nil {
		return err
	} else if !ok {
		return p.newParseError("expected '('")
	}
	p.pos++
	return nil
}

// slurpClose looks for a ')', which is strictly required
func (p *parser) slurpClose() error {
	if ok, err := p.nextChar(')'); err != nil {
		return err
	} else if !ok {
		return p.newParseError("expected ')'")
	}
	p.pos++
	return nil
}

// slurpRandom looks for '~', which is always optional
func (p *parser) slurpRandom() (bool, error) {
	if ok, err := p.nextChar('~'); err != nil {
		return false, err
	} else if ok {
		p.pos++
		return true, nil
	}
	return false, nil
}

// slurpMultiplier looks for an int multiplier, which is always optional but if
// present must be > 0 and must be followed by '*'
func (p *parser) slurpMultiplier() (multiplier int, err error) {
	var ok bool
	if multiplier, ok, err = p.slurpInteger(); err != nil {
		return 0, err
	} else if !ok {
		return 1, nil
	} else if multiplier < 0 {
		return 0, p.newParseError("expected integer >= 0")
	}
	if ok, err := p.nextChar('*'); err != nil {
		return 0, err
	} else if !ok {
		return 0, p.newParseError("expected '*'")
	}
	p.pos++
	return multiplier, nil
}

// slurpType looks for the strings "file" or "dir", which are strictly required
// to be next, nothing else is allowed
func (p *parser) slurpType() (string, error) {
	if !p.hasMore() {
		return "", p.newParseError("unexpected end")
	}
	if strings.HasPrefix(p.str[p.pos:], "file") {
		p.pos += 4
		return "file", nil
	}
	if strings.HasPrefix(p.str[p.pos:], "dir") {
		p.pos += 3
		return "dir", nil
	}
	return "", p.newParseError("expected 'file' or 'dir'")
}
