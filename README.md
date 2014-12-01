gopaths
=======
Go packages indexer. It searches for packages in GOROOT and GOPATH and
then services HTTP requests with full package names in response to
shorter names. Useful together with 'cd' and 'godoc' commands.

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
$ curl :6118/dirs/present
/Users/peter/src/golang.org/x/tools/cmd/present
/Users/peter/src/golang.org/x/tools/present
```

Update the index:
```shell
curl :6118/update
```
