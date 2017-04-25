all: fmt build

build: werifyd werifyctl

werifyd:
	go build ./cmd/werifyd

werifyctl:
	go build ./cmd/werifyctl

fmt:
	find . ! -path "*/vendor/*" -type f -name '*.go' -exec gofmt -l -s -w {} \;

clean:
	rm -vf werifyd werifyctl

.PHONY: werifyd werifyctl fmt clean build
