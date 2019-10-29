package tempstorage

import (
	"AID/solution/helper"
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"path"

	log "github.com/sirupsen/logrus"
)

// Read file lines and put in ch one by one
// Cleans remove files at end to save storage
func (ts *TempStorage) fileConsumer(ctx context.Context, parentPath string, info os.FileInfo) (<-chan string, error) {
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

	var ch chan string
	if ts.chanBuffSize > 1 {
		ch = make(chan string, ts.chanBuffSize)
	} else {
		ch = make(chan string)
	}

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

		reader := bufio.NewReader(file)
		var line string
		for {
			line, err = helper.GetNextLine(reader)
			if err != nil {
				if err == io.EOF {
					break
				}
				log.Errorf("error in reading file %s: %v", filePath, err)
			}
			select {
			case <-ctx.Done():
				log.Warningf("%s reading process is stopped before it finish", filePath)
				return

			case ch <- line:
			}
		}
	}()

	return ch, nil
}

// GetNextReadChs return read channels for the next up to k available
// files from read level directory
// ctx is context
// n is the number of files to read
// chs is slice of string channels, each element of slice is a channel that will
// have strings which are lines of a file in read directory
func (ts *TempStorage) GetNextReadChs(ctx context.Context, n int) (chs []<-chan string, err error) {
	infos, err := ts.readDirFile.Readdir(n)
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
		ch, fcErr := ts.fileConsumer(ctx, ts.readDirPath, info)
		if fcErr != nil {
			return nil, fcErr
		}
		chs = append(chs, ch)
	}

	return chs, nil
}
