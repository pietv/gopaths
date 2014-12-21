// The gopaths server responds to partial path requests with full package
// import or directory paths, depending on the request type.
//
// Usage: gopaths [-http [HOST]:PORT] [-root DIRS] [-exclude FILE]
//
//   -http=":6118"
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
// line. If there are no package matches, paths leading to the base path
// are returned; again, if there are any.
//
// For example, if the requested path is “io”, this path will be matched:
//
//   io
//
// but these will be not:
//
//   bufio
//   testing/iotest
//   cmd/internal/rsc.io/x86/x86asm
//
// On the other hand, if the requested path is “go.net”, and there are no
// indexed packages with “go.net” at the end, this path will be returned:
//
//   code.google.com/p/go.net
//
// It's a parent path to many other packages.
//
//
// The paths are queried using a Web browser, preferably a console one like
// curl(1) or wget(1), because gopaths is intended to be a CLI server.
//
// There are three request types, specified by path prefixes:
//
//   GET /dirs/{PATH}
//     Return directory paths matching PATH.
//
//   GET /imports/{PATH}
//     Return import paths matching PATH.
//
//   GET /update
//     Update the directory index. The directory index updates itself
//     every 45 minutes. Occasionally, a faster update might be needed.
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
