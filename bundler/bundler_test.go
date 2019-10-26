package bundler

import (
	"context"
	"sort"
	"strings"
	"testing"
	"time"
	"unicode"
)

type bundleValidator = func(t *testing.T, bundle []string) (isValid bool)

func runBundler(t *testing.T, k int, sampleInput []string, validator bundleValidator, transforms ...TransformFunc) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()

	inputCh := make(chan string)
	bundler := GetNewBundler(k)
	for _, t := range transforms {
		bundler.AddTransformFunc(t)
	}
	outputCh := bundler.GetBundlerCh(ctx, inputCh)

	if sampleInput == nil {
		sampleInput = []string{"one", "two", "three", "four", "five", "six", "seven", "eight", "nine", "ten", "eleven", "twelve"}
	}

	go func() {
		defer close(inputCh)
		for _, s := range sampleInput {
			select {
			case <-ctx.Done():
				t.Error("Timed out")
				return
			case inputCh <- s:
				t.Logf("Writed %s to inputCh\n", s)
			}
		}
	}()

	for {
		select {
		case <-ctx.Done():
			t.Error("Timed out")
			return
		case bundle, ok := <-outputCh:
			// Closed
			if !ok {
				return
			}

			t.Logf("got bundle: %v", bundle)

			isValid := validator(t, bundle)
			if !isValid {
				return
			}
		}
	}
}

func checkBundleSize(t *testing.T, k int) {
	var seenIncompleteBundle bool // bundle with len less than k is seen, it should be the last one
	validator := func(t *testing.T, bundle []string) bool {
		if seenIncompleteBundle {
			t.Error("There is some bundle after an incomplete bundle")
			return false
		}

		if len(bundle) > k {
			t.Errorf("Invalid bundle size %d (expected %d)", len(bundle), k)
			return false
		}

		if len(bundle) < k {
			seenIncompleteBundle = true
		}
		return true
	}

	runBundler(t, k, nil, validator)
}

func TestBundlerSize(t *testing.T) {
	checkBundleSize(t, 4)
}

func TestBundlerSize2(t *testing.T) {
	checkBundleSize(t, 5)
}

func TestTransformSort(t *testing.T) {
	sampleInput := []string{
		"zzz", "hhh", "ddd", "aaa",
		"aaa", "ddd", "bbb", "aba",
		"aaa", "bbb", "ccc", "ddd",
		"zzz", "aaa",
	}

	validator := func(t *testing.T, bundle []string) bool {
		if !sort.StringsAreSorted(bundle) {
			t.Errorf("bundle is not sorted: %v", bundle)
			return false
		}
		return true
	}
	runBundler(t, 4, sampleInput, validator, SortTransform)
}

func TestMultipleTransform(t *testing.T) {
	sampleInput := []string{
		"zzz", "hhh", "ddd", "aaa",
		"aaa", "ddd", "bbb", "aba",
		"aaa", "bbb", "ccc", "ddd",
		"zzz", "aaa",
	}

	toUpperTransform := func(input []string) {
		for i, s := range input {
			input[i] = strings.ToUpper(s)
		}
	}

	validator := func(t *testing.T, bundle []string) bool {
		if !sort.StringsAreSorted(bundle) {
			t.Errorf("bundle is not sorted: %v", bundle)
			return false
		}
		for _, s := range bundle {
			if strings.IndexFunc(s, unicode.IsLower) != -1 {
				t.Errorf("%s has lowercase character", s)
			}

		}
		return true
	}

	runBundler(t, 4, sampleInput, validator, SortTransform, toUpperTransform)
}
