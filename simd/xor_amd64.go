// Code generated from " ../gtl/generate.py --prefix=Xor -DOPCHAR=^ --package=simd --output=xor_amd64.go bitwise_amd64.go.tpl ". DO NOT EDIT.
// Copyright 2018 GRAIL, Inc.  All rights reserved.
// Use of this source code is governed by the Apache-2.0
// license that can be found in the LICENSE file.

// +build amd64,!appengine

package simd

import (
	"reflect"
	"unsafe"
)

// XorUnsafeInplace sets main[pos] := arg[pos] ^ main[pos] for every position
// in main[].
//
// WARNING: This is a function designed to be used in inner loops, which makes
// assumptions about length and capacity which aren't checked at runtime.  Use
// the safe version of this function when that's a problem.
// Assumptions #2-3 are always satisfied when the last
// potentially-size-increasing operation on arg[] is {Re}makeUnsafe(),
// ResizeUnsafe(), or XcapUnsafe(), and the same is true for main[].
//
// 1. len(arg) and len(main) must be equal.
//
// 2. Capacities are at least RoundUpPow2(len(main) + 1, bytesPerVec).
//
// 3. The caller does not care if a few bytes past the end of main[] are
// changed.
func XorUnsafeInplace(main, arg []byte) {
	mainLen := len(main)
	argHeader := (*reflect.SliceHeader)(unsafe.Pointer(&arg))
	mainHeader := (*reflect.SliceHeader)(unsafe.Pointer(&main))
	argWordsIter := unsafe.Pointer(argHeader.Data)
	mainWordsIter := unsafe.Pointer(mainHeader.Data)
	if mainLen > 2*BytesPerWord {
		nWordMinus2 := (mainLen - BytesPerWord - 1) >> Log2BytesPerWord
		for widx := 0; widx < nWordMinus2; widx++ {
			mainWord := *((*uintptr)(mainWordsIter))
			argWord := *((*uintptr)(argWordsIter))
			*((*uintptr)(mainWordsIter)) = mainWord ^ argWord
			mainWordsIter = unsafe.Pointer(uintptr(mainWordsIter) + BytesPerWord)
			argWordsIter = unsafe.Pointer(uintptr(argWordsIter) + BytesPerWord)
		}
	} else if mainLen <= BytesPerWord {
		mainWord := *((*uintptr)(mainWordsIter))
		argWord := *((*uintptr)(argWordsIter))
		*((*uintptr)(mainWordsIter)) = mainWord ^ argWord
		return
	}
	// The last two read-and-writes to main[] usually overlap.  To avoid a
	// store-to-load forwarding slowdown, we read both words before writing
	// either.
	// shuffleLookupOddInplaceSSSE3Asm() uses the same strategy.
	mainWord1 := *((*uintptr)(mainWordsIter))
	argWord1 := *((*uintptr)(argWordsIter))
	finalOffset := uintptr(mainLen - BytesPerWord)
	mainFinalWordPtr := unsafe.Pointer(mainHeader.Data + finalOffset)
	argFinalWordPtr := unsafe.Pointer(argHeader.Data + finalOffset)
	mainWord2 := *((*uintptr)(mainFinalWordPtr))
	argWord2 := *((*uintptr)(argFinalWordPtr))
	*((*uintptr)(mainWordsIter)) = mainWord1 ^ argWord1
	*((*uintptr)(mainFinalWordPtr)) = mainWord2 ^ argWord2
}

