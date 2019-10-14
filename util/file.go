package util

import (
	"bufio"
	"os"
)

// WriteSliceToFile fill the file located at path with one line per each string in source
func WriteSliceToFile(path string, source []string) error {

	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	for _, s := range source {
		_, err = writer.WriteString(s + "\n")
		if err != nil {
			return err
		}
	}

	err = writer.Flush()
	if err != nil {
		return err
	}

	return nil
}
