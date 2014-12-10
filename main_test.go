package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"runtime"
	"strings"
	"testing"
)

const (
	hostPrefix    = "http://localhost/"
	packagePrefix = "github.com/pietv/gopaths/"
)

var basicTestDetails = []details{
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

var BasicTests = []struct {
	query string
	out   string
}{
	{"imports/a", "a/a b/a a"},
	{"imports/a/a", "a/a"},
	{"imports/b", "a/b"},
	{"imports/ab", "ab ab/ab"},
	{"imports/a.b", "ab/a.b"},
	{"imports/c", "a/b/c"},
	{"imports/c-c/c.c", "c-c/c.c"},
	{"imports/c.c", "c-c/c.c/c.c"},

	{"dirs/a", "/root/a/a /root/b/a /root/a"},
	{"dirs/a/a", "/root/a/a"},
	{"dirs/b", "/root/a/b"},
	{"dirs/ab", "/root/ab /long path/ab/ab"},
	{"dirs/a.b", "/long/path/ab/a.b"},
	{"dirs/c", "/a/b/c"},
	{"dirs/c-c/c.c", "/c-c/c.c"},
	{"dirs/c.c", "/c-c/c.c/c.c"},

	// Without a prefix it behaves like /dirs/...
	{"a", "/root/a/a /root/b/a /root/a"},
	{"a/a", "/root/a/a"},
	{"b", "/root/a/b"},
	{"ab", "/root/ab /long path/ab/ab"},

	{"dir/a", ""},
	{"import/a", ""},
	{"//dirs/a", ""},
	{"//import/a", ""},
}

var IndexerImportsTests = []struct {
	query string
	out   string
}{
	{"imports/fmt", "fmt"},
	{"imports/math/rand", "math/rand"},
	{"imports/rand", "crypto/rand math/rand"},
	{"imports/template", "html/template text/template"},
}

var IndexerDirsTests = []struct {
	query string
	out   []string
}{
	{"dirs/fmt", []string{"fmt"}},
	{"dirs/math/rand", []string{"math/rand"}},
	{"dirs/cmd/gofmt", []string{"cmd/gofmt"}},
	{"dirs/rand", []string{"crypto/rand", "math/rand"}},
	{"dirs/template", []string{"html/template", "text/template"}},
}

var ExclusionsTests = []struct {
	query string
	out   string
}{
	{"imports/math/rand", ""},
	{"imports/rand", "crypto/rand"},
	{"imports/io", ""},
	{"imports/os", "os"},
	{"imports/os/exec", ""},
}

func init() {
	// Turn off the logger.
	log.SetOutput(ioutil.Discard)
}

// Trim extra '\n' from responce, convert '\n' to ' '.
func trimResponse(resp string) string {
	return strings.Join(strings.Split(strings.Trim(resp, "\n"), "\n"), " ")
}

func TestBasic(t *testing.T) {
	dirs := index{
		index: basicTestDetails,
	}

	for _, test := range BasicTests {
		req, err := http.NewRequest("GET", hostPrefix+test.query, nil)
		if err != nil {
			t.Errorf("GET %q failed", test.query)
		}

		rec := httptest.NewRecorder()
		dirs.ServeHTTP(rec, req)

		if actual := trimResponse(rec.Body.String()); actual != test.out {
			t.Errorf("%q: got %q, want %q", test.query, actual, test.out)
		}
	}
}

func TestIndexerImports(t *testing.T) {
	dirs := index{}
	dirs.Roots([]string{runtime.GOROOT()})
	dirs.Index()

	for _, test := range IndexerImportsTests {
		req, err := http.NewRequest("GET", hostPrefix+test.query, nil)
		if err != nil {
			t.Errorf("GET %q failed", test.query)
		}

		rec := httptest.NewRecorder()
		dirs.ServeHTTP(rec, req)

		if actual := trimResponse(rec.Body.String()); actual != test.out {
			t.Errorf("%q: got %q, want %q", test.query, actual, test.out)
		}
	}
}

// Trim extra '\n' from the responce.
func trimResponseSlice(resp string) []string {
	return strings.Split(strings.Trim(resp, "\n"), "\n")
}

func compareSuffixes(actual, out []string) bool {
	if len(actual) != len(out) {
		fmt.Println("YYY")
		return false
	}
	for i, _ := range actual {
		if !strings.HasSuffix(actual[i], out[i]) {
			fmt.Println("XXX", actual[i], out[i])
			return false
		}
	}
	return true
}

func TestIndexerDirs(t *testing.T) {
	dirs := index{}
	dirs.Roots([]string{runtime.GOROOT()})
	dirs.Index()

	for _, test := range IndexerDirsTests {
		req, err := http.NewRequest("GET", hostPrefix+test.query, nil)
		if err != nil {
			t.Errorf("GET %q failed", test.query)
		}

		rec := httptest.NewRecorder()
		dirs.ServeHTTP(rec, req)

		if actual := trimResponseSlice(rec.Body.String()); compareSuffixes(actual, test.out) != true {
			t.Errorf("%q: got %q, want %q", test.query, actual, test.out)
		}
	}
}

func TestExclusions(t *testing.T) {
	dirs := index{}
	dirs.Roots([]string{runtime.GOROOT()})
	dirs.Exclusions(strings.NewReader(`io math exec`))
	dirs.Index()

	for _, test := range ExclusionsTests {
		req, err := http.NewRequest("GET", hostPrefix+test.query, nil)
		if err != nil {
			t.Errorf("GET %q failed", test.query)
		}

		rec := httptest.NewRecorder()
		dirs.ServeHTTP(rec, req)

		if actual := trimResponse(rec.Body.String()); actual != test.out {
			t.Errorf("%q: got %q, want %q", test.query, actual, test.out)
		}
	}
}

func TestDuplicateDirs(t *testing.T) {
	dirs := index{}
	dirs.Roots([]string{
		runtime.GOROOT(),
		runtime.GOROOT(),
	})
	dirs.Index()

	for _, test := range IndexerImportsTests {
		req, err := http.NewRequest("GET", hostPrefix+test.query, nil)
		if err != nil {
			t.Errorf("GET %q failed", test.query)
		}

		rec := httptest.NewRecorder()
		dirs.ServeHTTP(rec, req)

		if actual := trimResponse(rec.Body.String()); actual != test.out {
			t.Errorf("%q: got %q, want %q", test.query, actual, test.out)
		}
	}
}
