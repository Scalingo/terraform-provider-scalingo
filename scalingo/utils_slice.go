package scalingo

// keepIf deletes every element of slice for which the given test function return false.
// It returns the filtered slice.
func keepIf[S any](slice []S, test func(S) bool) []S {
	var filteredElem []S

	for _, s := range slice {
		if test(s) {
			filteredElem = append(filteredElem, s)
		}
	}
	return filteredElem
}

func Contains(s []string, v string) bool {
	for _, c := range s {
		if c == v {
			return true
		}
	}

	return false
}
