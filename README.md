informer
====

## Description
Informer which you can watch other users action on your Linux/Unix servers.

## Usage
Check running tty.

```bash
$ informer list
```

Watching other users action.

```bash
# root
$ informer watch pts/0
```

If you want review for later.

```bash
# root
$ informer watch -o /tmp/hoge.trace pts/0
$ informer review /tmp/hoge.trace
```

## Install

To install, use `go get`:

```bash
$ go get -d github.com/nashiox/informer
```

or

```bash
$ git clone https://github.com/nashoix/informer.git
$ cd informer
$ go get ./...
$ GOOS=linux GOARCH=amd64 go build
$ sudo mv informer /usr/local/bin/
```

## Author

[nashiox](https://github.com/nashiox)
