package inputserializer

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// FileSerializer implements serializing input in file(s)
type FileSerializer struct {
	path string
}

// NewFileSerializer creates new FileSerializer entity to serialized file(s) located at upder path
func NewFileSerializer(path string) *FileSerializer {
	return &FileSerializer{path: path}
}

// GetSeralizerCh creates single reader to read content of all files in input directory
// params
// root input directory root path
// returns
// res a read-only string channel, one string for each line in input files
// err error
func (f *FileSerializer) GetSeralizerCh(ctx context.Context) (<-chan string, error) {

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
		err := filepath.Walk(f.path, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				fmt.Printf("Error in reading %s: %v\n", path, err)
				return nil // Don't stop processing next files
			}

			if !info.Mode().IsRegular() {
				// fmt.Printf("%s is not a regular file, is not readed\n", path)
				return nil // Don't stop processing next files
			}

			file, err := os.OpenFile(path, os.O_RDONLY, 0666)
			if err != nil {
				fmt.Printf("Error in opening %s: %v\n", path, err)
				return nil // Don't stop processing next files
			}
			defer file.Close()

			scanner := bufio.NewScanner(file)
			for scanner.Scan() {
				select {
				case <-ctx.Done():
					return io.EOF // Return error (EOF) to stop filepath walk from processing next files

				case ch <- scanner.Text():

				}
			}

			return nil
		})
		if err != nil {
			fmt.Println(err)
		}
		close(ch)
	}()

	return ch, nil
}
