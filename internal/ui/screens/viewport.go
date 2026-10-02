package screens

const narrowLayoutWidth = 72

// visibleRowsByID keeps the selected row in a fixed-height table viewport.
// A stable ID survives sorting and refreshing the backing slice.
func visibleRowsByID[T any](rows []T, selectedID int64, capacity int, id func(T) int64) []T {
	if capacity <= 0 {
		return nil
	}
	if len(rows) <= capacity {
		return rows
	}
	selected := 0
	for index, row := range rows {
		if id(row) == selectedID {
			selected = index
			break
		}
	}
	start := selected - capacity + 1
	if start < 0 {
		start = 0
	}
	return rows[start : start+capacity]
}
