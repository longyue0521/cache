package cache

import (
	"context"
	"errors"
	"time"
)

var (
	ErrKeyNotFound = errors.New("cache: key not found")
)

type Cache interface {
	Get(ctx context.Context, key string) (any, error)
	Set(ctx context.Context, key string, val any, expiration time.Duration) error
	Del(ctx context.Context, key string) error
}

// type AnyValue struct {
// 	val any
// 	err error
// }
//
// func (a AnyValue) String() (string, error) {
// 	if a.err != nil {
// 		return "", a.err
// 	}
//
// 	str, ok := a.val.(string)
// 	if !ok {
// 		return "", errors.New("can't covert to string")
// 	}
//
// 	return str, nil
// }
