package scalingo

// keepIf deletes every element of slice for which the given test function return false.
// It returns the filtered slice.
func keepIf[S any](slice []S, test func(S) bool) []S {
	filteredSlice := []S{}

	for _, elem := range slice {
		if test(elem) {
			filteredSlice = append(filteredSlice, elem)
		}
	}
	return filteredSlice
}

func Contains(s []string, v string) bool {
	for _, c := range s {
		if c == v {
			return true
		}
	}

	return false
}
