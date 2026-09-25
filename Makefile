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
# no hardcoded color, no external origin, font delivery, theme
# variable completeness, theme contrast, and the built bundle's
# stamped theme values, and no root-absolute URL). It builds the bundle twice and fails unless
# the two build trees match file for file, so the build is
# reproducible. `check` runs it first because the Go tests that read
# the embedded bundle need it current before `go test` runs; the
# rendered (browser) gates are the separate web-rendered target.
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
	build_first="$$(mktemp)"; \
	build_second="$$(mktemp)"; \
	trap 'rm -f "$$build_first" "$$build_second"' EXIT; \
	(cd web/build && find . -type f -print0 | LC_ALL=C sort -z | xargs -0 sha256sum) >"$$build_first"; \
	npm run build --prefix web; \
	(cd web/build && find . -type f -print0 | LC_ALL=C sort -z | xargs -0 sha256sum) >"$$build_second"; \
	if ! diff "$$build_first" "$$build_second" >&2; then \
		echo "make web: two builds of the same source differ (the lines above); kit.version.name in web/svelte.config.js must be deterministic." >&2; \
		exit 1; \
	fi; \
	echo "make web: two builds of the same source match ($$(sha256sum <"$$build_first" | cut -d' ' -f1))."; \
	npm run check --prefix web; \
	(cd web && node scripts/gates/server-files.ts); \
	(cd web && node scripts/gates/g1.ts); \
	(cd web && node scripts/gates/g3.ts); \
	(cd web && node scripts/gates/g4.ts); \
	(cd web && node scripts/gates/g2.ts); \
	(cd web && node scripts/gates/g9.ts); \
	(cd web && node scripts/gates/theme-bundle.ts); \
	(cd web && node scripts/gates/root-absolute.ts)

# RENDERED_TIMEOUT bounds the rendered test binary. It is explicit
# because go test's ten-minute default would bind long before a stuck
# browser is noticed, and it stays at least twice the leg's measured
# wall time (about 17s for the test, 20s with its compile).
RENDERED_TIMEOUT := 3m

# web-rendered is the rendered leg: cmd/dubplate-server's
# rendered_test.go, under its build tag, serves the real shell on
# loopback and drives Playwright's own Chromium against it through
# web/scripts/rendered/run.ts. It vets and lints the tagged file,
# since the default vet and lint targets never compile it. `go test`
# exits 0 with "no tests to run" when -run selects nothing, so the
# -json stream goes to a file and test-events.ts fails unless it
# carries TestRendered's own pass event; the target keeps go test's
# exit status either way. It reads the bundle `web` built, so it runs
# after `web`.
web-rendered:
	go vet -tags rendered ./cmd/dubplate-server
	golangci-lint run --build-tags rendered ./cmd/dubplate-server/...
	set -e; \
	events="$$(mktemp)"; \
	trap 'rm -f "$$events"' EXIT; \
	status=0; \
	go test -json -tags rendered -count=1 -timeout $(RENDERED_TIMEOUT) -run '^TestRendered$$' ./cmd/dubplate-server >"$$events" || status=$$?; \
	(cd web && node scripts/rendered/test-events.ts "$$events") || status=1; \
	exit "$$status"

check: web vet lint test web-rendered tidy-check

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

.PHONY: build build-linux test vet lint install tidy-check check web web-rendered derive clean
