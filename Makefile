BINARY := dubplate-server

# LDFLAGS defaults to empty so a bare `make build-linux` from a
# workstation still produces a "dev"-stamped binary (main.go's own
# version and matcherCommit defaults), which keeps shadow observe's
# refusal firing for workstation builds. deploy.sh is the only caller
# that sets it, passing the two -X main.version=/-X main.matcherCommit=
# flags a real release stamps.
LDFLAGS :=

build:
	go build -o $(BINARY) ./cmd/dubplate-server

build-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o $(BINARY)-linux-amd64 ./cmd/dubplate-server

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
# outside package oracle. -timeout is explicit and generous (90m): the
# full oracle package run already takes ~585s and only grows as the
# corpus does, past go test's 10m per-binary default.
derive:
	go test -tags oracle -timeout 90m ./internal/harness/... ./cmd/harness/...

clean:
	rm -f $(BINARY) $(BINARY)-linux-amd64

.PHONY: build build-linux test vet lint install tidy-check check derive clean
