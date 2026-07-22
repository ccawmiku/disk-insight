//go:build windows

package scanner

import (
	"os"
	"strings"
	"syscall"
)

func platformMetadata(_ string, _ os.FileInfo) (*int64, string) {
	return nil, ""
}

func platformHidden(path string, name string, _ os.FileInfo) bool {
	if strings.HasPrefix(name, ".") {
		return true
	}
	pathPtr, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		return false
	}
	attributes, err := syscall.GetFileAttributes(pathPtr)
	return err == nil && attributes&syscall.FILE_ATTRIBUTE_HIDDEN != 0
}
