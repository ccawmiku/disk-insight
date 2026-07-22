//go:build !linux && !windows

package scanner

import (
	"os"
	"strings"
)

func platformMetadata(_ string, _ os.FileInfo) (*int64, string) { return nil, "" }

func platformHidden(_ string, name string, _ os.FileInfo) bool {
	return strings.HasPrefix(name, ".")
}
