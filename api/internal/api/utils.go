package api

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
			filtered = append(filtered, elems[i])
		}
	}

	return filtered
}
