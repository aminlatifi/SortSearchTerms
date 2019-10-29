package merger

import (
	"container/heap"
	"fmt"
)

type sourceItem struct {
	ch    <-chan string // The channel to read the next value from
	value string        // The head string of file which is processed
}

// A sourceHeap implements heap.Interface and holds Items.
type sourceHeap []*sourceItem

// Len length of sourceHeap
func (sh sourceHeap) Len() int { return len(sh) }

// Less compare two source item
func (sh sourceHeap) Less(i, j int) bool {
	return sh[i].value < sh[j].value
}

// Swap swap two item in sourceHeap
func (sh sourceHeap) Swap(i, j int) {
	sh[i], sh[j] = sh[j], sh[i]
}

// Push add one sourceItem to sourceHeap
func (sh *sourceHeap) Push(x interface{}) {
	item := x.(*sourceItem)
	*sh = append(*sh, item)
}

// Pop get and remove one sourceItem from sourceHeap
func (sh *sourceHeap) Pop() interface{} {
	old := *sh
	n := len(old)
	item := old[n-1]
	old[n-1] = nil // avoid memory leak
	*sh = old[0 : n-1]
	return item
}

func (sh *sourceHeap) getHead() (head string, err error) {
	if sh.Len() < 1 {
		err = fmt.Errorf("heap is empty")
		return
	}

	head = (*sh)[0].value
	return
}

// updateHead updates sourceItem at head and fixes the heap after modification
func (sh *sourceHeap) updateHead() (err error) {
	if sh.Len() < 1 {
		err = fmt.Errorf("heap is empty")
		return
	}
	item := (*sh)[0]
	newValue, ok := <-item.ch
	if ok {
		item.value = newValue
		heap.Fix(sh, 0)
	} else {
		heap.Pop(sh)
	}

	return
}

// newSourceHeap creates and initializes Source Heap with read channels(chs)
func newSourceHeap(chs []<-chan string) *sourceHeap {
	sh := &sourceHeap{}
	// Initial filling underneath slice without initializing heap
	// to have O(k) complexity rather than O(k*log k) at inserting k elements
	for _, ch := range chs {
		item := &sourceItem{ch: ch, value: <-ch}
		// Just append to the underneath slice and needles to initialize heap yet
		sh.Push(item)
	}

	// Initialize heap
	heap.Init(sh)

	return sh
}
