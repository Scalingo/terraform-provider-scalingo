package scalingo

func uintAddr(i uint) *uint {
	return &i
}

func stringAddr(i string) *string {
	return &i
}

func boolAddr(i bool) *bool {
	return &i
}
