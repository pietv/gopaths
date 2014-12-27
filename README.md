gopaths [![GoDoc](https://godoc.org/github.com/pietv/gopaths?status.png)](https://godoc.org/github.com/pietv/gopaths) [![Build Status](https://drone.io/github.com/pietv/gopaths/status.png)](https://drone.io/github.com/pietv/gopaths/latest) [![Build status](https://ci.appveyor.com/api/projects/status/u2xfqdwb6t6c8b35/branch/master?svg=true)](https://ci.appveyor.com/project/pietv/gopaths/branch/master)
=======
Go packages indexer. It searches for Go packages in GOROOT and GOPATH
directories and then responds to shorter package paths with full paths.
Useful together with 'cd' and 'godoc' commands.

Install
=======
```shell
$ go get github.com/pietv/gopaths
$ go install github.com/pietv/gopaths
```

Usage
=====
Service start:
```shell
$ gopaths -http=:6118 &
```

Search for a package:
```shell
$ curl :6118/imports/log
log
google.golang.org/appengine/internal/log
google.golang.org/appengine/log
```

Search for a directory containing the package:
```shell
$ curl :6118/dirs/rand
/Users/peter/go/src/crypto/rand
/Users/peter/go/src/math/rand
```

Update the index:
```shell
curl :6118/update
```
