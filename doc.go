// The gopaths server responds to partial path requests with full package
// import or directory paths, depending on the request type.
//
// Usage: gopaths [-http=[HOST]:PORT] [-root DIRS] [exclusions FILE]
//
//   -http="localhost:6118"
// 	Listen on HOST on PORT.
//
//   -root=""
//      Directories to look for Go packages in, separated by ‘:’ in Unix
//      and ‘;’ in Windows. By default, the packages are looked for
//      in GOROOT and GOPATH.
//
//   -exclude=""
//      FILE containing a list of whitespace separated directory names
//      in which gopaths won't be looking into when searching for packages.
//
//
// Paths are matched against the base path (deepest sitting directory):
//
// If there are many matches, all matches are returned; each on a separate
// line. If there are no matches, paths leading to the base are returned.
//
// For example, if the requested path is “io”, these paths are matched:
//
//   io
//
// but these are not:
//
//   bufio
//   testing/iotest
//   cmd/internal/rsc.io/x86/x86asm
//
// On the other hand, if the requested path is “go.net”, and there are no
// indexed packages with “go.net” at the end, this path is returned:
//
//   code.google.com/p/go.net
//
// It's a parent path to many other packages.
//
//
// The paths are queried using a Web browser, most usefully: curl(1) and
// wget(1), because gopaths is a CLI server.
//
// There are three request types, specified by prefixes:
//
//   curl :6118/dirs/PATH
//     Return directory paths matching PATH.
//
//   curl :6118/imports/PATH
//     Return import paths matching PATH.
//
//   curl :6118/update
//     Update the directory index. The directory index updates itself
//     every 45 minutes, but occasionally faster update might be needed.
//
// Examples:
//
//   $ curl :6118/imports/log
//   log
//   google.golang.org/appengine/internal/log
//   google.golang.org/appengine/log
//
//   $ curl :6118/dirs/rand
//   /Users/peter/go/src/crypto/rand
//   /Users/peter/go/src/math/rand
//
package main
