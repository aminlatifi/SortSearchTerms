package tempstorage

import (
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

		// TODO: migrate to bufio
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
func (ts *TempStorage) GetNextReadChs(ctx context.Context, k int) (chs []<-chan string, err error) {
	infos, err := ts.readDirFile.Readdir(k)
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
		ch, fcErr := fileConsumer(ctx, ts.readDirPath, info)
		if fcErr != nil {
			return nil, fcErr
		}
		chs = append(chs, ch)
	}

	return chs, nil
}
