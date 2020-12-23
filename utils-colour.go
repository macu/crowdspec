package main

import (
	"regexp"
)

var validColourPattern = regexp.MustCompile("^rgb\\((\\d{1,3}), (\\d{1,3}), (\\d{1,3})\\)$")

func isValidColour(colour string) bool {
	return validColourPattern.MatchString(colour)
}
