BINARY := dubplate-server

build:
	go build -o $(BINARY) ./cmd/dubplate-server

build-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o $(BINARY)-linux-amd64 ./cmd/dubplate-server

test:
	go test ./...

vet:
	go vet ./...

lint:
	golangci-lint run ./...

install:
	go install ./cmd/dubplate-server

tidy-check:
	go mod tidy
	git diff --exit-code -- go.mod go.sum

# The web legs rejoin at rung 7, when web/ is rebuilt against the
# greenfield server.
check: vet lint test tidy-check

# derive runs the oracle-tagged tests, the only ones that launch the
# real Python shim and so the only ones needing the uv venv
# (tools/venv/README.md). Deliberately outside check: the gate protects
# the frozen corpus records and never touches Python. cmd/harness/...
# is named explicitly because its own oracle-tagged tests (harness
# freeze's real-venv coverage, and future case-authoring tools) sit
# outside package oracle.
derive:
	go test -tags oracle ./internal/harness/... ./cmd/harness/...

clean:
	rm -f $(BINARY) $(BINARY)-linux-amd64

.PHONY: build build-linux test vet lint install tidy-check check derive clean
