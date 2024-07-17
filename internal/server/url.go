package server

import (
	"net/url"
	"strings"
)

func matchRoute(pattern, path string) bool {
	if pattern == "*" {
		return true
	}

	patternParts := strings.Split(pattern, "/")
	pathParts := strings.Split(path, "/")

	if len(patternParts) != len(pathParts) {
		return false
	}

	for i := range patternParts {
		if patternParts[i] == "*" {
			continue
		}
		if patternParts[i] != pathParts[i] {
			return false
		}
	}
	return true
}

func removeFilterPrefix(pattern, path string) (string, error) {
	patternParts := strings.Split(pattern, "/")
	pathParts := strings.Split(path, "/")
	var resultParts []string

	for i := range patternParts {
		if patternParts[i] == "*" {
			resultParts = append(resultParts, pathParts[i])
		}
	}

	return url.JoinPath("/", resultParts...)
}
