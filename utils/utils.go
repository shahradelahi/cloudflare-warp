package utils

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"math/big"
)

func Uint32ToBytes(n uint32) []byte {
	b := make([]byte, 4)
	binary.LittleEndian.PutUint32(b, n)
	return b
}

func RandomInt(min, max uint64) (uint64, error) {
	if min > max {
		return 0, fmt.Errorf("min cannot be greater than max")
	}
	if min == max {
		return min, nil
	}
	rangeVal := max - min + 1

	n, err := rand.Int(rand.Reader, big.NewInt(int64(rangeVal)))
	if err != nil {
		return 0, err
	}

	return min + n.Uint64(), nil
}
