package generator

import (
	"testing"

	"github.com/test-go/testify/require"
)

func TestParsing(t *testing.T) {
	testCases := []struct {
		input     string
		expected  Entity
		explained string
		err       string
	}{
		{
			input:     `file:1kib`,
			expected:  File{Multiplier: 1, Size: 1024},
			explained: "A file of 1.0 KiB",
		},
		{
			input:     `file:~1kB`,
			expected:  File{Multiplier: 1, Size: 1000, RandomSize: true},
			explained: "A file of approximately 1.0 kB",
		},
		{
			input:     `file:101{name:"beep boop"}`,
			expected:  File{Multiplier: 1, Size: 101, Name: "beep boop"},
			explained: "A file named \"beep boop\" of 101 B",
		},
		{
			input:     `file:1MiB{zero}`,
			expected:  File{Multiplier: 1, Size: 1 << 20, ZeroContent: true},
			explained: "A file of 1.0 MiB containing just zeros",
		},
		{
			input:     `file:101{zero,name:"beep boop"}`,
			expected:  File{Multiplier: 1, Size: 101, ZeroContent: true, Name: "beep boop"},
			explained: "A file named \"beep boop\" of 101 B containing just zeros",
		},
		{
			input:     `file:101{name:"beep boop",zero}`,
			expected:  File{Multiplier: 1, Size: 101, ZeroContent: true, Name: "beep boop"},
			explained: "A file named \"beep boop\" of 101 B containing just zeros",
		},
		{
			input:     `dir(file:1K)`,
			expected:  Directory{Multiplier: 1, Type: DirType_Plain, Children: []Entity{File{Multiplier: 1, Size: 1000}}},
			explained: "A directory containing:\n  → A file of 1.0 kB",
		},
		{
			input:     `dir{sharded}(file:1K)`,
			expected:  Directory{Multiplier: 1, Type: DirType_Sharded, ShardBitwidth: 4, Children: []Entity{File{Multiplier: 1, Size: 1000}}},
			explained: "A directory sharded with bitwidth 4 containing:\n  → A file of 1.0 kB",
		},
		{
			input:     `dir{sharded:2}(file:1K)`,
			expected:  Directory{Multiplier: 1, Type: DirType_Sharded, ShardBitwidth: 2, Children: []Entity{File{Multiplier: 1, Size: 1000}}},
			explained: "A directory sharded with bitwidth 2 containing:\n  → A file of 1.0 kB",
		},
		{
			input:     `dir{name:"blip blop"}(file:1K)`,
			expected:  Directory{Multiplier: 1, Type: DirType_Plain, Name: "blip blop", Children: []Entity{File{Multiplier: 1, Size: 1000}}},
			explained: "A directory named \"blip blop\" containing:\n  → A file of 1.0 kB",
		},
		{
			input:     `dir{sharded,name:"blip blop"}(file:1K)`,
			expected:  Directory{Multiplier: 1, Type: DirType_Sharded, ShardBitwidth: 4, Name: "blip blop", Children: []Entity{File{Multiplier: 1, Size: 1000}}},
			explained: "A directory named \"blip blop\" sharded with bitwidth 4 containing:\n  → A file of 1.0 kB",
		},
		{
			input:     `dir{name:"blip blop",sharded}(file:1K)`,
			expected:  Directory{Multiplier: 1, Type: DirType_Sharded, ShardBitwidth: 4, Name: "blip blop", Children: []Entity{File{Multiplier: 1, Size: 1000}}},
			explained: "A directory named \"blip blop\" sharded with bitwidth 4 containing:\n  → A file of 1.0 kB",
		},
		{
			input:     `dir{sharded:3,name:"blip blop"}(file:1K)`,
			expected:  Directory{Multiplier: 1, Type: DirType_Sharded, ShardBitwidth: 3, Name: "blip blop", Children: []Entity{File{Multiplier: 1, Size: 1000}}},
			explained: "A directory named \"blip blop\" sharded with bitwidth 3 containing:\n  → A file of 1.0 kB",
		},
		{
			input:     `dir{name:"blip blop",sharded:3}(file:1K)`,
			expected:  Directory{Multiplier: 1, Type: DirType_Sharded, ShardBitwidth: 3, Name: "blip blop", Children: []Entity{File{Multiplier: 1, Size: 1000}}},
			explained: "A directory named \"blip blop\" sharded with bitwidth 3 containing:\n  → A file of 1.0 kB",
		},
		{
			input: `dir(file:1,file:2,file:3)`,
			expected: Directory{Multiplier: 1, Type: DirType_Plain, Children: []Entity{
				File{Multiplier: 1, Size: 1},
				File{Multiplier: 1, Size: 2},
				File{Multiplier: 1, Size: 3},
			}},
			explained: "A directory containing:\n  → A file of 1 B\n  → A file of 2 B\n  → A file of 3 B",
		},
		{
			input: `dir(file:1{name:"a"},file:2{zero},file:3)`,
			expected: Directory{Multiplier: 1, Type: DirType_Plain, Children: []Entity{
				File{Multiplier: 1, Name: "a", Size: 1},
				File{Multiplier: 1, Size: 2, ZeroContent: true},
				File{Multiplier: 1, Size: 3},
			}},
			explained: "A directory containing:\n  → A file named \"a\" of 1 B\n  → A file of 2 B containing just zeros\n  → A file of 3 B",
		},
		{
			input: `dir(dir(file:1),dir(file:2),dir(file:3))`,
			expected: Directory{Multiplier: 1, Type: DirType_Plain, Children: []Entity{
				Directory{Multiplier: 1, Type: DirType_Plain, Children: []Entity{File{Multiplier: 1, Size: 1}}},
				Directory{Multiplier: 1, Type: DirType_Plain, Children: []Entity{File{Multiplier: 1, Size: 2}}},
				Directory{Multiplier: 1, Type: DirType_Plain, Children: []Entity{File{Multiplier: 1, Size: 3}}},
			}},
			explained: "A directory containing:\n  → A directory containing:\n    → A file of 1 B\n  → A directory containing:\n    → A file of 2 B\n  → A directory containing:\n    → A file of 3 B",
		},
		{
			input: "5*file:1kib",
			err:   "root entity must be strictly signular",
		},
		{
			input: "~5*file:1kib",
			err:   "root entity must be strictly signular",
		},
		{
			input: "~5*file:~1kib",
			err:   "root entity must be strictly signular",
		},
		{
			input:     "dir(5*file:1kib)",
			expected:  Directory{Multiplier: 1, Type: DirType_Plain, Children: []Entity{File{Multiplier: 5, Size: 1024}}},
			explained: "A directory containing:\n  → 5 files of 1.0 KiB",
		},
		{
			input:     "dir(~5*file:1kib)",
			expected:  Directory{Multiplier: 1, Type: DirType_Plain, Children: []Entity{File{Multiplier: 5, RandomMultiplier: true, Size: 1024}}},
			explained: "A directory containing:\n  → Approximately 5 files of 1.0 KiB",
		},
		{
			input:     "dir(~5*file:~1kib)",
			expected:  Directory{Multiplier: 1, Type: DirType_Plain, Children: []Entity{File{Multiplier: 5, RandomMultiplier: true, Size: 1024, RandomSize: true}}},
			explained: "A directory containing:\n  → Approximately 5 files of approximately 1.0 KiB",
		},
		{
			input: "10*dir(5*file:1kib)",
			err:   "root entity must be strictly signular",
		},
		{
			input: "~10*dir(~5*file:1kib)",
			err:   "root entity must be strictly signular",
		},
		{
			input:     "dir(10*dir(5*file:1kib))",
			expected:  Directory{Multiplier: 1, Type: DirType_Plain, Children: []Entity{Directory{Multiplier: 10, Type: DirType_Plain, Children: []Entity{File{Multiplier: 5, Size: 1024}}}}},
			explained: "A directory containing:\n  → 10 directories containing:\n    → 5 files of 1.0 KiB",
		},
		{
			input:     "dir(~10*dir(~5*file:1kib))",
			expected:  Directory{Multiplier: 1, Type: DirType_Plain, Children: []Entity{Directory{Multiplier: 10, RandomMultiplier: true, Type: DirType_Plain, Children: []Entity{File{Multiplier: 5, RandomMultiplier: true, Size: 1024}}}}},
			explained: "A directory containing:\n  → Approximately 10 directories containing:\n    → Approximately 5 files of 1.0 KiB",
		},
		{
			input: `dir(1*dir(1*file:1),2*dir(2*file:2),~3*dir(~3*file:~3))`,
			expected: Directory{Multiplier: 1, Type: DirType_Plain, Children: []Entity{
				Directory{Multiplier: 1, Type: DirType_Plain, Children: []Entity{File{Multiplier: 1, Size: 1}}},
				Directory{Multiplier: 2, Type: DirType_Plain, Children: []Entity{File{Multiplier: 2, Size: 2}}},
				Directory{Multiplier: 3, RandomMultiplier: true, Type: DirType_Plain, Children: []Entity{File{Multiplier: 3, RandomMultiplier: true, Size: 3, RandomSize: true}}},
			}},
			explained: "A directory containing:\n  → A directory containing:\n    → A file of 1 B\n  → 2 directories containing:\n    → 2 files of 2 B\n  → Approximately 3 directories containing:\n    → Approximately 3 files of approximately 3 B",
		},
		{
			// sanity check for the next 2
			input:     `1*dir{name:"boop"}(file:1kib)`,
			expected:  Directory{Multiplier: 1, Type: DirType_Plain, Name: "boop", Children: []Entity{File{Multiplier: 1, Size: 1024}}},
			explained: "A directory named \"boop\" containing:\n  → A file of 1.0 KiB",
		},
		{
			input: `2*dir{name:"boop"}(file:1kib)`,
			err:   "can't name a directory with a multiplier",
		},
		{
			input: `~1*dir{name:"boop"}(file:1kib)`,
			err:   "can't name a directory with a multiplier",
		},
		{
			// sanity check for the next 2
			input:     `dir(1*file:1kib{name:"boop"})`,
			expected:  Directory{Multiplier: 1, Type: DirType_Plain, Children: []Entity{File{Multiplier: 1, Size: 1024, Name: "boop"}}},
			explained: "A directory containing:\n  → A file named \"boop\" of 1.0 KiB",
		},
		{
			input: `dir(2*file:1kib{name:"boop"})`,
			err:   "can't name a file with a multiplier",
		},
		{
			input: `dir(~1*file:1kib{name:"boop"})`,
			err:   "can't name a file with a multiplier",
		},
		{
			input: `dir(~5*file:1.0kB,~5*file:~102kB,2*dir{sharded}(~10*file:51kB),file:1.0MB{zero},file:10B,file:20B)`,
			expected: Directory{
				Multiplier: 1,
				Type:       DirType_Plain,
				Children: []Entity{
					File{Multiplier: 5, RandomMultiplier: true, Size: 1000},
					File{Multiplier: 5, RandomMultiplier: true, Size: 102000, RandomSize: true},
					Directory{Multiplier: 2, Type: DirType_Sharded, ShardBitwidth: 4, Children: []Entity{File{Multiplier: 10, RandomMultiplier: true, Size: 51000}}},
					File{Multiplier: 1, Size: 1000000, ZeroContent: true},
					File{Multiplier: 1, Size: 10},
					File{Multiplier: 1, Size: 20},
				},
			},
			explained: "A directory containing:\n  → Approximately 5 files of 1.0 kB\n  → Approximately 5 files of approximately 102 kB\n  → 2 directories sharded with bitwidth 4 containing:\n    → Approximately 10 files of 51 kB\n  → A file of 1.0 MB containing just zeros\n  → A file of 10 B\n  → A file of 20 B",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			actual, err := Parse(tc.input)
			if tc.err != "" {
				require.NotNil(t, err)
				require.Contains(t, err.Error(), tc.err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tc.expected, actual)
			if tc.explained != "" {
				require.Equal(t, tc.explained, actual.Describe(""))
			}
		})
	}
}
