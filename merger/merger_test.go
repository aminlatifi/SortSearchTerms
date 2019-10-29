package merger

import (
	"AID/solution/helper"
	"AID/solution/tempstorage"
	"bufio"
	"context"
	"io"
	"os"
	"sort"
	"sync"
	"testing"
)

func TestStartMerge(t *testing.T) {
	ctx, _ := context.WithCancel(context.Background())

	ts, err := tempstorage.NewTempStorage("testData")
	if err != nil {
		t.Error(err)
		return
	}
	// We will clean ts at the end

	sampleData := []string{
		"aaa",
		"zzz",
		"bbb",
		"kkk",
		"eee",
		"jjj",
		"yyy",
		"ccc",
		"hhh",
		"fff",
		"qqq",
		"aaa",
		"vvv",
		"eee",
		"yyy",
		"dfs",
		"ere",
		"viu",
		"mmn",
		"mmm",
		"ppp",
		"qqq",
		"red",
		"rwe",
		"vfs",
		"pwj",
	}

	k := 4 // Bundle size

	var ch chan<- string
	bundle := make([]string, 0, k)
	var wg sync.WaitGroup
	for i := 0; i < len(sampleData); i++ {
		bundle = append(bundle, sampleData[i])
		if len(bundle) == k || i == len(sampleData)-1 {
			ch, err = ts.GetNextStoreCh(ctx, &wg)
			if err != nil {
				t.Error(err)
				return
			}
			wg.Add(1)

			sort.Strings(bundle)
			for _, v := range bundle {
				ch <- v
			}
			close(ch)
			bundle = bundle[:0] // Make bundle clear
		}
	}

	wg.Wait()

	outputPath := "testData/out.txt"

	err = StartMerge(ctx, ts, outputPath, k)

	hasSingle, _ := ts.HasSingleStoredFile()
	if !hasSingle {
		t.Error("after running merge should only one file be remained at store directory")
	}

	sort.Strings(sampleData) // Expected output

	file, err := os.Open(outputPath)
	if err != nil {
		t.Error(err)
		return
	}
	defer func() {
		err = file.Close()
		if err != nil {
			t.Error(err)
		}

		err = ts.Clean()
		if err != nil {
			t.Error(err)
		}
	}()

	reader := bufio.NewReader(file)
	var line string
	for i := 0; i < len(sampleData); i++ {
		line, err = helper.GetNextLine(reader)
		if err != nil {
			t.Error(err)
			return
		}
		if line != sampleData[i] {
			t.Errorf(
				"line %d of output is \"%s\", but should be \"%s\"",
				i, line, sampleData[i])
			return
		}
	}

	_, err = helper.GetNextLine(reader)
	if err != io.EOF {
		if err != nil {
			t.Error(err)
			return
		}
		t.Error("output data has more data than expected")
	}

}