// XorInplace sets main[pos] := arg[pos] ^ main[pos] for every position in
// main[].  It panics if slice lengths don't match.
func XorInplace(main, arg []byte) {
	// This takes ~6-8% longer than XorUnsafeInplace on the short-array benchmark
	// on my Mac.
	mainLen := len(main)
	if len(arg) != mainLen {
		panic("XorInplace() requires len(arg) == len(main).")
	}
	if mainLen < BytesPerWord {
		// It's probably possible to do better here (e.g. when mainLen is in 4..7,
		// operate on uint32s), but I won't worry about it unless/until that's
		// actually a common case.
		for pos, argByte := range arg {
			main[pos] = main[pos] ^ argByte
		}
		return
	}
	argHeader := (*reflect.SliceHeader)(unsafe.Pointer(&arg))
	mainHeader := (*reflect.SliceHeader)(unsafe.Pointer(&main))
	argWordsIter := unsafe.Pointer(argHeader.Data)
	mainWordsIter := unsafe.Pointer(mainHeader.Data)
	if mainLen > 2*BytesPerWord {
		nWordMinus2 := (mainLen - BytesPerWord - 1) >> Log2BytesPerWord
		for widx := 0; widx < nWordMinus2; widx++ {
			mainWord := *((*uintptr)(mainWordsIter))
			argWord := *((*uintptr)(argWordsIter))
			*((*uintptr)(mainWordsIter)) = mainWord ^ argWord
			mainWordsIter = unsafe.Pointer(uintptr(mainWordsIter) + BytesPerWord)
			argWordsIter = unsafe.Pointer(uintptr(argWordsIter) + BytesPerWord)
		}
	}
	mainWord1 := *((*uintptr)(mainWordsIter))
	argWord1 := *((*uintptr)(argWordsIter))
	finalOffset := uintptr(mainLen - BytesPerWord)
	mainFinalWordPtr := unsafe.Pointer(mainHeader.Data + finalOffset)
	argFinalWordPtr := unsafe.Pointer(argHeader.Data + finalOffset)
	mainWord2 := *((*uintptr)(mainFinalWordPtr))
	argWord2 := *((*uintptr)(argFinalWordPtr))
	*((*uintptr)(mainWordsIter)) = mainWord1 ^ argWord1
	*((*uintptr)(mainFinalWordPtr)) = mainWord2 ^ argWord2
}

// XorUnsafe sets dst[pos] := src1[pos] ^ src2[pos] for every position in dst.
//
// WARNING: This is a function designed to be used in inner loops, which makes
// assumptions about length and capacity which aren't checked at runtime.  Use
// the safe version of this function when that's a problem.
// Assumptions #2-3 are always satisfied when the last
// potentially-size-increasing operation on src1[] is {Re}makeUnsafe(),
// ResizeUnsafe(), or XcapUnsafe(), and the same is true for src2[] and dst[].
//
// 1. len(src1), len(src2), and len(dst) must be equal.
//
// 2. Capacities are at least RoundUpPow2(len(dst) + 1, bytesPerVec).
//
// 3. The caller does not care if a few bytes past the end of dst[] are
// changed.
func XorUnsafe(dst, src1, src2 []byte) {
	src1Header := (*reflect.SliceHeader)(unsafe.Pointer(&src1))
	src2Header := (*reflect.SliceHeader)(unsafe.Pointer(&src2))
	dstHeader := (*reflect.SliceHeader)(unsafe.Pointer(&dst))
	nWord := DivUpPow2(len(dst), BytesPerWord, Log2BytesPerWord)

	src1Iter := unsafe.Pointer(src1Header.Data)
	src2Iter := unsafe.Pointer(src2Header.Data)
	dstIter := unsafe.Pointer(dstHeader.Data)
	for widx := 0; widx < nWord; widx++ {
		src1Word := *((*uintptr)(src1Iter))
		src2Word := *((*uintptr)(src2Iter))
		*((*uintptr)(dstIter)) = src1Word ^ src2Word
		src1Iter = unsafe.Pointer(uintptr(src1Iter) + BytesPerWord)
		src2Iter = unsafe.Pointer(uintptr(src2Iter) + BytesPerWord)
		dstIter = unsafe.Pointer(uintptr(dstIter) + BytesPerWord)
	}
}

