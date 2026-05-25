package app

func indexByID[T any](items []T, currentID int64, idOf func(T) int64) int {
	for index, item := range items {
		if idOf(item) == currentID {
			return index
		}
	}

	return 0
}

func selectPrevious[T any](items []T, currentID int64, idOf func(T) int64) int64 {
	if len(items) == 0 {
		return 0
	}

	index := indexByID(items, currentID, idOf)
	if index <= 0 {
		index = len(items)
	}

	return idOf(items[index-1])
}

func selectNext[T any](items []T, currentID int64, idOf func(T) int64) int64 {
	if len(items) == 0 {
		return 0
	}

	index := indexByID(items, currentID, idOf)
	return idOf(items[(index+1)%len(items)])
}

func selectPreviousBounded[T any](items []T, currentID int64, idOf func(T) int64) int64 {
	if len(items) == 0 {
		return 0
	}

	index := indexByID(items, currentID, idOf)
	if index <= 0 {
		return idOf(items[0])
	}

	return idOf(items[index-1])
}

func selectNextBounded[T any](items []T, currentID int64, idOf func(T) int64) int64 {
	if len(items) == 0 {
		return 0
	}

	index := indexByID(items, currentID, idOf)
	if index >= len(items)-1 {
		return idOf(items[len(items)-1])
	}

	return idOf(items[index+1])
}

func ensureSelected[T any](items []T, currentID int64, idOf func(T) int64) int64 {
	if len(items) == 0 {
		return 0
	}
	for _, item := range items {
		if idOf(item) == currentID {
			return currentID
		}
	}

	return idOf(items[0])
}

func findByID[T any](items []T, currentID int64, idOf func(T) int64) (T, bool) {
	for _, item := range items {
		if idOf(item) == currentID {
			return item, true
		}
	}

	var zero T
	return zero, false
}
