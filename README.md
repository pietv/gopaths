gopaths [![Build Status](https://drone.io/github.com/pietv/gopaths/status.png)](https://drone.io/github.com/pietv/gopaths/latest)
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
$ nohup gopaths -http=:6118 &
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
