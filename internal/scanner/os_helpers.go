package scanner

import (
	"os"
)

func osStat(path string) (os.FileInfo, error) { return os.Stat(path) }
func osModeSymlink() os.FileMode              { return os.ModeSymlink }