// Xor sets dst[pos] := src1[pos] ^ src2[pos] for every position in dst.  It
// panics if slice lengths don't match.
func Xor(dst, src1, src2 []byte) {
	dstLen := len(dst)
	if (len(src1) != dstLen) || (len(src2) != dstLen) {
		panic("Xor() requires len(src1) == len(src2) == len(dst).")
	}
	if dstLen < BytesPerWord {
		for pos, src1Byte := range src1 {
			dst[pos] = src1Byte ^ src2[pos]
		}
		return
	}
	src1Header := (*reflect.SliceHeader)(unsafe.Pointer(&src1))
	src2Header := (*reflect.SliceHeader)(unsafe.Pointer(&src2))
	dstHeader := (*reflect.SliceHeader)(unsafe.Pointer(&dst))
	nWordMinus1 := (dstLen - 1) >> Log2BytesPerWord

	src1Iter := unsafe.Pointer(src1Header.Data)
	src2Iter := unsafe.Pointer(src2Header.Data)
	dstIter := unsafe.Pointer(dstHeader.Data)
	for widx := 0; widx < nWordMinus1; widx++ {
		src1Word := *((*uintptr)(src1Iter))
		src2Word := *((*uintptr)(src2Iter))
		*((*uintptr)(dstIter)) = src1Word ^ src2Word
		src1Iter = unsafe.Pointer(uintptr(src1Iter) + BytesPerWord)
		src2Iter = unsafe.Pointer(uintptr(src2Iter) + BytesPerWord)
		dstIter = unsafe.Pointer(uintptr(dstIter) + BytesPerWord)
	}
	// No store-forwarding problem here.
	finalOffset := uintptr(dstLen - BytesPerWord)
	src1Iter = unsafe.Pointer(src1Header.Data + finalOffset)
	src2Iter = unsafe.Pointer(src2Header.Data + finalOffset)
	dstIter = unsafe.Pointer(dstHeader.Data + finalOffset)
	src1Word := *((*uintptr)(src1Iter))
	src2Word := *((*uintptr)(src2Iter))
	*((*uintptr)(dstIter)) = src1Word ^ src2Word
}

// XorConst8UnsafeInplace sets main[pos] := main[pos] ^ val for every position
// in main[].
//
// WARNING: This is a function designed to be used in inner loops, which makes
// assumptions about length and capacity which aren't checked at runtime.  Use
// the safe version of this function when that's a problem.
// These assumptions are always satisfied when the last
// potentially-size-increasing operation on main[] is {Re}makeUnsafe(),
// ResizeUnsafe(), or XcapUnsafe().
//
// 1. cap(main) is at least RoundUpPow2(len(main) + 1, bytesPerVec).
//
// 2. The caller does not care if a few bytes past the end of main[] are
// changed.
func XorConst8UnsafeInplace(main []byte, val byte) {
	mainLen := len(main)
	argWord := 0x101010101010101 * uintptr(val)
	mainHeader := (*reflect.SliceHeader)(unsafe.Pointer(&main))
	mainWordsIter := unsafe.Pointer(mainHeader.Data)
	if mainLen > 2*BytesPerWord {
		nWordMinus2 := (mainLen - BytesPerWord - 1) >> Log2BytesPerWord
		for widx := 0; widx < nWordMinus2; widx++ {
			mainWord := *((*uintptr)(mainWordsIter))
			*((*uintptr)(mainWordsIter)) = mainWord ^ argWord
			mainWordsIter = unsafe.Pointer(uintptr(mainWordsIter) + BytesPerWord)
		}
	} else if mainLen <= BytesPerWord {
		mainWord := *((*uintptr)(mainWordsIter))
		*((*uintptr)(mainWordsIter)) = mainWord ^ argWord
		return
	}
	mainWord1 := *((*uintptr)(mainWordsIter))
	finalOffset := uintptr(mainLen - BytesPerWord)
	mainFinalWordPtr := unsafe.Pointer(mainHeader.Data + finalOffset)
	mainWord2 := *((*uintptr)(mainFinalWordPtr))
	*((*uintptr)(mainWordsIter)) = mainWord1 ^ argWord
	*((*uintptr)(mainFinalWordPtr)) = mainWord2 ^ argWord
}

