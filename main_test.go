package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"runtime"
	"strings"
	"testing"
)

const (
	hostPrefix    = "http://localhost/"
	packagePrefix = "github.com/pietv/gopaths/"

	sep = string(os.PathSeparator)
)

var basicTestDetails = convertDetails([]details{
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
})

func init() {
	// Turn off the logger.
	log.SetOutput(ioutil.Discard)
}

// convertDetails converts directory paths on Windows.
func convertDetails(d []details) []details {
	if runtime.GOOS != "windows" {
		return d
	}

	converted := []details{}
	for _, elem := range d {
		fullPath := strings.Replace(elem.fullPath, "/", sep, -1)

		converted = append(converted, details{
			fullPath:   fullPath,
			importPath: elem.importPath,
			valid:      elem.valid,
		})
	}

	return converted
}

// Trim extra '\n' from the response and then split.
func slice(resp string) []string {
	return strings.Split(strings.Trim(resp, "\n"), "\n")
}

var BasicTests = []struct {
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

	// Without prefixes it behaves like /dirs/...
	{"a", []string{"/root/a/a", "/root/b/a", "/root/a"}},
	{"a/a", []string{"/root/a/a"}},
	{"b", []string{"/root/a/b"}},
	{"ab", []string{"/root/ab", "/long path/ab/ab"}},

	{"dir/a", []string{""}},
	{"import/a", []string{""}},
	{"//dirs/a", []string{""}},
	{"//import/a", []string{""}},
}

func TestBasic(t *testing.T) {
	dirs := index{
		index: basicTestDetails,
	}

	fmt.Println(basicTestDetails)

	for _, test := range BasicTests {
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

var IndexerImportsTests = []struct {
	query string
	out   []string
}{
	{"imports/fmt", []string{"fmt"}},
	{"imports/math/rand", []string{"math/rand"}},
	{"imports/rand", []string{"crypto/rand", "math/rand"}},
	{"imports/template", []string{"html/template", "text/template"}},
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

		actual := slice(rec.Body.String())
		if reflect.DeepEqual(actual, test.out) != true {
			t.Errorf("%q: got %q, want %q", test.query, actual, test.out)
		}
	}
}

func convertSlashes(path string) string {
	return strings.Join(strings.Split(path, "/"), sep)
}

func compareDirs(actual, out []string) bool {
	if len(actual) != len(out) {
		return false
	}

	for i, _ := range actual {
		if !strings.HasSuffix(
			actual[i],
			convertSlashes(out[i])) {
			return false
		}
	}
	return true
}

var IndexerDirTests = []struct {
	query string
	out   []string
}{
	{"dirs/fmt", []string{"fmt"}},
	{"dirs/math/rand", []string{"math/rand"}},
	{"dirs/cmd/gofmt", []string{"cmd/gofmt"}},
	{"dirs/rand", []string{"crypto/rand", "math/rand"}},
	{"dirs/template", []string{"html/template", "text/template"}},
}

func TestIndexerDirs(t *testing.T) {
	dirs := index{}
	dirs.Roots([]string{runtime.GOROOT()})
	dirs.Index()

	for _, test := range IndexerDirTests {
		req, err := http.NewRequest("GET", hostPrefix+test.query, nil)
		if err != nil {
			t.Errorf("GET %q failed", test.query)
		}

		rec := httptest.NewRecorder()
		dirs.ServeHTTP(rec, req)

		if actual := slice(rec.Body.String()); compareDirs(actual, test.out) != true {
			t.Errorf("%q: got %q, want %q", test.query, actual, test.out)
		}
	}
}

var ExclusionsTests = []struct {
	query string
	out   []string
}{
	{"imports/math/rand", []string{""}},
	{"imports/rand", []string{"crypto/rand"}},
	{"imports/io", []string{""}},
	{"imports/os", []string{"os"}},
	{"imports/os/exec", []string{""}},
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

		actual := slice(rec.Body.String())
		if reflect.DeepEqual(actual, test.out) != true {
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

		actual := slice(rec.Body.String())
		if reflect.DeepEqual(actual, test.out) != true {
			t.Errorf("%q: got %q, want %q", test.query, actual, test.out)
		}
	}
}
