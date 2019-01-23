package nfd

// #cgo LDFLAGS: -L${SRCDIR} -lnfd
// #cgo windows LDFLAGS: -lole32 -luuid
// #include <stdlib.h>
// #include "nfd.h"
import "C"
import (
	"errors"
	"unsafe"
)

func OpenDialog(filterList, defaultPath string) (res string, err error) {
	cFilterList := cstring(filterList)
	cDefaultPath := cstring(defaultPath)

	var cOutPath *C.char
	switch C.NFD_OpenDialog(cFilterList, cDefaultPath, &cOutPath) {
	case C.NFD_OKAY:
		res = C.GoString(cOutPath)
		free(cOutPath)
	case C.NFD_CANCEL:
		// empty
	default:
		err = getError()
	}

	free(cDefaultPath)
	free(cFilterList)
	return
}

func OpenDialogMultiple(filterList, defaultPath string) (res []string, err error) {
	cFilterList := cstring(filterList)
	cDefaultPath := cstring(defaultPath)

	var cOutPaths C.nfdpathset_t
	switch C.NFD_OpenDialogMultiple(cFilterList, cDefaultPath, &cOutPaths) {
	case C.NFD_OKAY:
		var i _Ctype_ulonglong
		for i = 0; i < C.NFD_PathSet_GetCount(&cOutPaths); i++ {
			res = append(res, C.GoString(C.NFD_PathSet_GetPath(&cOutPaths, i)))
		}
		C.NFD_PathSet_Free(&cOutPaths)
	case C.NFD_CANCEL:
		// empty
	default:
		err = getError()
	}

	free(cDefaultPath)
	free(cFilterList)
	return
}

func SaveDialog(filterList, defaultPath string) (res string, err error) {
	cFilterList := cstring(filterList)
	cDefaultPath := cstring(defaultPath)

	var cOutPath *C.char
	switch C.NFD_SaveDialog(cFilterList, cDefaultPath, &cOutPath) {
	case C.NFD_OKAY:
		res = C.GoString(cOutPath)
		free(cOutPath)
	case C.NFD_CANCEL:
		// empty
	default:
		err = getError()
	}

	free(cDefaultPath)
	free(cFilterList)
	return
}

func PickFolder(defaultPath string) (res string, err error) {
	cDefaultPath := cstring(defaultPath)

	var cOutPath *C.char
	switch C.NFD_PickFolder(cDefaultPath, &cOutPath) {
	case C.NFD_OKAY:
		res = C.GoString(cOutPath)
		free(cOutPath)
	case C.NFD_CANCEL:
		// empty
	default:
		err = getError()
	}

	free(cDefaultPath)
	return
}

func cstring(str string) *C.char {
	if str == "" {
		return nil
	}
	return C.CString(str)
}

func free(str *C.char) {
	C.free(unsafe.Pointer(str))
}

func getError() error {
	return errors.New(C.GoString(C.NFD_GetError()))
}
