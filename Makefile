BINARY := dubplate
WEB_DIR := web

build:
	go build -o $(BINARY) ./cmd/dubplate

build-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o $(BINARY)-linux-amd64 ./cmd/dubplate

test:
	go test ./...

vet:
	go vet ./...

lint:
	golangci-lint run ./...

install:
	go install ./cmd/dubplate

tidy-check:
	go mod tidy
	git diff --exit-code -- go.mod go.sum

# web-install is every other web-* target's prerequisite rather than a
# step check runs on its own: `npm ci` is idempotent and each of
# svelte-check/vitest/vite build needs node_modules present regardless
# of which other web-* targets already ran this invocation.
web-install:
	cd $(WEB_DIR) && npm ci

web-check: web-install
	cd $(WEB_DIR) && npm run check

web-test: web-install
	cd $(WEB_DIR) && npm run test

web-build: web-install
	cd $(WEB_DIR) && npm run build

check: vet lint test tidy-check web-check web-test web-build

clean:
	rm -f $(BINARY) $(BINARY)-linux-amd64

.PHONY: build build-linux test vet lint install tidy-check web-install web-check web-test web-build check clean
