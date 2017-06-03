CWD=$(shell pwd)
GOPATH := $(CWD)

prep:
	if test -d pkg; then rm -rf pkg; fi

rmdeps:
	if test -d src; then rm -rf src; fi 

self:   prep
	if test -d src; then rm -rf src; fi
	mkdir src
	mkdir -p src/github.com/whosonfirst/go-whosonfirst-pinboard
	cp -r internetarchive src/github.com/whosonfirst/go-whosonfirst-pinboard/
	cp -r pinboard src/github.com/whosonfirst/go-whosonfirst-pinboard/
	cp -r webpage src/github.com/whosonfirst/go-whosonfirst-pinboard/
	cp -r whosonfirst src/github.com/whosonfirst/go-whosonfirst-pinboard/
	cp -r vendor/src/* src/

deps:
	@GOPATH=$(GOPATH) go get -u "github.com/facebookgo/grace/gracehttp"
	@GOPATH=$(GOPATH) go get -u "github.com/tidwall/gjson"
	@GOPATH=$(GOPATH) go get -u "github.com/tidwall/pretty"
	@GOPATH=$(GOPATH) go get -u "github.com/whosonfirst/go-sanitize"
	@GOPATH=$(GOPATH) go get -u "github.com/whosonfirst/go-whosonfirst-uri"
	@GOPATH=$(GOPATH) go get -u "golang.org/x/net/html"

vendor-deps: rmdeps deps
	if test ! -d vendor; then mkdir vendor; fi
	if test -d vendor/src; then rm -rf vendor/src; fi
	cp -r src vendor/src
	find vendor -name '.git' -print -type d -exec rm -rf {} +
	rm -rf src

fmt:
	go fmt cmd/*.go
	go fmt internetarchive/*.go
	go fmt pinboard/*.go
	go fmt webpage/*.go
	go fmt whosonfirst/*.go

bin:	self
	@GOPATH=$(GOPATH) go build -o bin/wof-archive-daemon cmd/wof-archive-daemon.go
	@GOPATH=$(GOPATH) go build -o bin/wof-archive-url cmd/wof-archive-url.go
