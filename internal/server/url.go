package server

import (
	"fmt"
	"net/url"
	"strings"
)

func matchRoute(pattern, path string) bool {
	if pattern == "*" {
		return true
	}

	// Parse the URL to separate path from query
	parsedURL, err := url.Parse(path)
	if err != nil {
		return false
	}
	path = parsedURL.Path

	patternParts := strings.Split(strings.Trim(pattern, "/"), "/")
	pathParts := strings.Split(strings.Trim(path, "/"), "/")

	for i := range patternParts {
		if patternParts[i] == "*" {
			continue
		}

		if i >= len(pathParts) {
			return false
		}

		if patternParts[i] != pathParts[i] {
			return false
		}
	}
	return true
}

func removeFilterPrefix(pattern, path string) (string, error) {
	u, err := url.Parse(path)
	if err != nil {
		return "", fmt.Errorf("invalid URL: %v", err)
	}

	if strings.Count(path, "/") == 1 {
		path += "/"
	}

	patternParts := strings.Split(pattern, "/")
	pathParts := strings.Split(path, "/")
	var resultParts []string

	for i := range patternParts {
		if patternParts[i] == "*" {
			resultParts = append(resultParts, pathParts[i:]...)
		}
	}

	resultPath, err := url.JoinPath("/", resultParts...)
	if err != nil {
		return "", err
	}

	if u.RawQuery != "" {
		resultPath = strings.TrimSuffix(resultPath, "/")
		resultPath += "?" + u.RawQuery
	}
	return resultPath, nil
}
