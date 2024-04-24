package utils

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

type hashTest struct {
	input  []byte
	output map[*HashType]string
}

var hashTestSet = []hashTest{
	{
		input: []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14},
		output: map[*HashType]string{
			MD5:    "bf13fc19e5151ac57d4252e0e0f87abe",
			SHA1:   "3ab6543c08a75f292a5ecedac87ec41642d12166",
			SHA256: "c839e57675862af5c21bd0a15413c3ec579e0d5522dab600bc6c3489b05b8f54",
		},
	},
	// Empty data set
	{
		input: []byte{},
		output: map[*HashType]string{
			MD5:    "d41d8cd98f00b204e9800998ecf8427e",
			SHA1:   "da39a3ee5e6b4b0d3255bfef95601890afd80709",
			SHA256: "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
		},
	},
}

func TestMultiHasher(t *testing.T) {
	for _, test := range hashTestSet {
		mh := NewMultiHasher([]*HashType{MD5, SHA1, SHA256})
		n, err := CopyWithBuffer(mh, bytes.NewBuffer(test.input))
		require.NoError(t, err)
		assert.Len(t, test.input, int(n))
		hashInfo := mh.GetHashInfo()
		for k, v := range hashInfo.h {
			expect, ok := test.output[k]
			require.True(t, ok, "test output for hash not found")
			assert.Equal(t, expect, v)
		}
		// Test that all are present
		for k, v := range test.output {
			expect, ok := hashInfo.h[k]
			require.True(t, ok, "test output for hash not found")
			assert.Equal(t, expect, v)
		}
		for k, v := range test.output {
			expect := hashInfo.GetHash(k)
			require.True(t, len(expect) > 0, "test output for hash not found")
			assert.Equal(t, expect, v)
		}
		expect := hashInfo.GetHash(nil)
		require.True(t, len(expect) == 0, "unknown type should return empty string")
		str := hashInfo.String()
		Log.Info("str=" + str)
		newHi := FromString(str)
		assert.Equal(t, newHi.h, hashInfo.h)

	}
}
