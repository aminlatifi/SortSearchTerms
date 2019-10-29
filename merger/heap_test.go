package merger

import (
	"fmt"
	"testing"
)

func padNumberWithZero(value int) string {
	return fmt.Sprintf("%05d", value)
}
func TestSourceHeap(t *testing.T) {
	numberOfChannels := 4000

	chs := make([]<-chan string, 0, numberOfChannels)

	for i := 0; i < numberOfChannels; i++ {
		ch := make(chan string)
		chs = append(chs, ch)
		go func(i int) {
			ch <- "Hello " + padNumberWithZero(i)
			close(ch)
		}(i)
	}

	sh := newSourceHeap(chs)

	if sh.Len() != numberOfChannels {
		t.Errorf("heap length is %d, but should be %d", sh.Len(), numberOfChannels)
	}

	for i := 0; i < numberOfChannels; i++ {
		head, err := sh.getHead()
		if err != nil {
			t.Error(err)
			return
		}

		if head != "Hello "+padNumberWithZero(i) {
			t.Errorf("head is \"%s\", but should be \"%s\"", head, "Hello "+padNumberWithZero(i))
			return
		}

		_ = sh.updateHead() // We are sure there would not be error
	}

	if sh.Len() != 0 {
		t.Error("heap should be empty, but is not")
	}
}

func TestSourceHeap_PutReverse(t *testing.T) {
	numberOfChannels := 40001

	chs := make([]<-chan string, 0, numberOfChannels)

	for i := 0; i < numberOfChannels; i++ {
		ch := make(chan string)
		chs = append(chs, ch)
		go func(i int) {
			ch <- "Hello " + padNumberWithZero(numberOfChannels-i-1)
			close(ch)
		}(i)
	}

	sh := newSourceHeap(chs)

	if sh.Len() != numberOfChannels {
		t.Errorf("heap length is %d, but should be %d", sh.Len(), numberOfChannels)
	}

	for i := 0; i < numberOfChannels; i++ {
		head, err := sh.getHead()
		if err != nil {
			t.Error(err)
			return
		}

		if head != "Hello "+padNumberWithZero(i) {
			t.Errorf("head is \"%s\", but should be \"%s\"", head, "Hello "+padNumberWithZero(i))
			return
		}

		_ = sh.updateHead() // We are sure there would not be error
	}

	if sh.Len() != 0 {
		t.Error("heap should be empty, but is not")
	}
}

func TestSourceHeap_UpdateHead(t *testing.T) {
	numberOfChannels := 40
	numberOfStringPerChannel := 30

	chs := make([]<-chan string, 0, numberOfChannels)

	for i := 0; i < numberOfChannels; i++ {
		ch := make(chan string)
		chs = append(chs, ch)
		go func(i int) {
			for j := 0; j < numberOfStringPerChannel; j++ {
				ch <- "Hello" +
					" " + padNumberWithZero(numberOfChannels-i-1) +
					" " + padNumberWithZero(j)
			}
			close(ch)
		}(i)
	}

	sh := newSourceHeap(chs)

	if sh.Len() != numberOfChannels {
		t.Errorf("heap length is %d, but should be %d", sh.Len(), numberOfChannels)
	}

	for i := 0; i < numberOfChannels; i++ {
		for j := 0; j < numberOfStringPerChannel; j++ {
			head, err := sh.getHead()
			if err != nil {
				t.Error(err)
				return
			}

			expected := "Hello " + padNumberWithZero(i) + " " + padNumberWithZero(j)
			if head != expected {
				t.Errorf("head is \"%s\", but should be \"%s\"", head, expected)
				return
			}

			_ = sh.updateHead() // We are sure there would not be error
		}
	}

	if sh.Len() != 0 {
		t.Error("heap should be empty, but is not")
	}

}
