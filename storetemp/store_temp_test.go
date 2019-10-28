package storetemp

import (
	"AID/solution/helper"
	"context"
	"io/ioutil"
	"os"
	"path"
	"strconv"
	"testing"
	"time"
)

func TestGetTempLevelPath(t *testing.T) {
	st, err := NewStoreTemp("./testData")
	if err != nil {
		t.Error(err)
		return
	}
	_, err = st.getTempLevelPath(-1)
	if err == nil {
		t.Error("getTempLevelPath does not return error on negative input")
		return
	}

	_, err = st.getTempLevelPath(-13232)
	if err == nil {
		t.Error("getTempLevelPath does not return error on negative input")
		return
	}

	tempPath, err := st.getTempLevelPath(0)
	if err != nil {
		t.Errorf("getTempLevelPath returnx error on valid input: %v", err)
		return
	}
	expectedPath := path.Join("testData", "0")

	if tempPath != expectedPath {
		t.Errorf("getTempLevelPath returns %s, expected %s", tempPath, expectedPath)
	}

	tempPath, err = st.getTempLevelPath(322)
	if err != nil {
		t.Errorf("getTempLevelPath returnx error on valid input: %v", err)
		return
	}
	expectedPath = path.Join("testData", "322")

	if tempPath != expectedPath {
		t.Errorf("getTempLevelPath returns %s, expected %s", tempPath, expectedPath)
	}
}

func TestNewStoreTemp(t *testing.T) {
	// Test with invalid path
	_, err := NewStoreTemp("")
	if err == nil {
		t.Error("NewStoreTemp should return error on empty path")
		return
	}

	tempPath := "notExistsDir"
	_, err = NewStoreTemp(tempPath)
	if err == nil {
		t.Errorf("NewStoreTemp should return error on invalid path %s", tempPath)
		return
	}

	tempPath = "asd sdfsd32 2"
	_, err = NewStoreTemp(tempPath)
	if err == nil {
		t.Errorf("NewStoreTemp should return error on invalid path %s", tempPath)
		return
	}

	tempPath = "testData"
	st, err := NewStoreTemp(tempPath)
	if err != nil {
		t.Error(err)
		return
	}
	if st.storeLevel != 0 {
		t.Errorf("initial value of storeLevel should be 0, but is %d", st.storeLevel)
		return
	}

	if st.readLevel != -1 {
		t.Errorf("initial value of readLevel should be -1, but is %d", st.readLevel)
		return
	}

	levelZeroDir := path.Join("testData", "0")
	isWritable, err := helper.IsWritableDir(levelZeroDir)
	if err != nil {
		t.Error(err)
	}
	if !isWritable {
		t.Errorf("level zero store directory %s should be a writable directory", levelZeroDir)
	}
}
func TestStoreTemp_SetupNextLevel(t *testing.T) {
	st, err := NewStoreTemp("testData")
	if err != nil {
		t.Error(err)
		return
	}

	err = st.SetupNextLevel()
	if err != nil {
		t.Error(err)
		return
	}

	levelZeroDir := path.Join("testData", "0")
	levelOneDir := path.Join("testData", "1")
	levelTwoDir := path.Join("testData", "2")

	isWritable, err := helper.IsWritableDir(levelZeroDir)
	if err != nil {
		t.Error(err)
		return
	}
	if !isWritable {
		t.Errorf("Level zero dir %s should exists and writable", levelZeroDir)
	}

	isWritable, err = helper.IsWritableDir(levelOneDir)
	if err != nil {
		t.Error(err)
		return
	}
	if !isWritable {
		t.Errorf("Level one dir %s should exists and writable", levelOneDir)
	}

	// Go one step further
	err = st.SetupNextLevel()
	if err != nil {
		t.Error(err)
		return
	}

	_, err = helper.IsWritableDir(levelZeroDir)
	if err == nil {
		t.Errorf("level zero directory should be cleaned after second step")
		return
	}
	if !os.IsNotExist(err) {
		t.Errorf("should IsNotExist error returned for level zero directory after second step")
		return
	}

	isWritable, err = helper.IsWritableDir(levelOneDir)
	if err != nil {
		t.Error(err)
		return
	}
	if !isWritable {
		t.Errorf("Level one dir %s should exists and writable", levelOneDir)
	}

	isWritable, err = helper.IsWritableDir(levelTwoDir)
	if err != nil {
		t.Error(err)
		return
	}
	if !isWritable {
		t.Errorf("Level two dir %s should exists and writable", levelTwoDir)
	}
}

