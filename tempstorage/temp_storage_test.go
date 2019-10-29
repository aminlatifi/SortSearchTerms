package tempstorage

import (
	"AID/solution/helper"
	"os"
	"path"
	"testing"
)

func TestGetTempLevelPath(t *testing.T) {
	st, err := NewTempStorage("./testData")
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

func TestNewTempStorage(t *testing.T) {
	// Test with invalid path
	_, err := NewTempStorage("")
	if err == nil {
		t.Error("NewTempStorage should return error on empty path")
		return
	}

	tempPath := "notExistsDir"
	_, err = NewTempStorage(tempPath)
	if err == nil {
		t.Errorf("NewTempStorage should return error on invalid path %s", tempPath)
		return
	}

	tempPath = "asd sdfsd32 2"
	_, err = NewTempStorage(tempPath)
	if err == nil {
		t.Errorf("NewTempStorage should return error on invalid path %s", tempPath)
		return
	}

	tempPath = "testData"
	st, err := NewTempStorage(tempPath)
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
func TestTempStorage_SetupNextLevel(t *testing.T) {
	st, err := NewTempStorage("testData")
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
