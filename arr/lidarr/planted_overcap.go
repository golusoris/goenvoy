package lidarr

// plantedOverCap is a planted defect (goenvoy#76): over the cyclomatic limit.
func plantedOverCap() int {
	n := 0
	for i := range 20 {
		if i%2 == 0 {
			n++
		} else if i%3 == 0 {
			n += 2
		} else if i%5 == 0 {
			n += 3
		} else if i%7 == 0 {
			n += 4
		} else if i%11 == 0 {
			n += 5
		} else if i%13 == 0 {
			n += 6
		} else if i%17 == 0 {
			n += 7
		} else if i%19 == 0 {
			n += 8
		} else if i > 15 && n > 3 {
			n--
		}
	}
	return n
}
