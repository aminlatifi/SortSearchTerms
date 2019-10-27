package helper

import (
	"os"

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
		log.Warning("Write permission bit is not set on %s for user", path)
		return
	}

	isWritable = true
	return
}
