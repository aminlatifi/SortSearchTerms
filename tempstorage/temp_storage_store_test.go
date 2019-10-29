package tempstorage

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestTempStorage_GetNextStoreCh(t *testing.T) {
	ts, err := NewTempStorage("testData")
	if err != nil {
		t.Error(err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer func() {
		cancel()
		err = ts.Clean()
		if err != nil {
			t.Error(err)
		}
	}()

	var wg sync.WaitGroup
	wg.Add(1)
	ch, err := ts.GetNextStoreCh(ctx, &wg)
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

	wg.Wait()

	err = ts.SetupNextLevel()
	if err != nil {
		t.Error(err)
		return
	}

	chs, err := ts.GetNextReadChs(ctx, k)
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
