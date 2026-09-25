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

# The web legs run today: `web` requires Node and npm, refuses a stale
# or drifted `npm ci --prefix web`, builds the SvelteKit static bundle,
# then runs svelte-check and the static web gates (no server-only file,
# no hardcoded color, no external origin). The rendered (browser) gates
# are not yet part of this leg. `check` runs it first because the Go
# tests that read the embedded bundle need it current before `go test`
# runs.
web:
	set -e; \
	if ! command -v node >/dev/null 2>&1; then \
		echo "make web: node not found on PATH; install it via mise (mise use -g node@lts)." >&2; \
		exit 1; \
	fi; \
	if ! command -v npm >/dev/null 2>&1; then \
		echo "make web: npm not found on PATH; install it via mise (mise use -g node@lts)." >&2; \
		exit 1; \
	fi; \
	node_version="$$(node --version)"; \
	node_major="$${node_version#v}"; \
	node_major="$${node_major%%.*}"; \
	case "$$node_major" in \
		''|*[!0-9]*) \
			echo "make web: could not parse a node major version from '$$node_version'; 24 or later required." >&2; \
			exit 1; \
			;; \
	esac; \
	if [ "$$node_major" -lt 24 ]; then \
		echo "make web: node major $$node_major found, 24 or later required." >&2; \
		exit 1; \
	fi; \
	if [ ! -d web/node_modules ]; then \
		echo "make web: web/node_modules absent; run \`npm ci --prefix web\`." >&2; \
		exit 1; \
	fi; \
	npm_ls_log="$$(mktemp)"; \
	if ! npm ls --prefix web >"$$npm_ls_log" 2>&1; then \
		cat "$$npm_ls_log" >&2; \
		rm -f "$$npm_ls_log"; \
		echo "make web: npm ls --prefix web found a missing or out-of-range install; run \`npm ci --prefix web\`." >&2; \
		exit 1; \
	fi; \
	rm -f "$$npm_ls_log"; \
	(cd web && node scripts/gates/install-state.ts); \
	npm run build --prefix web; \
	npm run check --prefix web; \
	(cd web && node scripts/gates/server-files.ts); \
	(cd web && node scripts/gates/g1.ts); \
	(cd web && node scripts/gates/g3.ts)

check: web vet lint test tidy-check

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

.PHONY: build build-linux test vet lint install tidy-check check web derive clean
