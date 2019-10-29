package helper

import (
	"bufio"
	"io"

	log "github.com/sirupsen/logrus"
	"os"
)

// WriteSliceToFile fill the file located at path with one line per each string in source
func WriteSliceToFile(path string, source []string) error {

	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer func() {
		closeErr := file.Close()
		if err != nil {
			err = closeErr
		}
	}()

	writer := bufio.NewWriter(file)
	for _, s := range source {
		_, err = writer.WriteString(s + "\n")
		if err != nil {
			return err
		}
	}

	err = writer.Flush()

	return err
}

// GetNextLine get next line from reader
// It handles big lines too
func GetNextLine(reader *bufio.Reader) (line string, err error) {

	isPrefix := true
	var chunk, buffer []byte
	for isPrefix {
		chunk, isPrefix, err = reader.ReadLine()
		if err != nil {
			if err == io.EOF {
				return
			}
			log.Error("error in reading", err)
			return
		}
		buffer = append(buffer, chunk...)
	}

	line = string(buffer)

	return
}
