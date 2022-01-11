package utils

import (
	"unsafe"
)

func IntToBits(integer int64) (bits []int) {
	bits = make([]int, 64)

	// we need to use unsafe in order to correctly represent negative numbers
	integerU := *(*uint64)(unsafe.Pointer(&integer))

	for i, j := integerU, 63; i > 0; i, j = i/2, j-1 {
		bits[j] = int(i % 2)
	}

	return
}

// function to convert a signed integer to a slice of bits

func BitsToInt(bits []int) (r int) {
	for _, b := range bits {
		r = r<<1 | b
	}

	return
}
