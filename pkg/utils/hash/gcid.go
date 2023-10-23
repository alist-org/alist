package hash_extend

import (
	"crypto/sha1"
	"encoding"
	"fmt"
	"hash"
	"strconv"

	"github.com/alist-org/alist/v3/pkg/utils"
)

var GCID = utils.RegisterHashWithParam("gcid", "GCID", 40, func(a ...any) hash.Hash {
	var (
		size int64
		err  error
	)
	if len(a) > 0 {
		size, err = strconv.ParseInt(fmt.Sprint(a[0]), 10, 64)
		if err != nil {
			panic(err)
		}
	}
	return NewGcid(size)
})

func NewGcid(size int64) hash.Hash {
	calcBlockSize := func(j int64) int64 {
		var psize int64 = 0x40000
		for float64(j)/float64(psize) > 0x200 && psize < 0x200000 {
			psize = psize << 1
		}
		return psize
	}

	return &gcid{
		hash:      sha1.New(),
		hashState: sha1.New(),
		blockSize: int(calcBlockSize(size)),
	}
}

type gcid struct {
	hash      hash.Hash
	hashState hash.Hash
	blockSize int

	offset int
}

func (h *gcid) Write(p []byte) (n int, err error) {
	n = len(p)
	for len(p) > 0 {
		if h.offset < h.blockSize {
			var lastSize = h.blockSize - h.offset
			if lastSize > len(p) {
				lastSize = len(p)
			}

			h.hashState.Write(p[:lastSize])
			h.offset += lastSize
			p = p[lastSize:]
		}

		if h.offset >= h.blockSize {
			h.hash.Write(h.hashState.Sum(nil))
			h.hashState.Reset()
			h.offset = 0
		}
	}
	return
}

func (h *gcid) Sum(b []byte) []byte {
	if h.offset != 0 {
		if hashm, ok := h.hash.(encoding.BinaryMarshaler); ok {
			if hashum, ok := h.hash.(encoding.BinaryUnmarshaler); ok {
				tempData, _ := hashm.MarshalBinary()
				defer hashum.UnmarshalBinary(tempData)
				h.hash.Write(h.hashState.Sum(nil))
			}
		}
	}
	return h.hash.Sum(b)
}

func (h *gcid) Reset() {
	h.hash.Reset()
	h.hashState.Reset()
}

func (h *gcid) Size() int {
	return h.hash.Size()
}

func (h *gcid) BlockSize() int {
	return h.blockSize
}
