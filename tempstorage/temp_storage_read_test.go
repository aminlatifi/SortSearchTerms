package tempstorage

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

func TestTempStorage_GetNextReadChs(t *testing.T) {

	st, err := NewTempStorage("testData")
	if err != nil {
		t.Error(err)
		return
	}

	if st.HasFilesToRead() {
		t.Error("TempStorage should not have files to read initially")
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
	if !st.HasFilesToRead() {
		t.Error("TempStorage should have files to read")
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

func TestTempStorage_GetNextReadChs2(t *testing.T) {
	st, err := NewTempStorage("testData")
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
