package handler

import "strings"

func pathHasSuffix(path string, suffixes []string) bool {
	for _, suffix := range suffixes {
		if strings.HasSuffix(path, suffix) {
			return true
		}
	}

	return false
}

func pathContainsString(path string, subStrings []string) bool {
	for _, subString := range subStrings {
		if strings.Contains(path, subString) {
			return true
		}
	}

	return false
}
