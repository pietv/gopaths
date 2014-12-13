package main

import (
	"bufio"
	"go/build"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type index struct {
	mu         sync.RWMutex
	index      []details
	rootDirs   []string
	exclusions map[string]struct{}
}

type details struct {
	fullPath   string
	importPath string
	valid      bool
}

type queryKind uint

const (
	kindImports queryKind = iota + 1
	kindDirs
)

// Index walks the directory trees and creates an index with path information.
func (dirs *index) Index() {
	dirs.mu.Lock()
	defer dirs.mu.Unlock()

	dirs.index = []details{}

	for _, root := range dirs.rootDirs {
		filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
			if !info.IsDir() {
				return nil
			}

			// Skip directories in the exclusion list.
			dir := filepath.Base(path)
			if _, ok := dirs.exclusions[dir]; ok {
				return filepath.SkipDir
			}

			p, err := build.Default.ImportDir(path, 0)
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

// UpdateIndex updates packages' index at regular intervals.
func (dirs *index) UpdateIndex() {
	for {
		select {
		case <-time.Tick(45 * time.Minute):
			dirs.Index()
		}
	}
}

// Exclusions loads a list of directory names to exclude from indexing.
func (dirs *index) Exclusions(r io.Reader) {
	dirs.mu.Lock()
	defer dirs.mu.Unlock()

	dirs.exclusions = make(map[string]struct{})
	s := bufio.NewScanner(r)
	s.Split(bufio.ScanWords)

	for s.Scan() {
		dirs.exclusions[s.Text()] = struct{}{}
	}
}

// QueryIndex returns a list of absolute directory paths or
// full import paths matching a partial path query.
func (dirs *index) QueryIndex(query string, kind queryKind) (out []string) {
	dirs.mu.RLock()
	defer dirs.mu.RUnlock()

	sep := string(os.PathSeparator)

	// Match full names. (e.g., if typed "os", match a package or dir
	// named "os", but not "paxos".)
	switch kind {
	case kindDirs:
		// Reverse the slashes in Windows.
		query = strings.Join(strings.Split(query, "/"), sep)

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

// Roots sets a list of directory paths where Go packages are going to be
// searched for in.
func (dirs *index) Roots(roots []string) error {
	dirs.mu.Lock()
	defer dirs.mu.Unlock()

	dirs.rootDirs = []string{}

	// Remove duplicate directories and check for existence.
	seen := map[string]bool{}
	for _, root := range roots {
		absPath, _ := filepath.Abs(root)

		fi, err := os.Stat(root)
		if err != nil {
			return err
		}
		if fi.IsDir() == false {
			return os.ErrInvalid
		}

		if _, ok := seen[absPath]; ok {
			continue
		}
		seen[absPath] = true
		dirs.rootDirs = append(dirs.rootDirs, absPath)
	}

	return nil
}
