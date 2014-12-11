package main

import (
	"go/build"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

var (
	hostPrefix    = "http://localhost:6118/"
	packagePrefix = "github.com/pietv/gopaths/testdata"

	// Package location.
	dirPrefix = func() string {
		p, err := build.Default.Import(packagePrefix, "", build.FindOnly)
		if err != nil {
			panic(err)
		}

		return p.Dir
	}()

	sep = string(os.PathSeparator)
)

func init() {
	// Turn off the logger.
	log.SetOutput(ioutil.Discard)
}

// slice trims extra '\n' from the HTTP response and then splits it.
func slice(resp string) []string {
	return strings.Split(strings.Trim(resp, "\n"), "\n")
}

// prefixDir appends the prefix to the paths,
// and also converts the path separators on Windows.
func prefixDir(paths []string, prefix string) []string {
	prefixed := []string{}
	for _, path := range paths {
		osPath := strings.Join(strings.Split(filepath.Join(prefix, path), "/"), sep)

		prefixed = append(prefixed, osPath)
	}
	return prefixed
}

// prefixImp appends the prefix to the import paths.
func prefixImp(paths []string, prefix string) []string {
	prefixed := []string{}
	for _, path := range paths {
		if path != "" {
			prefixed = append(prefixed, prefix+"/"+path)
		} else {
			prefixed = append(prefixed, path)
		}
	}
	return prefixed
}

var QueryTestDetails = []details{
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

var QueryImportsTests = []struct {
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
}

var QueryDirTests = []struct {
	query string
	out   []string
}{
	{"dirs/a", []string{"/root/a/a", "/root/b/a", "/root/a"}},
	{"dirs/a/a", []string{"/root/a/a"}},
	{"dirs/b", []string{"/root/a/b"}},
	{"dirs/ab", []string{"/root/ab", "/long path/ab/ab"}},
	{"dirs/a.b", []string{"/long/path/ab/a.b"}},
	{"dirs/c", []string{"/a/b/c"}},
	{"dirs/c-c/c.c", []string{"/c-c/c.c"}},
	{"dirs/c.c", []string{"/c-c/c.c/c.c"}},
}

func TestQueryImports(t *testing.T) {
	dirs := index{
		index: QueryTestDetails,
	}

	for _, test := range QueryImportsTests {
		req, err := http.NewRequest("GET", hostPrefix+test.query, nil)
		if err != nil {
			t.Errorf("GET %q failed", test.query)
		}

		rec := httptest.NewRecorder()
		dirs.ServeMux().ServeHTTP(rec, req)

		if actual := slice(rec.Body.String()); reflect.DeepEqual(actual, test.out) != true {
			t.Errorf("%q: got %q, want %q", test.query, actual, test.out)
		}
	}
}

func TestQueryDirs(t *testing.T) {
	// In Windows, convert indexed directory paths.
	queryDetails := []details{}
	for _, data := range QueryTestDetails {
		fullPath := strings.Join(strings.Split(data.fullPath, "/"), sep)
		queryDetails = append(queryDetails, details{
			fullPath:   fullPath,
			importPath: data.importPath,
			valid:      data.valid,
		})
	}

	dirs := index{index: queryDetails}

	for _, test := range QueryDirTests {
		req, err := http.NewRequest("GET", hostPrefix+test.query, nil)
		if err != nil {
			t.Errorf("GET %q failed", test.query)
		}

		rec := httptest.NewRecorder()
		dirs.ServeMux().ServeHTTP(rec, req)

		out := prefixDir(test.out, "")

		if actual := slice(rec.Body.String()); reflect.DeepEqual(actual, out) != true {
			t.Errorf("%q: got %q, want %q", test.query, actual, out)
		}
	}
}

var IndexerImportsTests = []struct {
	query string
	out   []string
}{
	{"imports/a", []string{"a"}},
	{"imports/c", []string{"a/b/c"}},
	{"imports/d", []string{"a/b/c/d", "d", "d/d"}},
	{"imports/bb", []string{"aa/bb"}},
	{"imports/c-c", []string{"c-c"}},
	{"imports/c.c", []string{"c-c/c.c"}},
}

func TestIndexerImports(t *testing.T) {
	imps := index{}
	imps.Roots([]string{"testdata"})
	imps.Index()

	for _, test := range IndexerImportsTests {
		req, err := http.NewRequest("GET", hostPrefix+test.query, nil)
		if err != nil {
			t.Errorf("GET %q failed", test.query)
		}

		rec := httptest.NewRecorder()
		imps.ServeMux().ServeHTTP(rec, req)

		out := prefixImp(test.out, packagePrefix)

		if actual := slice(rec.Body.String()); reflect.DeepEqual(actual, out) != true {
			t.Errorf("%q: got %q, want %q", test.query, actual, out)
		}
	}
}

var IndexerDirTests = []struct {
	query string
	out   []string
}{
	{"a", []string{"/a"}},
	{"c", []string{"/a/b/c"}},
	{"d", []string{"/a/b/c/d", "/d", "/d/d"}},
	{"bb", []string{"/aa/bb"}},
	{"c-c", []string{"/c-c"}},
	{"c.c", []string{"/c-c/c.c"}},
}

func TestIndexerDirs(t *testing.T) {
	imps := index{}
	imps.Roots([]string{"testdata"})
	imps.Index()

	for _, test := range IndexerDirTests {
		req, err := http.NewRequest("GET", hostPrefix+test.query, nil)
		if err != nil {
			t.Errorf("GET %q failed", test.query)
		}

		rec := httptest.NewRecorder()
		imps.ServeMux().ServeHTTP(rec, req)

		out := prefixDir(test.out, dirPrefix)

		if actual := slice(rec.Body.String()); reflect.DeepEqual(actual, out) != true {
			t.Errorf("%q: got %q, want %q", test.query, actual, out)
		}
	}
}

var ExclusionsTests = []struct {
	query string
	out   []string
}{
	{"imports/a", []string{"a"}},
	{"imports/c", []string{""}},
	{"imports/d", []string{"d", "d/d"}},
	{"imports/bb", []string{"aa/bb"}},
	{"imports/c-c", []string{"c-c"}},
	{"imports/c.c", []string{"c-c/c.c"}},
}

func TestExclusions(t *testing.T) {
	imps := index{}
	imps.Roots([]string{"testdata"})
	imps.Exclusions(strings.NewReader(`b c`))
	imps.Index()

	for _, test := range ExclusionsTests {
		req, err := http.NewRequest("GET", hostPrefix+test.query, nil)
		if err != nil {
			t.Errorf("GET %q failed", test.query)
		}

		rec := httptest.NewRecorder()
		imps.ServeMux().ServeHTTP(rec, req)

		out := prefixImp(test.out, packagePrefix)

		if actual := slice(rec.Body.String()); reflect.DeepEqual(actual, out) != true {
			t.Errorf("%q: got %q, want %q", test.query, actual, out)
		}
	}
}

func TestDuplicateRoots(t *testing.T) {
	imps := index{}
	imps.Roots([]string{
		"testdata",
		"testdata",
	})
	imps.Index()

	for _, test := range IndexerImportsTests {
		req, err := http.NewRequest("GET", hostPrefix+test.query, nil)
		if err != nil {
			t.Errorf("GET %q failed", test.query)
		}

		rec := httptest.NewRecorder()
		imps.ServeMux().ServeHTTP(rec, req)

		out := prefixImp(test.out, packagePrefix)

		if actual := slice(rec.Body.String()); reflect.DeepEqual(actual, out) != true {
			t.Errorf("%q: got %q, want %q", test.query, actual, out)
		}
	}
}
