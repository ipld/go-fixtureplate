package fixtureplate

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

type DagScope string

const DagScopeAll DagScope = "all"
const DagScopeEntity DagScope = "entity"
const DagScopeBlock DagScope = "block"

func AsScope(scope string) (DagScope, error) {
	switch scope {
	case "all":
		return DagScopeAll, nil
	case "entity":
		return DagScopeEntity, nil
	case "block":
		return DagScopeBlock, nil
	default:
		return "", fmt.Errorf("invalid scope: %s", scope)
	}
}

type ByteRange struct {
	From int64
	To   int64
}

func (br ByteRange) String() string {
	to := strconv.FormatInt(br.To, 10)
	if br.To == math.MaxInt64 {
		to = "*"
	}
	return fmt.Sprintf("%d:%s", br.From, to)
}

func (br ByteRange) IsDefault() bool {
	return br.From == 0 && br.To == math.MaxInt64
}

func ParseByteRange(s string) (ByteRange, error) {
	br := ByteRange{From: 0, To: math.MaxInt64}
	if s == "" {
		return br, nil
	}
	parts := strings.Split(s, ":")
	if len(parts) != 2 {
		return br, fmt.Errorf("invalid byte range: %s", s)
	}
	var err error
	br.From, err = strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return br, err
	}
	if br.From < 0 {
		return br, fmt.Errorf("invalid byte range: %s", s)
	}
	if parts[1] != "*" {
		br.To, err = strconv.ParseInt(parts[1], 10, 64)
		if err != nil {
			return br, err
		}
	}
	return br, nil
}
