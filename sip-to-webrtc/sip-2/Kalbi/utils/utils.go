package utils

// Contains returns bool if search exists within the array
func Contains(search string, list []string) bool {
	for index := range list {
		if list[index] == search {
			return true
		}
	}
	return false
}