func TestStoreTemp_GetNextReadChs(t *testing.T) {

	st, err := NewStoreTemp("testData")
	if err != nil {
		t.Error(err)
		return
	}

	numberOfFiles := 5
	k := 3

	// Write to file manually
	bytes := []byte("Hello\n")
	for i := 0; i < numberOfFiles; i++ {
		filePath := path.Join(st.storeDirPath, strconv.Itoa(i))
		err = ioutil.WriteFile(filePath, bytes, os.ModePerm)
		if err != nil {
			t.Error(err)
			return
		}
	}

	err = st.SetupNextLevel()
	if err != nil {
		t.Error(err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer func() {
		cancel()

		if st.readLevel >= 0 {
			err = helper.CleanDir(st.readDirPath)
			if err != nil {
				t.Logf("error in cleaning tempdir: %v", err)
			}
		}
		err = helper.CleanDir(st.storeDirPath)
		if err != nil {
			t.Logf("error in cleaning tempdir: %v", err)
		}
	}()

	chs, err := st.GetNextReadChs(ctx, k)
	if err != nil {
		t.Error(err)
		return
	}

	if len(chs) != k {
		t.Errorf("number of channels returned are %d, but should be %d", len(chs), numberOfFiles)
		return
	}

	for i, ch := range chs {
		var counter = 0
		notFinished := true
		var s string
		for notFinished {
			select {
			case <-ctx.Done():
				t.Errorf("reading from file %d took long time", i)
				return
			case s, notFinished = <-ch:
				if notFinished {
					counter++
					if s != "Hello" {
						t.Errorf("line content is \"%s\", but it should be \"Hello\"", s)
					}
				}
			}
		}
		if counter != 1 {
			t.Errorf("read %d lines from %d, which should be 1", counter, i)
			return
		}
	}

}

func TestStoreTemp_GetNextStoreCh(t *testing.T) {
	st, err := NewStoreTemp("testData")
	if err != nil {
		t.Error(err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer func() {
		cancel()

		if st.readLevel >= 0 {
			err = helper.CleanDir(st.readDirPath)
			if err != nil {
				t.Logf("error in cleaning tempdir: %v", err)
			}
		}
		err = helper.CleanDir(st.storeDirPath)
		if err != nil {
			t.Logf("error in cleaning tempdir: %v", err)
		}
	}()

	ch, err := st.GetNextStoreCh(ctx)
	if err != nil {
		t.Error(err)
		return
	}

	numberOfLines := 5
	k := 3

	for i := 0; i < numberOfLines; i++ {
		select {
		case ch <- "Hello":
		case <-ctx.Done():
			t.Error("write to file took long time")
			return
		}
	}
	close(ch)

	err = st.SetupNextLevel()
	if err != nil {
		t.Error(err)
		return
	}

	chs, err := st.GetNextReadChs(ctx, k)
	if err != nil {
		t.Error(err)
		return
	}

	if len(chs) != 1 {
		t.Errorf("number of read channels should be 1, but is %d", len(chs))
		return
	}

	for i, ch := range chs {
		var counter = 0
		notFinished := true
		var s string
		for notFinished {
			select {
			case <-ctx.Done():
				t.Errorf("reading from file %d took long time", i)
				return
			case s, notFinished = <-ch:
				if notFinished {
					counter++
					if s != "Hello" {
						t.Errorf("line content is \"%s\", but it should be \"Hello\"", s)
					}
				}
			}
		}
		if counter != numberOfLines {
			t.Errorf("read %d lines from %d, which should be %d", counter, i, numberOfLines)
			return
		}
	}
}

func TestStoreTemp_GetNextReadChs2(t *testing.T) {
	st, err := NewStoreTemp("testData")
	if err != nil {
		t.Error(err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer func() {
		cancel()

		if st.readLevel >= 0 {
			err = helper.CleanDir(st.readDirPath)
			if err != nil {
				t.Logf("error in cleaning tempdir: %v", err)
			}
		}
		err = helper.CleanDir(st.storeDirPath)
		if err != nil {
			t.Logf("error in cleaning tempdir: %v", err)
		}
	}()

	numberOfFiles := 10
	k := 3

	var ch chan<- string
	for i := 0; i < numberOfFiles; i++ {
		ch, err = st.GetNextStoreCh(ctx)
		if err != nil {
			t.Error(err)
			return
		}

		ch <- "Hello"

		close(ch)
	}

	err = st.SetupNextLevel()
	if err != nil {
		t.Error(err)
		return
	}

	remainingFiles := numberOfFiles
	var chs []<-chan string
	for remainingFiles > 0 {
		chs, err = st.GetNextReadChs(ctx, k)
		if err != nil {
			t.Error(err)
			return
		}

		var min int

		if k < remainingFiles {
			min = k
		} else {
			min = remainingFiles
		}

		if len(chs) != min {
			t.Errorf("length of returned channel is %d, but should be %d", len(chs), min)
			return
		}

		remainingFiles -= min
		cancel()

		ctx, cancel = context.WithTimeout(context.Background(), 2*time.Second)
	}

	chs, err = st.GetNextReadChs(ctx, k)
	if err != nil {
		t.Error(err)
		return
	}
	if len(chs) != 0 {
		t.Errorf("len of chs should is %d, but should be zero", len(chs))
	}
}
