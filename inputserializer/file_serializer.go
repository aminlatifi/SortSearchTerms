package inputserializer

import (
	"AID/solution/helper"
	"bufio"
	"context"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io"
	"os"
	"path/filepath"
)

// DirSerializer implements serializing input file(s) under directory
type DirSerializer struct {
	path string
}

// NewDirSerializer creates new DirSerializer entity to serialized file(s) located under path directory
func NewDirSerializer(path string) *DirSerializer {
	return &DirSerializer{path: path}
}

// GetSerializerCh creates single reader to read content of all files in input directory
// params
// root input directory root path
// returns
// res a read-only string channel, one string for each line in input files
// err error
func (f *DirSerializer) GetSerializerCh(ctx context.Context) (<-chan string, error) {

	// Check whether path exists!
	fileInfo, err := os.Stat(f.path)
	if err != nil {
		return nil, err
	}

	if !fileInfo.IsDir() {
		err = fmt.Errorf("%s is not directory", f.path)
		return nil, err
	}

	ch := make(chan string)

	go func() {
		defer close(ch)
		err := filepath.Walk(f.path, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				log.Warningf("Error in reading %s: %v", path, err)
				return nil // Don't stop processing next files
			}

			if info.IsDir() {
				// Their contents will be processed
				log.Infof("Content of files under directory %s will be serialized", path)
				return nil
			}

			if !info.Mode().IsRegular() {
				log.Warningf("%s is not a regular file, is not reade\n", path)
				return nil // Don't stop processing next files
			}

			file, err := os.Open(path)
			if err != nil {
				log.Warningf("Error in opening %s: %v", path, err)
				return nil // Don't stop processing next files
			}
			defer func() {
				err = file.Close()
				if err != nil {
					log.Errorf("error in closding fil %s: %v", path, err)
				}
			}()

			log.Debugf("Serialize content of: %s", path)

			reader := bufio.NewReader(file)
			var line string

			for {
				line, err = helper.GetNextLine(reader)
				if err != nil {
					if err == io.EOF {
						break
					}
					log.Errorf("error in reading file %s: %v", path, err)
				}
				select {
				case <-ctx.Done():
					return io.EOF // Return error (EOF) to stop filepath walk from processing next files

				case ch <- line:
				}
			}

			return nil
		})
		if err != nil {
			fmt.Println(err)
		}
	}()

	return ch, nil
}
