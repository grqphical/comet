package proxy

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMatchRoute(t *testing.T) {
	tests := []struct {
		pattern string
		path    string
		matches bool
	}{
		{"/foo/*", "/foo/bar", true},
		{"/foo/*", "/foo/baz", true},
		{"/foo/bar", "/foo/bar", true},
		{"/foo/bar", "/foo/baz", false},
		{"/foo/*", "/bar/foo", false},
	}

	for _, test := range tests {
		result := matchRoute(test.pattern, test.path)
		assert.Equal(t, test.matches, result)
	}
}

func TestRemoveFilterPrefix(t *testing.T) {
	tests := []struct {
		pattern string
		path    string
		result  string
	}{
		{"/foo/*", "/foo/bar", "/bar"},
		{"/foo", "/foo", "/"},
		{"/foo/*", "/foo/baz", "/baz"},
		{"/foo/bar", "/foo/bar", "/"},
	}

	for _, test := range tests {
		result, err := removeFilterPrefix(test.pattern, test.path)
		assert.NoError(t, err)
		assert.Equal(t, test.result, result)
	}
}
