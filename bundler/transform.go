package bundler

import "sort"

// SortTransform sort bundle by quick sort algorithm implemented by sort package
func SortTransform(input []string) {
	sort.Strings(input)
}
