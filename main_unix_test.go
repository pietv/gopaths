// +build !windows

package main

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

var unixBasicTestDetails = []details{
	{"/root/a/a", "a/a", true},
	{"/root/b/a", "b/a", true},
	{"/root/a", "a", true},
	{"/root/ab", "ab", true},
	{"/root/a/b", "a/b", false},
	{"/long path/ab/ab", "ab/ab", true},
	{"/long/path/ab/a.b", "ab/a.b", true},
	{"/c-c/c.c", "c-c/c.c", false},
	{"/c-c/c.c/c.c", "c-c/c.c/c.c", true},
	{"/a/b/c", "a/b/c", true},
	{"./d/d", "d/d", false},
}

var unixBasicTests = []struct {
	query string
	out   []string
}{
	{"imports/a", []string{"a/a", "b/a", "a"}},
	{"imports/a/a", []string{"a/a"}},
	{"imports/b", []string{"a/b"}},
	{"imports/ab", []string{"ab", "ab/ab"}},
	{"imports/a.b", []string{"ab/a.b"}},
	{"imports/c", []string{"a/b/c"}},
	{"imports/c-c/c.c", []string{"c-c/c.c"}},
	{"imports/c.c", []string{"c-c/c.c/c.c"}},

	{"dirs/a", []string{"/root/a/a", "/root/b/a", "/root/a"}},
	{"dirs/a/a", []string{"/root/a/a"}},
	{"dirs/b", []string{"/root/a/b"}},
	{"dirs/ab", []string{"/root/ab", "/long path/ab/ab"}},
	{"dirs/a.b", []string{"/long/path/ab/a.b"}},
	{"dirs/c", []string{"/a/b/c"}},
	{"dirs/c-c/c.c", []string{"/c-c/c.c"}},
	{"dirs/c.c", []string{"/c-c/c.c/c.c"}},

	{"/update", []string{""}},
	{"/update/a", []string{""}},

	// Without prefixes, behave like /dirs/...
	{"a", []string{"/root/a/a", "/root/b/a", "/root/a"}},
	{"a/a", []string{"/root/a/a"}},
	{"b", []string{"/root/a/b"}},
	{"ab", []string{"/root/ab", "/long path/ab/ab"}},

	{"dir/a", []string{""}},
	{"import/a", []string{""}},
	{"//dirs/a", []string{""}},
	{"//import/a", []string{""}},
}

func TestUnixBasic(t *testing.T) {
	dirs := index{
		index: unixBasicTestDetails,
	}

	for _, test := range unixBasicTests {
		req, err := http.NewRequest("GET", hostPrefix+test.query, nil)
		if err != nil {
			t.Errorf("GET %q failed", test.query)
		}

		rec := httptest.NewRecorder()
		dirs.ServeHTTP(rec, req)

		actual := slice(rec.Body.String())
		if reflect.DeepEqual(actual, test.out) != true {
			t.Errorf("%q: got %q, want %q", test.query, actual, test.out)
		}
	}
}
