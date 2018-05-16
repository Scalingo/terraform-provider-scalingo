package scalingo

type MapDiffRes struct {
	Added    []string
	Deleted  []string
	Modified []string
}

func MapDiff(a map[string]interface{}, b map[string]interface{}) MapDiffRes {
	res := MapDiffRes{}
	for key, value := range a {
		newValue, found := b[key]
		if !found {
			res.Deleted = append(res.Deleted, key)
		} else if newValue.(string) != value.(string) {
			res.Modified = append(res.Modified, key)
		}
	}

	for key, _ := range b {
		_, found := a[key]
		if !found {
			res.Added = append(res.Added, key)
		}
	}
	return res
}
