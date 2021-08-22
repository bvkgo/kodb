package kodb

import (
	"context"
)

type Reader interface {
	Load(ctx context.Context, key string, ob interface{}) error
}

type Writer interface {
	Store(ctx context.Context, key string, ob interface{}) error
}

type Iterator interface {
	LoadNext(ctx context.Context, key *string, ob interface{}) error
}

type Finder interface {
	// FindByIndex returns zero or more objects through the iterator. Indexed
	// fields with non-zero value in the input object are used to select the
	// index keys.
	FindByIndex(ctx context.Context, partial interface{}, it Iterator) error
}
