package inputserializer

import (
	"bufio"
	"context"
	"flag"
	"os"
	"path/filepath"
	"testing"

	"AID/solution/util"
)

var update = flag.Bool("update", false, "update .golden files")

func TestSampleInput(t *testing.T) {
	inputPath := filepath.Join("testData", "input")
	expctedPath := filepath.Join("testData", "output.golden")
	fileSerializer := NewFileSerializer(inputPath)
	ch, err := fileSerializer.GetSeralizerCh(context.Background())

	if err != nil {
		t.Error(err)
		return
	}

	result := []string{}
	for s := range ch {
		result = append(result, s)
	}

	if *update {
		err = util.WriteSliceToFile(expctedPath, result)
		if err != nil {
			t.Error(err)
			return
		}
	}

	file, err := os.OpenFile(expctedPath, os.O_RDONLY, 0666)
	if err != nil {
		t.Error(err)
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for i, s := range result {
		if scanner.Scan() {
			e := scanner.Text()
			if s != e {
				t.Errorf("Mismatch on %s, line %d: expected %s, but is %s", expctedPath, i, e, s)
			}
		} else { // Expected data is finished soon (input has more result than expected)
			t.Errorf("Result from %s is more than expected %s", inputPath, expctedPath)
			return
		}
	}

	// Check whether expected has more data
	if scanner.Scan() {
		r := scanner.Text()

		if len(r) > 0 {
			t.Errorf("Result from %s is less than expected %s", inputPath, expctedPath)
		}
	}
}
