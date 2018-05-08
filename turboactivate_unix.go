// Copyright 2018 wyDay, LLC. All rights reserved.

// +build !windows

package turboactivate // import "golang.wyday.com/turboactivate"

/*
#cgo CFLAGS: -I .

#include "TurboActivate.h"
*/
import "C"

type TAStrPtrType *C.CHAR

// getTAStrPtr gets the cstring on Unix.
func getTAStrPtr(s string) TAStrPtrType {
	return C.CString(s)
}

// getTAStrBufferPtr allocates and returns a buffer of the string length (including null)
func getTAStrBufferPtr(strLen C.size_t) TAStrPtrType {
	p := C.calloc(strLen, 1)
	return (TAStrPtrType)(p)
}

// stringFromTAStrPtr converts ptr to a Go string
func stringFromTAStrPtr(cstr TAStrPtrType) string {
	return C.GoString(cstr)
}
