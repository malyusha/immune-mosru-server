package fs

import (
	"os"
	"syscall"
)

func FileExists(path string) bool {
	stat, err := os.Stat(path)
	if err != nil {
		return !os.IsNotExist(err) && err != syscall.ENOENT
	}

	return stat != nil
}
