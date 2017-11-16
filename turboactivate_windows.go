// Copyright 2017 wyDay, LLC. All rights reserved.

// +build windows

package turboactivate // import "golang.wyday.com/turboactivate"

/*
#cgo CFLAGS: -I .

#include "TurboActivate.h"
*/
import "C"

import (
	"unicode/utf16"
	"unsafe"
)

// TAStrPtrType is the data type of string pointers that will be passed
// to the TurboActivate library on this particular platform.
type TAStrPtrType *C.WCHAR

// getTAStrPtr gets the cwstring on Windows.
func getTAStrPtr(s string) TAStrPtrType {
	wstr := utf16.Encode([]rune(s))

	p := C.calloc(C.size_t(len(wstr)+1), 2)
	pp := (*[1 << 30]uint16)(p)
	copy(pp[:], wstr)

	return (TAStrPtrType)(p)
}

// getTAStrBufferPtr allocates and returns a buffer of the string length (including null)
func getTAStrBufferPtr(strLen C.size_t) TAStrPtrType {
	p := C.calloc(strLen, 2)
	return (TAStrPtrType)(p)
}

// stringFromTAStrPtr converts ptr to a Go string
func stringFromTAStrPtr(cwstr TAStrPtrType) string {
	ptr := unsafe.Pointer(cwstr)
	sz := C.wcslen((*C.wchar_t)(ptr))
	wstr := (*[1<<30 - 1]uint16)(ptr)[:sz:sz]
	return string(utf16.Decode(wstr))
}