// XorConst8Inplace sets main[pos] := main[pos] ^ val for every position in
// main[].
func XorConst8Inplace(main []byte, val byte) {
	mainLen := len(main)
	if mainLen < BytesPerWord {
		for pos, mainByte := range main {
			main[pos] = mainByte ^ val
		}
		return
	}
	argWord := 0x101010101010101 * uintptr(val)
	mainHeader := (*reflect.SliceHeader)(unsafe.Pointer(&main))
	mainWordsIter := unsafe.Pointer(mainHeader.Data)
	if mainLen > 2*BytesPerWord {
		nWordMinus2 := (mainLen - BytesPerWord - 1) >> Log2BytesPerWord
		for widx := 0; widx < nWordMinus2; widx++ {
			mainWord := *((*uintptr)(mainWordsIter))
			*((*uintptr)(mainWordsIter)) = mainWord ^ argWord
			mainWordsIter = unsafe.Pointer(uintptr(mainWordsIter) + BytesPerWord)
		}
	}
	mainWord1 := *((*uintptr)(mainWordsIter))
	finalOffset := uintptr(mainLen - BytesPerWord)
	mainFinalWordPtr := unsafe.Pointer(mainHeader.Data + finalOffset)
	mainWord2 := *((*uintptr)(mainFinalWordPtr))
	*((*uintptr)(mainWordsIter)) = mainWord1 ^ argWord
	*((*uintptr)(mainFinalWordPtr)) = mainWord2 ^ argWord
}

// XorConst8Unsafe sets dst[pos] := src[pos] ^ val for every position in dst.
//
// WARNING: This is a function designed to be used in inner loops, which makes
// assumptions about length and capacity which aren't checked at runtime.  Use
// the safe version of this function when that's a problem.
// Assumptions #2-3 are always satisfied when the last
// potentially-size-increasing operation on src[] is {Re}makeUnsafe(),
// ResizeUnsafe(), or XcapUnsafe(), and the same is true for dst[].
//
// 1. len(src) and len(dst) must be equal.
//
// 2. Capacities are at least RoundUpPow2(len(dst) + 1, bytesPerVec).
//
// 3. The caller does not care if a few bytes past the end of dst[] are
// changed.
func XorConst8Unsafe(dst, src []byte, val byte) {
	srcHeader := (*reflect.SliceHeader)(unsafe.Pointer(&src))
	dstHeader := (*reflect.SliceHeader)(unsafe.Pointer(&dst))
	nWord := DivUpPow2(len(dst), BytesPerWord, Log2BytesPerWord)
	argWord := 0x101010101010101 * uintptr(val)

	srcIter := unsafe.Pointer(srcHeader.Data)
	dstIter := unsafe.Pointer(dstHeader.Data)
	for widx := 0; widx < nWord; widx++ {
		srcWord := *((*uintptr)(srcIter))
		*((*uintptr)(dstIter)) = srcWord ^ argWord
		srcIter = unsafe.Pointer(uintptr(srcIter) + BytesPerWord)
		dstIter = unsafe.Pointer(uintptr(dstIter) + BytesPerWord)
	}
}

// XorConst8 sets dst[pos] := src[pos] ^ val for every position in dst.  It
// panics if slice lengths don't match.
func XorConst8(dst, src []byte, val byte) {
	dstLen := len(dst)
	if len(src) != dstLen {
		panic("XorConst8() requires len(src) == len(dst).")
	}
	if dstLen < BytesPerWord {
		for pos, srcByte := range src {
			dst[pos] = srcByte ^ val
		}
		return
	}
	srcHeader := (*reflect.SliceHeader)(unsafe.Pointer(&src))
	dstHeader := (*reflect.SliceHeader)(unsafe.Pointer(&dst))
	nWordMinus1 := (dstLen - 1) >> Log2BytesPerWord
	argWord := 0x101010101010101 * uintptr(val)

	srcIter := unsafe.Pointer(srcHeader.Data)
	dstIter := unsafe.Pointer(dstHeader.Data)
	for widx := 0; widx < nWordMinus1; widx++ {
		srcWord := *((*uintptr)(srcIter))
		*((*uintptr)(dstIter)) = srcWord ^ argWord
		srcIter = unsafe.Pointer(uintptr(srcIter) + BytesPerWord)
		dstIter = unsafe.Pointer(uintptr(dstIter) + BytesPerWord)
	}
	finalOffset := uintptr(dstLen - BytesPerWord)
	srcIter = unsafe.Pointer(srcHeader.Data + finalOffset)
	dstIter = unsafe.Pointer(dstHeader.Data + finalOffset)
	srcWord := *((*uintptr)(srcIter))
	*((*uintptr)(dstIter)) = srcWord ^ argWord
}
