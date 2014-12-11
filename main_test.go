package main

import (
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
)

func init() {
	// Turn off the logger.
	log.SetOutput(ioutil.Discard)
}

// slice trims extra '\n' from the HTTP response and then splits.
func slice(resp string) []string {
	return strings.Split(strings.Trim(resp, "\n"), "\n")
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

		if actual := slice(rec.Body.String()); reflect.DeepEqual(actual, test.out) != true {
			t.Errorf("%q: got %q, want %q", test.query, actual, test.out)
		}
	}
}

func compareDirs(actual, out []string) bool {
	if len(actual) != len(out) {
		return false
	}

	sep := string(os.PathSeparator)

	for i, _ := range actual {
		if !strings.HasSuffix(actual[i],
			strings.Join(strings.Split(out[i], "/"), sep)) {

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

		if actual := slice(rec.Body.String()); reflect.DeepEqual(actual, test.out) != true {
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

		if actual := slice(rec.Body.String()); reflect.DeepEqual(actual, test.out) != true {
			t.Errorf("%q: got %q, want %q", test.query, actual, test.out)
		}
	}
}
