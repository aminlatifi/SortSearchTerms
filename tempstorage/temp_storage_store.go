package tempstorage

import (
	"bufio"
	"context"
	"os"
	"path"
	"strconv"
	"sync"

	log "github.com/sirupsen/logrus"
)

// GetNextStoreCh return write channel for next file in store level directory
// strings are put in chan will be written as lines in the file
// caller is responsible for closing the chan, after that file writer is closed too
func (ts *TempStorage) GetNextStoreCh(ctx context.Context, wg *sync.WaitGroup) (chan<- string, error) {
	fileName := strconv.Itoa(ts.storeFileCounter)
	ts.storeFileCounter++

	filePath := path.Join(ts.storeDirPath, fileName)

	file, err := os.Create(filePath)
	if err != nil {
		return nil, err
	}

	var ch chan string
	if ts.chanBuffSize > 1 {
		ch = make(chan string, ts.chanBuffSize)
	} else {
		ch = make(chan string)
	}

	go func(ch <-chan string, file *os.File) {
		defer wg.Done()

		writer := bufio.NewWriter(file)

		for {
			select {
			case s, ok := <-ch:
				if !ok {
					err = writer.Flush()
					if err != nil {
						log.Errorf("error on flushing data on %s: %v", filePath, err)
					}
					err = file.Close()
					if err != nil {
						log.Errorf("error in closing %s: %v", filePath, err)
					}
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

	}(ch, file)

	return ch, nil
}
