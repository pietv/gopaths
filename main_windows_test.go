package main

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

var winBasicTestDetails = []details{
	{"C:\\a\\a", "a/a", true},
	{"C:\\b\\a", "b/a", true},
	{"C:\\a", "a", true},
	{"C:\\ab", "ab", true},
	{"C:\\a\\b", "a/b", false},
	{"C:\\Documents and Settings\\ab\\ab", "ab/ab", true},
	{"C:\\Ducuments and Settings\\Go\\ab\\a.b", "ab/a.b", true},
	{"C:\\c-c\\c.c", "c-c/c.c", false},
	{"\\c-c\\c.c\\c.c", "c-c/c.c/c.c", true},
	{"\\a\\b\\c", "a/b/c", true},
	{"\\.\\d\\d", "d/d", false},
}

var winBasicTests = []struct {
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

	{"dirs/a", []string{"C:\\a\\a", "C:\\b\\a", "C:\\a"}},
	{"dirs/a/a", []string{"C:\\a\\a"}},
	{"dirs/b", []string{"C:\\a\\b"}},
	{"dirs/ab", []string{"C:\\ab", "C:\\long path\\ab\\ab"}},
	{"dirs/a.b", []string{"C:\\Ducuments and Settings\\Go\\ab\\a.b"}},
	{"dirs/c", []string{"\\a\\b\\c"}},
	{"dirs/c-c/c.c", []string{"\\c-c\\c.c"}},
	{"dirs/c.c", []string{"\\c-c\\c.c\\c.c"}},

	{"/update", []string{""}},
	{"/update/a", []string{""}},

	{"dir/a", []string{""}},
	{"import/a", []string{""}},
	{"//dirs/a", []string{""}},
	{"//import/a", []string{""}},
}

func TestWindowsBasic(t *testing.T) {
	dirs := index{
		index: winBasicTestDetails,
	}

	for _, test := range winBasicTests {
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
