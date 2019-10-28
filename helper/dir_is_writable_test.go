// +build !windows

package helper

import (
	"os"
	"testing"
)

func TestIsWritableReadOnly(t *testing.T) {
	dirPath := "./testData/testWritableReadOnly"

	err := CleanDir(dirPath)
	if err != nil {
		t.Error(err)
		return
	}

	err = os.Mkdir(dirPath, 0555)
	if err != nil {
		t.Error(err)
		return
	}
	defer func() {
		err = CleanDir(dirPath)
	}()

	isWritable, err := IsWritableDir(dirPath)

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

	err := MakeCleanDir(dirPath)
	if err != nil {
		t.Error(err)
		return
	}
	defer func() {
		err = CleanDir(dirPath)
		if err != nil {
			t.Error(err)
		}
	}()

	isWritable, err := IsWritableDir(dirPath)

	if err != nil {
		t.Error(err)
		return
	}

	if !isWritable {
		t.Errorf("Directory %s should be writable", dirPath)
	}
}
