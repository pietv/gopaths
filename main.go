package main

import (
	"bufio"
	"flag"
	"fmt"
	"go/build"
	"log"
	"net/http"
	"os"
	"strings"
)

var (
	httpFlag = flag.String("http", ":6118", "HTTP service address, e.g. 'localhost:6118'")
	exclFlag = flag.String("exclude", "", "List of directories to exclude from indexing")
	rootFlag = flag.String("root", "", "List of root directories containing go packages")

	defaultExclusions = `.git .hg`
)

func main() {
	flag.Usage = func() {
		fmt.Println(`gopaths [-http=[HOST]:PORT] [-exclusions FILE] [-root DIRS]`)
		flag.PrintDefaults()
		os.Exit(2)
	}
	flag.Parse()

	dirs := index{}

	if *exclFlag != "" {
		f, err := os.Open(*exclFlag)
		if err != nil {
			log.Fatalf("%v\n", err)
		}

		dirs.Exclusions(bufio.NewReader(f))
		f.Close()
	} else {
		dirs.Exclusions(strings.NewReader(defaultExclusions))
	}

	if *rootFlag != "" {
		dirs.Roots(strings.Split(*rootFlag, string(os.PathListSeparator)))
	} else {
		dirs.Roots(build.Default.SrcDirs())
	}

	dirs.Index()
	go dirs.UpdateIndex()

	log.Fatal(http.ListenAndServe(*httpFlag, dirs.ServeMux()))
}
