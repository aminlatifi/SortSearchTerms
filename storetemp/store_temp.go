package storetmp

import (
	"fmt"
	"os"
)

// StoreTemp store temporary files data structure
type StoreTemp struct {
	path                  string
	readLevel, storeLevel int
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
	return &StoreTemp{
		path: path,
	}, nil
}
