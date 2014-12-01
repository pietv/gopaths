gopaths
=======
Go packages indexer. It searches for Go packages in GOROOT and GOPATH 
directories and then responds to shorter package paths with full paths.
Useful together with 'cd' and 'godoc' commands.

Install
=======
```shell
$ go install github.com/pietv/gopaths/cmd/gopaths
```

Usage
=====
Service start:
```shell
$ gopaths -http=:6118 &
```

Search for a package:
```shell
$ curl :6118/imports/present
golang.org/x/tools/cmd/present
golang.org/x/tools/present
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
