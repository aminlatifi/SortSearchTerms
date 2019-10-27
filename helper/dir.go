// +build !windows

package helper

import (
	"os"
	"syscall"

	log "github.com/sirupsen/logrus"
)

// IsWritable checks whether path is a writable directory
func IsWritable(path string) (isWritable bool, err error) {
	isWritable = false
	info, err := os.Stat(path)
	if err != nil {
		log.Warningf("Path %s doesn't exist", path)
		return
	}

	err = nil
	if !info.IsDir() {
		log.Warningf("Path %s isn't a directory", path)
		return
	}

	// Check if the user bit is enabled in file permission
	if info.Mode().Perm()&(1<<(uint(7))) == 0 {
		log.Warningf("Write permission bit is not set on %s for user", path)
		return
	}

	var stat syscall.Stat_t
	if err = syscall.Stat(path, &stat); err != nil {
		log.Warningf("Unable to get stat on %s", path)
		return
	}

	err = nil
	if uint32(os.Geteuid()) != stat.Uid {
		isWritable = false
		log.Warningf("User doesn't have permission to write to %s", path)
		return
	}

	isWritable = true
	return
}
