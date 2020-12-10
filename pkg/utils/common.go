package utils

// IndexOf to find the index of a string in a string array
func IndexOf(s []string, str string) int {
	for i, v := range s {
		if v == str {
			return i
		}
	}
	return -1
}
