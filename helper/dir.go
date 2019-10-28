package helper

import (
	log "github.com/sirupsen/logrus"
	"os"
)

func MakeDir(path string) error {
	err := os.Mkdir(path, os.ModePerm)
	if err != nil {
		log.Errorf("error on creating store path %s: %v", path, err)
		return err
	}
	return nil
}

func CleanDir(path string) error {
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		err := os.RemoveAll(path)
		if err != nil {
			log.Errorf("error on cleaning store path %s: %v", path, err)
			return err
		}
	}
	return nil
}

func MakeCleanDir(path string) error {
	err := CleanDir(path)
	if err != nil {
		return err
	}

	return MakeDir(path)
}
