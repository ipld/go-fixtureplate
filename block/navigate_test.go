package block

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/ipfs/go-unixfsnode"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"
	trustlessutils "github.com/ipld/go-trustless-utils"
	trustlesspathing "github.com/ipld/ipld/specs/pkg-go/trustless-pathing"
	"github.com/test-go/testify/require"
	"github.com/warpfork/go-testmark"
)

func TestNavigateUnixfs20MVariety(t *testing.T) {
	storage, closer, err := trustlesspathing.Unixfs20mVarietyReadableStorage()
	require.NoError(t, err)
	defer closer.Close()

	lsys := cidlink.DefaultLinkSystem()
	lsys.TrustedStorage = true
	unixfsnode.AddUnixFSReificationToLinkSystem(&lsys)
	lsys.SetReadStorage(storage)

	testCases, err := unixfs20mVarietyCases()
	require.NoError(t, err)

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			req := require.New(t)
			t.Logf("query=%s", tc.Query)

			root, path, scope, duplicates, byteRange, err := ParseQuery(tc.Query)
			req.NoError(err)
			blk, err := NewBlock(lsys, root)
			req.NoError(err)

			var buf bytes.Buffer
			visitor := WritingVisitor(&buf, duplicates, true)
			br := trustlessutils.ByteRange{}
			if byteRange != nil {
				br = *byteRange
			}
			req.NoError(blk.Navigate(path, scope, br, visitor))

			req.Equal(tc.Execution, buf.String())
		})
	}
}

// --- This is a very close re-implementation of
// ipld/ipld/specs/pkg-go/trustless-pathing/unixfs_20m_variety.go, because we
// need direct access to the "execution" block rather than the parsed form
// TODO: consider exposing the execution block directly over there

type testCase struct {
	Name      string
	Query     string
	Execution string
}

func unixfs20mVarietyCases() ([]testCase, error) {
	carPath := trustlesspathing.Unixfs20mVarietyCARPath()
	// replace the .car with .md of carPath
	mdPath := carPath[:len(carPath)-3] + "md"
	doc, err := testmark.ReadFile(mdPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read testcases: %w", err)
	}
	doc.BuildDirIndex()
	testCases := make([]testCase, 0)
	for _, test := range doc.DirEnt.Children["test"].ChildrenList {
		for _, scope := range test.ChildrenList {
			name := test.Name + "/" + scope.Name
			query := strings.ReplaceAll(dstr(scope, "query"), "\n", "")
			execution := dstr(scope, "execution")
			testCases = append(testCases, testCase{name, query, execution})
		}
	}
	return testCases, nil
}

func dstr(dir *testmark.DirEnt, ch string) string {
	return string(dir.Children[ch].Hunk.Body)
}
