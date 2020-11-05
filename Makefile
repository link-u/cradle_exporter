.PHONY: all
all: cradle_exporter test/servers/dying-linux-x86_64

cradle_exporter: $(shell find . -type f -name *.go)
	CGO_ENABLED=0 go build \
	  -o "$@" ./cmd/cradle_exporter
	@if ! ldd "$@" 2> /dev/null; then echo "OK: not a dynamic executable!"; fi

test/servers/dying-linux-x86_64: $(shell find test/servers/dying -type f -name *.go)
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
	  -o "$@" ./test/servers/dying
	@if ! ldd "$@" 2> /dev/null; then echo "OK: not a dynamic executable!"; fi

.PHONY: clean
clean:
	rm -Rfv cradle_exporter

.PHONY: cl
cl:
	find . -type f -name *.go | xargs wc -l

.PHONY: test
test:
	go test ./...

.PHONY: test-run
test-run: cradle_exporter
	./cradle_exporter --config=example/config/config.yml

.PHONY: test-call
test-call:
	curl --key example/config/ca/test.key --cert example/config/ca/test.crt -k https://localhost:9231/

.PHONY: test-reload
test-reload:
	killall -s HUP cradle_exporter
