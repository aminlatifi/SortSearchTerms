package bundler

import (
	"reflect"
	"testing"
)

func TestTransformSort1(t *testing.T) {
	input := []string{
		"ddd",
		"aaaa",
		"bbbb",
		"cccc",
	}

	expected := []string{
		"aaaa",
		"bbbb",
		"cccc",
		"ddd",
	}

	SortTransform(input)

	if !reflect.DeepEqual(input, expected) {
		t.Errorf("Sort transform problem\nResult: %v\nExpected: %v\n", input, expected)
	}
}
