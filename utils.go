package main

// https://stackoverflow.com/a/15323988/1597274
func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}
