// +build !windows

package helper

import (
	"os"
	"testing"
)

func TestIsWritableReadOnly(t *testing.T) {
	dirPath := "./testData/testWritableReadOnly"

	if _, err := os.Stat(dirPath); !os.IsNotExist(err) {
		err := os.RemoveAll(dirPath)
		if err != nil {
			t.Error(err)
			return
		}
	}

	err := os.Mkdir(dirPath, 0555)
	if err != nil {
		t.Error(err)
		return
	}
	defer func() {
		err = os.RemoveAll(dirPath)
		if err != nil {
			t.Error(err)
		}
	}()

	isWritable, err := IsWritable(dirPath)

	if err != nil {
		t.Error(err)
		return
	}

	if isWritable {
		t.Errorf("Directory %s should not be writable", dirPath)
	}

}

func TestIsWritable(t *testing.T) {
	dirPath := "./testData/testWritable"

	if _, err := os.Stat(dirPath); !os.IsNotExist(err) {
		err := os.RemoveAll(dirPath)
		if err != nil {
			t.Error(err)
			return
		}
	}

	err := os.Mkdir(dirPath, 0777)
	if err != nil {
		t.Error(err)
		return
	}
	defer func() {
		err = os.RemoveAll(dirPath)
		if err != nil {
			t.Error(err)
		}
	}()

	isWritable, err := IsWritable(dirPath)

	if err != nil {
		t.Error(err)
		return
	}

	if !isWritable {
		t.Errorf("Directory %s should be writable", dirPath)
	}
}
