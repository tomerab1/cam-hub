package utils

import (
	"context"
	"errors"
	"os"
	"os/signal"
)

var (
	ErrNoFound = errors.New("element was not found")
)

type Vec2D struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

type Pair[T, U any] struct {
	First  T
	Second U
}

func Zip[T, U any](s1 []T, s2 []U) []Pair[T, U] {
	minLength := len(s1)
	if len(s2) < minLength {
		minLength = len(s2)
	}

	zipped := make([]Pair[T, U], minLength)
	for i := 0; i < minLength; i++ {
		zipped[i] = Pair[T, U]{First: s1[i], Second: s2[i]}
	}

	return zipped
}

func CountElems[T any](elems []T, pred func(elem T) bool) int {
	count := 0
	for _, elem := range elems {
		if pred(elem) {
			count += 1
		}
	}

	return count
}

func FilterElems[T any](elems []T, allocSz int, pred func(idx int, elem T) bool) []T {
	filtered := make([]T, allocSz)

	for i, elem := range elems {
		if pred(i, elem) {
			filtered[i] = elem
		}
	}

	return filtered
}

func FindFirstIdx[T any](elems []T, pred func(idx int, elem T) bool) (int, error) {
	for i, elem := range elems {
		if pred(i, elem) {
			return i, nil
		}
	}

	return -1, ErrNoFound
}

func Map[T, U any](s []T, f func(T) U) []U {
	result := make([]U, len(s))
	for i, v := range s {
		result[i] = f(v)
	}

	return result
}

func GracefullShutdown(ctx context.Context, onShutdown func(), sigs ...os.Signal) (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(ctx)
	c := make(chan os.Signal, 1)

	signal.Notify(c, sigs...)
	go func() {
		<-c
		onShutdown()
		cancel()
	}()

	return ctx, cancel
}
