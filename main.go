package main

import (
	"bufio"
	"flag"
	"fmt"
	"go/build"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
)

type details struct {
	fullPath   string
	importPath string
	valid      bool
}

type index struct {
	index []details
	mu    sync.RWMutex
}

type queryKind uint

const (
	kindImports queryKind = iota + 1
	kindDirs
)

var (
	httpFlag = flag.String("http", "localhost:6118", "HTTP service address, e.g. 'localhost:6118'")
	exclFlag = flag.String("exclude", "", "List of directories to exclude from indexing")
	rootFlag = flag.String("root", "", "List of root directories with go packages")

	dirs index

	exclusionList = `.git .hg`
	exclusions    = loadExclusions(strings.NewReader(exclusionList))
)

// indexPackages walks the directory trees and creates an index dirs
// with path information.
func indexPackages() {
	dirs.mu.Lock()
	defer dirs.mu.Unlock()

	dirs.index = []details{}
	ctx := build.Default

	roots := []string{}
	if *rootFlag == "" {
		roots = ctx.SrcDirs()
	} else {
		roots = strings.Split(*rootFlag, string(os.PathListSeparator))
	}
	for _, root := range roots {
		filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
			if !info.IsDir() {
				return nil
			}

			// Skip directories in the exclusion list.
			dir := filepath.Base(path)
			if _, ok := exclusions[dir]; ok {
				return filepath.SkipDir
			}

			p, err := ctx.ImportDir(path, 0)
			dirs.index = append(dirs.index, details{
				fullPath:   path,
				importPath: p.ImportPath,
				valid:      err == nil,
			})

			return nil
		})
	}
	log.Printf("Indexed %d directories", len(dirs.index))
}

// updatePackages() updates packages' index at regular intervals.
func updatePackages() {
	for {
		select {
		case <-time.Tick(10 * time.Minute):
			indexPackages()
		}
	}
}

// loadExclusions loads a list of directory names to exclude from indexing.
func loadExclusions(r io.Reader) map[string]struct{} {
	exclusions := make(map[string]struct{})
	s := bufio.NewScanner(r)
	s.Split(bufio.ScanWords)

	for s.Scan() {
		exclusions[s.Text()] = struct{}{}
	}
	return exclusions
}

// queryIndex returns a list of absolute directory paths or
// full import paths matching a partial path query.
func queryIndex(query string, kind queryKind) (out []string) {
	dirs.mu.RLock()
	defer dirs.mu.RUnlock()

	sep := string(os.PathSeparator)

	// Match full names. (e.g., if typed "os", match a package or dir
	// named "os", but not "paxos").
	switch kind {
	case kindDirs:
		// Reverse the slashes in Windows.
		if runtime.GOOS == "windows" {
			query = strings.Join(strings.Split(query, "/"), sep)
		}

		query = sep + query
	case kindImports:
		query = "/" + query
	}

	// Valid paths are the paths with packages.
	// Invalid paths are subdirectories leading to valid paths.
	valid, invalid := []string{}, []string{}
	for _, c := range dirs.index {
		var path string

		switch kind {
		case kindImports:
			if strings.HasSuffix("/"+c.importPath, query) {
				path = c.importPath
			}
		case kindDirs:
			if strings.HasSuffix(c.fullPath, query) {
				path = c.fullPath

			}
		}

		if path == "" {
			continue
		}

		if c.valid {
			valid = append(valid, path)
		} else {
			invalid = append(invalid, path)
		}
	}

	// Return invalid paths if there are no valid ones.
	out = valid
	if len(valid) == 0 {
		out = invalid
	}
	return
}

func handler(w http.ResponseWriter, r *http.Request) {
	query, queryKind := r.URL.Path[1:], kindDirs

	switch strings.Split(query, "/")[0] {
	case "imports":
		query = strings.Replace(query, "imports/", "", 1)
		queryKind = kindImports
	case "dirs":
		query = strings.Replace(query, "dirs/", "", 1)
	case "update":
		indexPackages()
		return
	}

	fmt.Fprintln(w, strings.Join(queryIndex(query, queryKind), "\n"))
}

func main() {
	flag.Usage = func() {
		fmt.Println(`gopaths [-http=[HOST]:PORT] [-exclusions FILE] [-root DIRS]`)
		flag.PrintDefaults()
		os.Exit(2)
	}
	flag.Parse()

	if *exclFlag != "" {
		f, err := os.Open(*exclFlag)
		if err != nil {
			log.Fatalf("%v\n", err)
		}

		exclusions = loadExclusions(bufio.NewReader(f))
		f.Close()
	}

	indexPackages()
	go updatePackages()

	http.HandleFunc("/", handler)
	log.Fatal(http.ListenAndServe(*httpFlag, nil))
}
