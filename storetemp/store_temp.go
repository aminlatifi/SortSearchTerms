package storetemp

import (
	"AID/solution/helper"
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"path"
	"strconv"

	log "github.com/sirupsen/logrus"
)

// StoreTemp store temporary files data structure
type StoreTemp struct {
	path                      string
	readLevel, storeLevel     int
	readDirPath, storeDirPath string // Store to generate once during opening files under directories
	readDirFile               *os.File
	storeFileCounter          int
}

// NewStoreTemp creates new StoreTemp module
func NewStoreTemp(path string) (*StoreTemp, error) {
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
	st := &StoreTemp{
		path:             path,
		readLevel:        -1,
		storeLevel:       0,
		storeFileCounter: 0,
	}
	// Create zero level (initial input) store directory
	levelPath, err := st.getTempLevelPath(st.storeLevel)
	if err != nil {
		return nil, err
	}
	st.storeDirPath = levelPath

	err = helper.MakeCleanDir(levelPath)
	if err != nil {
		return nil, err
	}

	log.Infof("StoreTemp ready to store at level %d: %s", st.storeLevel, st.storeDirPath)

	return st, nil

}

func (st *StoreTemp) getTempLevelPath(level int) (string, error) {
	if level < 0 {
		err := fmt.Errorf("level %d cannot be less than zero", level)
		return "", err
	}

	result := path.Join(st.path, strconv.Itoa(level))

	return result, nil
}

// SetupNextLevel moves Temporary Storage one step further
func (st *StoreTemp) SetupNextLevel() error {
	// Clean previous read directory
	if st.readLevel >= 0 {
		err := st.readDirFile.Close()
		if err != nil {
			return err
		}

		readPath, err := st.getTempLevelPath(st.readLevel)
		if err != nil {
			return err
		}

		err = helper.CleanDir(readPath)
		if err != nil {
			return err
		}
	}

	st.readLevel++
	st.storeLevel++

	// Initialize read at readLevel
	st.readDirPath = st.storeDirPath

	readDirFile, err := os.Open(st.readDirPath)

	log.Infof("StoreTemp ready to read at level %d: %s", st.readLevel, st.readDirPath)

	if err != nil {
		return err
	}
	st.readDirFile = readDirFile

	// Initialize store at storeLevel
	storePath, err := st.getTempLevelPath(st.storeLevel)
	if err != nil {
		return err
	}
	st.storeDirPath = storePath
	st.storeFileCounter = 0

	err = helper.MakeCleanDir(storePath)
	if err != nil {
		log.Errorf("error on creating store path %s: %v", storePath, err)
		return err
	}

	log.Infof("StoreTemp ready to store at level %d: %s", st.storeLevel, st.storeDirPath)

	return nil
}

// Read file lines and put in ch one by one
// Cleans remove files at end to save storage
func fileConsumer(ctx context.Context, parentPath string, info os.FileInfo) (<-chan string, error) {
	if info.IsDir() {
		// Their contents will be processed
		err := fmt.Errorf("no directory should be inside read directory of temporary storage: %s", info.Name())
		return nil, err
	}

	if !info.Mode().IsRegular() {
		err := fmt.Errorf("%s is not a regular file", info.Name())
		return nil, err // Don't stop processing next files
	}

	filePath := path.Join(parentPath, info.Name())
	file, err := os.Open(filePath)
	if err != nil {
		log.Errorf("Error in opening %s: %v", filePath, err)
		return nil, err
	}

	ch := make(chan string)

	go func() {
		defer close(ch)
		defer func() {
			err = file.Close()
			if err != nil {
				log.Errorf("error in closing %s: %v", filePath, err)
				return
			}

			err = os.Remove(filePath)
			if err != nil {
				log.Errorf("error in removing %s: %v", filePath, err)
			}
		}()

		log.Debugf("Serialize content of %s", filePath)

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			s := scanner.Text()
			log.Debugf("read from file %s: %s", filePath, s)
			select {
			case <-ctx.Done():
				log.Warningf("%s reading process is stopped before it finish", filePath)
				return
			case ch <- s:
			}
		}

	}()

	return ch, nil
}

// GetNextReadChs return read channels for the next up to k available
// files from read level directory
// ctx is context
// k is the number of files to read
// chs is slice of string channels, each element of slice is a channel that will
// have strings which are lines of a file in read directory
func (st *StoreTemp) GetNextReadChs(ctx context.Context, k int) (chs []<-chan string, err error) {
	infos, err := st.readDirFile.Readdir(k)
	// Reached the end
	if err == io.EOF {
		return nil, nil
	}
	if err != nil {
		log.Errorf("Error in Readdir: %v", err)
		return nil, err
	}

	chs = make([]<-chan string, 0, len(infos))

	for _, info := range infos {
		ch, fcErr := fileConsumer(ctx, st.readDirPath, info)
		if fcErr != nil {
			return nil, fcErr
		}
		chs = append(chs, ch)
	}

	return chs, nil
}

// GetNextStoreCh return write channel for next file in store level directory
// strings are put in chan will be written as lines in the file
// caller is responsible for closing the chan, after that file writer is closed too
func (st *StoreTemp) GetNextStoreCh(ctx context.Context) (chan<- string, error) {
	fileName := strconv.Itoa(st.storeFileCounter)
	st.storeFileCounter++

	filePath := path.Join(st.storeDirPath, fileName)

	file, err := os.Create(filePath)
	if err != nil {
		return nil, err
	}

	ch := make(chan string)

	go func() {
		writer := bufio.NewWriter(file)
		defer func() {
			err = writer.Flush()
			if err != nil {
				log.Errorf("error on flushing data on %s: %v", filePath, err)
			}
			err = file.Close()
			if err != nil {
				log.Errorf("error in closing %s: %v", filePath, err)
				return
			}
		}()

		for {
			select {
			case s, ok := <-ch:
				if !ok {
					return
				}
				_, err = writer.WriteString(s)
				if err == nil {
					err = writer.WriteByte('\n')
				}
				if err != nil {
					log.Errorf("error in writing to %s: %v", filePath, err)
				}
			case <-ctx.Done():
				log.Warningf("%s write process is stopped before it finish", filePath)
			}
		}

	}()

	return ch, nil
}
