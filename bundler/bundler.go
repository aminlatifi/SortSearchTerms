package bundler

import (
	"context"
)

// Bundler bundler entity
type Bundler interface {
	AddTransformFunc(f TransformFunc)
	GetBundlerCh(context.Context, <-chan string) <-chan []string
}

// TransformFunc define transform logic
type TransformFunc = func([]string)

type bundler struct {
	k          int // bundler size
	transforms []TransformFunc
}

func (b *bundler) AddTransformFunc(f TransformFunc) {
	b.transforms = append(b.transforms, f)
}

func (b *bundler) GetBundlerCh(ctx context.Context, inCh <-chan string) <-chan []string {
	// TODO: implements
	ch := make(chan []string)

	go func() {
		defer close(ch)
		bundle := make([]string, 0, b.k)
		for {
			select {
			case <-ctx.Done():
				return
			case s, ok := <-inCh:
				if ok {
					bundle = append(bundle, s)
					if len(bundle) == b.k {
						for _, t := range b.transforms {
							t(bundle)
						}
						ch <- bundle
						bundle = make([]string, 0, b.k)
					}
				} else {
					for _, t := range b.transforms {
						t(bundle)
					}
					ch <- bundle
					return
				}
			}
		}
	}()

	return ch
}

// GetNewBundler creates new bundler entity which creates bundles of size k
func GetNewBundler(k int) Bundler {
	return &bundler{
		k,
		[]TransformFunc{},
	}
}
