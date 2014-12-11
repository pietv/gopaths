package main

import (
	"fmt"
	"net/http"
	"strings"
)

func (dirs *index) ServeMux() *http.ServeMux {
	mux := http.NewServeMux()

	mux.Handle("/imports/", http.StripPrefix("/imports/", dirs.ImportsHandler()))
	mux.Handle("/dirs/", http.StripPrefix("/dirs/", dirs.DirsHandler()))
	mux.Handle("/update", dirs.UpdateHandler())
	mux.Handle("/", http.StripPrefix("/", dirs.DirsHandler()))

	return mux
}

func (dirs *index) DirsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, strings.Join(dirs.QueryIndex(r.URL.Path, kindDirs), "\n"))
	}
}

func (dirs *index) ImportsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, strings.Join(dirs.QueryIndex(r.URL.Path, kindImports), "\n"))
	}
}

func (dirs *index) UpdateHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		dirs.Index()
	}
}
