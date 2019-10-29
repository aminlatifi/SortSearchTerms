package tempstorage

import (
	"AID/solution/helper"
	"fmt"
	"os"
	"path"
	"strconv"

	log "github.com/sirupsen/logrus"
)

// TempStorage store temporary files data structure
type TempStorage struct {
	path                      string   // path of root directory
	readLevel                 int      // level from which data would read
	storeLevel                int      // level to which data would write
	readDirPath, storeDirPath string   // keep read and store paths to generate once and use multiple times
	readDirFile               *os.File // read directory os.File to go through files in read directory
	storeFileCounter          int      // number of files has been created in store directory, used to create next ones
}

// NewTempStorage creates new TempStorage module
func NewTempStorage(path string) (*TempStorage, error) {
	file, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	if !file.IsDir() {
		err = fmt.Errorf("%s should be a directory", path)
		return nil, err
	}
	isWritable, err := helper.IsWritableDir(path)
	if err != nil {
		return nil, err
	}
	if !isWritable {
		err = fmt.Errorf("%s should be a writable directory", path)
		return nil, err
	}
	ts := &TempStorage{
		path:             path,
		readLevel:        -1,
		storeLevel:       0,
		storeFileCounter: 0,
	}
	// Create zero level (initial input) store directory
	levelPath, err := ts.getTempLevelPath(ts.storeLevel)
	if err != nil {
		return nil, err
	}
	ts.storeDirPath = levelPath

	err = helper.MakeCleanDir(levelPath)
	if err != nil {
		return nil, err
	}

	log.Infof("TempStorage ready to store at level %d: %s", ts.storeLevel, ts.storeDirPath)

	return ts, nil

}

func (ts *TempStorage) getTempLevelPath(level int) (string, error) {
	if level < 0 {
		err := fmt.Errorf("level %d cannot be less than zero", level)
		return "", err
	}

	result := path.Join(ts.path, strconv.Itoa(level))

	return result, nil
}

// SetupNextLevel moves Temporary Storage one step further
func (ts *TempStorage) SetupNextLevel() error {
	// Clean previous read directory
	if ts.readLevel >= 0 {
		err := ts.readDirFile.Close()
		if err != nil {
			return err
		}

		readPath, err := ts.getTempLevelPath(ts.readLevel)
		if err != nil {
			return err
		}

		err = helper.CleanDir(readPath)
		if err != nil {
			return err
		}
	}

	ts.readLevel++
	ts.storeLevel++

	// Initialize read at readLevel
	ts.readDirPath = ts.storeDirPath

	readDirFile, err := os.Open(ts.readDirPath)

	log.Infof("TempStorage ready to read at level %d: %s", ts.readLevel, ts.readDirPath)

	if err != nil {
		return err
	}
	ts.readDirFile = readDirFile

	// Initialize store at storeLevel
	storePath, err := ts.getTempLevelPath(ts.storeLevel)
	if err != nil {
		return err
	}
	ts.storeDirPath = storePath
	ts.storeFileCounter = 0

	err = helper.MakeCleanDir(storePath)
	if err != nil {
		log.Errorf("error on creating store path %s: %v", storePath, err)
		return err
	}

	log.Infof("TempStorage ready to store at level %d: %s", ts.storeLevel, ts.storeDirPath)

	return nil
}
