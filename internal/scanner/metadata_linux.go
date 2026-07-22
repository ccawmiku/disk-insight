//go:build linux

package scanner

import (
	"fmt"
	"os"
	"strings"
	"syscall"
)

func platformMetadata(_ string, info os.FileInfo) (*int64, string) {
	stat, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		return nil, ""
	}
	allocated := stat.Blocks * 512
	return &allocated, fmt.Sprintf("%d:%d", stat.Dev, stat.Ino)
}

func platformHidden(_ string, name string, _ os.FileInfo) bool {
	return strings.HasPrefix(name, ".")
}
