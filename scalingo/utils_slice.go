package scalingo

func Contains[T comparable](s []T, v T) bool {
	for _, c := range s {
		if c == v {
			return true
		}
	}

	return false
}
