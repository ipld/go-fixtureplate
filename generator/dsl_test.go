package generator

import (
	"testing"

	"github.com/test-go/testify/require"
)

func TestParsing(t *testing.T) {
	testCases := []struct {
		input    string
		expected Entity
		err      string
	}{
		{
			input:    "file:1kib",
			expected: File{Multiplier: 1, Size: 1024},
		},
		{
			input:    "file:~1kib",
			expected: File{Multiplier: 1, Size: 1024, RandomSize: true},
		},
		{
			input:    `file:101{name:"beep boop"}`,
			expected: File{Multiplier: 1, Size: 101, Name: "beep boop"},
		},
		{
			input:    `file:1MiB{zero}`,
			expected: File{Multiplier: 1, Size: 1 << 20, ZeroContent: true},
		},
		{
			input:    `file:101{zero,name:"beep boop"}`,
			expected: File{Multiplier: 1, Size: 101, ZeroContent: true, Name: "beep boop"},
		},
		{
			input:    `file:101{name:"beep boop",zero}`,
			expected: File{Multiplier: 1, Size: 101, ZeroContent: true, Name: "beep boop"},
		},
		{
			input:    "dir(file:1K)",
			expected: Directory{Multiplier: 1, Type: DirType_Plain, Children: []Entity{File{Multiplier: 1, Size: 1000}}},
		},
		{
			input:    "dir{sharded}(file:1K)",
			expected: Directory{Multiplier: 1, Type: DirType_Sharded, ShardBitwidth: 4, Children: []Entity{File{Multiplier: 1, Size: 1000}}},
		},
		{
			input:    "dir{sharded:2}(file:1K)",
			expected: Directory{Multiplier: 1, Type: DirType_Sharded, ShardBitwidth: 2, Children: []Entity{File{Multiplier: 1, Size: 1000}}},
		},
		{
			input:    `dir{name:"blip blop"}(file:1K)`,
			expected: Directory{Multiplier: 1, Type: DirType_Plain, Name: "blip blop", Children: []Entity{File{Multiplier: 1, Size: 1000}}},
		},
		{
			input:    `dir{sharded,name:"blip blop"}(file:1K)`,
			expected: Directory{Multiplier: 1, Type: DirType_Sharded, ShardBitwidth: 4, Name: "blip blop", Children: []Entity{File{Multiplier: 1, Size: 1000}}},
		},
		{
			input:    `dir{name:"blip blop",sharded}(file:1K)`,
			expected: Directory{Multiplier: 1, Type: DirType_Sharded, ShardBitwidth: 4, Name: "blip blop", Children: []Entity{File{Multiplier: 1, Size: 1000}}},
		},
		{
			input:    `dir{sharded:3,name:"blip blop"}(file:1K)`,
			expected: Directory{Multiplier: 1, Type: DirType_Sharded, ShardBitwidth: 3, Name: "blip blop", Children: []Entity{File{Multiplier: 1, Size: 1000}}},
		},
		{
			input:    `dir{name:"blip blop",sharded:3}(file:1K)`,
			expected: Directory{Multiplier: 1, Type: DirType_Sharded, ShardBitwidth: 3, Name: "blip blop", Children: []Entity{File{Multiplier: 1, Size: 1000}}},
		},
		{
			input: `dir(file:1,file:2,file:3)`,
			expected: Directory{Multiplier: 1, Type: DirType_Plain, Children: []Entity{
				File{Multiplier: 1, Size: 1},
				File{Multiplier: 1, Size: 2},
				File{Multiplier: 1, Size: 3},
			}},
		},
		{
			input: `dir(file:1{name:"a"},file:2{zero},file:3)`,
			expected: Directory{Multiplier: 1, Type: DirType_Plain, Children: []Entity{
				File{Multiplier: 1, Name: "a", Size: 1},
				File{Multiplier: 1, Size: 2, ZeroContent: true},
				File{Multiplier: 1, Size: 3},
			}},
		},
		{
			input: `dir(dir(file:1),dir(file:2),dir(file:3))`,
			expected: Directory{Multiplier: 1, Type: DirType_Plain, Children: []Entity{
				Directory{Multiplier: 1, Type: DirType_Plain, Children: []Entity{File{Multiplier: 1, Size: 1}}},
				Directory{Multiplier: 1, Type: DirType_Plain, Children: []Entity{File{Multiplier: 1, Size: 2}}},
				Directory{Multiplier: 1, Type: DirType_Plain, Children: []Entity{File{Multiplier: 1, Size: 3}}},
			}},
		},
		{
			input:    "5*file:1kib",
			expected: File{Multiplier: 5, Size: 1024},
			err:      "root entity must be strictly signular",
		},
		{
			input:    "~5*file:1kib",
			expected: File{Multiplier: 5, RandomMultiplier: true, Size: 1024},
			err:      "root entity must be strictly signular",
		},
		{
			input:    "~5*file:~1kib",
			expected: File{Multiplier: 5, RandomMultiplier: true, Size: 1024, RandomSize: true},
			err:      "root entity must be strictly signular",
		},
		{
			input:    "dir(5*file:1kib)",
			expected: Directory{Multiplier: 1, Type: DirType_Plain, Children: []Entity{File{Multiplier: 5, Size: 1024}}},
		},
		{
			input:    "dir(~5*file:1kib)",
			expected: Directory{Multiplier: 1, Type: DirType_Plain, Children: []Entity{File{Multiplier: 5, RandomMultiplier: true, Size: 1024}}},
		},
		{
			input:    "dir(~5*file:~1kib)",
			expected: Directory{Multiplier: 1, Type: DirType_Plain, Children: []Entity{File{Multiplier: 5, RandomMultiplier: true, Size: 1024, RandomSize: true}}},
		},
		{
			input:    "10*dir(5*file:1kib)",
			expected: Directory{Multiplier: 10, Type: DirType_Plain, Children: []Entity{File{Multiplier: 5, Size: 1024}}},
			err:      "root entity must be strictly signular",
		},
		{
			input:    "~10*dir(~5*file:1kib)",
			expected: Directory{Multiplier: 10, RandomMultiplier: true, Type: DirType_Plain, Children: []Entity{File{Multiplier: 5, RandomMultiplier: true, Size: 1024}}},
			err:      "root entity must be strictly signular",
		},
		{
			input:    "dir(10*dir(5*file:1kib))",
			expected: Directory{Multiplier: 1, Type: DirType_Plain, Children: []Entity{Directory{Multiplier: 10, Type: DirType_Plain, Children: []Entity{File{Multiplier: 5, Size: 1024}}}}},
		},
		{
			input:    "dir(~10*dir(~5*file:1kib))",
			expected: Directory{Multiplier: 1, Type: DirType_Plain, Children: []Entity{Directory{Multiplier: 10, RandomMultiplier: true, Type: DirType_Plain, Children: []Entity{File{Multiplier: 5, RandomMultiplier: true, Size: 1024}}}}},
		},
		{
			input: `dir(1*dir(1*file:1),2*dir(2*file:2),~3*dir(~3*file:~3))`,
			expected: Directory{Multiplier: 1, Type: DirType_Plain, Children: []Entity{
				Directory{Multiplier: 1, Type: DirType_Plain, Children: []Entity{File{Multiplier: 1, Size: 1}}},
				Directory{Multiplier: 2, Type: DirType_Plain, Children: []Entity{File{Multiplier: 2, Size: 2}}},
				Directory{Multiplier: 3, RandomMultiplier: true, Type: DirType_Plain, Children: []Entity{File{Multiplier: 3, RandomMultiplier: true, Size: 3, RandomSize: true}}},
			}},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			actual, err := Parse(tc.input)
			if tc.err != "" {
				require.EqualError(t, err, tc.err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tc.expected, actual)
		})
	}
}
