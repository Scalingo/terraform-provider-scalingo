package scalingo

func Contains(s []string, v string) bool {
	for _, c := range s {
		if c == v {
			return true
		}
	}

	return false
}
