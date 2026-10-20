# Rungs 12 and 16 re-plan: contract and setup-ease review

Target: `docs/superpowers/specs/2026-10-05-rungs-12-16-replan-design.md` at `88c535e`.
Lens: contract and criteria (is every promise testable, does each criterion name its fixture and its failure),
plus ruling 7 (new-operator setup ease). Every claim below was checked against the tree at `88c535e`.

Findings are ranked by consequence. Counts: 0 blocker, 5 major, 7 minor, 2 over-ceremony.

## What the tree says a stranger configures today

`config.Load` requires 14 values for a default install with no integration enabled:

- the owned wall's four rows (`internal/config/keys.go:20-25`);
- `DUBPLATE_IDENTITY_SOURCE`, an owned row with no Pending (`cmd/dubplate-server/auth.go:21`);
- the magic-link source's nine Required keys (`cmd/dubplate-server/auth.go:43-55`).

`DUBPLATE_TRUSTED_PROXY_ENABLED` unset means direct mode (`cmd/dubplate-server/proxy.go:19-23`), so it costs
a stranger nothing. The Navidrome rows the spec retires are reserved (`keys.go:81-105`) or superseded
(`keys.go:137-146`). Both kinds are load errors when set, so no stranger ever set one. The spec's removals
therefore change the stranger's required set by zero. The shadow class is opt-in behind its flag
(`cmd/dubplate-server/shadow.go:28-41`), so deleting it is also zero. The real growth risk is the streamer's
own configuration (C1), which the spec leaves open.

## Correctness gaps

### C1 (major): the streamer's configuration shape is left on a footing that adds required keys

Location: spec:135 ("The streamer's own keys arrive with the slice that reads them"); spec:175-182 (the
streamer spec revision 3 edit list); streamer spec `2026-09-07-streamer-2-0-design.md:630-643`.

Defect: the streamer spec's "Cutover and operations" section says configuration "adds an integration class in
rung 4's shape: enabled, cache cap, concurrent encodes, the profile set, and the app-password key, each required
when enabled." That shape fit an optional second driver. Under ruling 1 the streamer is the product from the
first day. Kept as written, every stranger sets a flag plus four required values, one of them a secret to mint.
That is five new setup values, which ruling 7's first bullet forbids. Revision 3's edit list removes the shadow,
the flip, `shareRole`, and the escape hatch. It never touches this sentence, so slice 3's or slice 5's plan
inherits it.

Fold: add a ruling 1 consequence and a revision 3 bullet. The streamer is always on and declares no enable flag.
Its tunables (cache cap, concurrent encodes, profile set) are registered thresholds with defaults. Any streamer
secret, such as the app-password key, is minted at first start into the state directory or by `init`, never
required in the environment. Each streamer slice plan states its delta to the stranger's required-key set.

### C2 (major): rung 16 omits the ingress items a TLS listener needs on the box

Location: spec:75-79 (rung 16) and spec:191-211 (the carried-items table).

Defect: the spec says "The box reaches the internet through slice 8's TLS listener, since the tunnel is gone."
The tree blocks that path in four places, and the table lists none of them:

- `scripts/firewall-rules.json` opens only tcp/22.
- `systemd/dubplate-server-serve.service` sets `AmbientCapabilities=` and `CapabilityBoundingSet=` to empty
  under `User=music`, so the unit cannot bind 80 or 443.
- The unit's `ReadWritePaths` holds only `auth` and `roster`. autocert keeps its certificates under the state
  directory (streamer spec:536-538), so a certificate leaf needs a `ReadWritePaths` entry, a pre-create in
  `deploy.sh`, and a row in the state-dir agreement leg.
- The hostname's DNS record was deleted with the tunnel (CLAUDE.md, Cloudflare). `DUBPLATE_PUBLIC_URL` in
  `env/dubplate-server.env.template` still names it.

The first two also apply to any stranger. The streamer spec's "ports 80 and 443" says nothing about binding
them as an unprivileged user.

Fold: add four table rows. The unit capability (`CAP_NET_BIND_SERVICE` in both capability lines, TLS mode only)
and the certificate leaf go to slice 8, because the stranger meets them. The firewall's 80 and 443 rules and
the DNS record (by API) go to rung 16, because only the group's estate meets them.

### C3 (major, OWNER FORK): the stranger walk cannot happen "on the box", and its defects have no disposition

Location: spec:48-49 (ruling 7, last bullet) and spec:75-79.

Defect: the box is not a stranger's host. `deploy.sh` provisions it, renders `/etc/dubplate/server.env` from
the workstation age store through a template with estate values (`DUBPLATE_SITE_NAME=907`,
`https://music.907.life`), and pre-creates its leaves. A walk of `dubplate init` there either fights that
render (two writers of one configuration) or skips `init` and proves nothing about a stranger. The walk also
runs at rung 16, after slice 8 has closed. "Records every step ... as a defect against slice 8" says nothing
about whether a defect blocks launch, so the criterion passes with any number of defects recorded.

Options:

- **(a)** Slice 8's own close gate is a walk on a fresh host (a throwaway VM or small VPS, no age store, no
  `deploy.sh`), from the release artifact through `init` to a played track. Its defects block slice 8. Rung 16
  keeps the estate path through `deploy.sh`.
- **(b)** As written: the walk runs on the box at rung 16, and its defects are filed against slice 8.
- **(c)** The group's own estate installs through `init`. `deploy.sh` keeps only binary delivery and units,
  and the template's literals retire, so one configuration path serves both.

Recommendation: (a), with a pass/fail line in slice 8's acceptance: an unclear step either blocks or goes to
`IDEAS.md` with an owner ruling. Option (c) is the stronger ease-of-setup guarantee, but it re-plans
`deploy.sh` and the age-store render, so it belongs in the packaging initiative's scope.

### C4 (major): no existing verb can migrate the box's roster row

Location: spec:152-154.

Defect: the spec requires the box's one admin row to move "through an existing verb or the bootstrap path,
never by hand." The roster loader decodes with `json.RejectUnknownMembers(true)`
(`internal/auth/roster.go:340`). Once the field is renamed, a file still carrying `streamer_username` fails to
open. `member add` opens the existing roster first (`cmd/dubplate-server/member.go`, `addMember`, through
`auth.OpenRoster`), so it fails as well. The bootstrap path runs only when the file is absent
(`ErrRosterMissing`). Every sanctioned path therefore begins with deleting the file, which is the by-hand step
the spec forbids. The plan cannot "name which" verb, because none exists.

Fold: say so in the spec and pick the path. The deploy removes the roster file and re-bootstraps with
`member add --admin`, as one recorded runbook step run over ssh after the quiet-box check. This is safe
because `serve` is installed but disabled (STATUS). The criterion: `member list` on the box shows the one
admin under the new key, and `doctor` exits 0.

### C5 (major): three of the removal pass's six criteria have no grader or can pass vacuously

Location: spec:107-116.

Defects:

- "Each remaining type, function, and parameter earns its place" (spec:109) and "the rendered `go doc` reads as
  a coherent package" (spec:113) name no grader and no failing state. Ruling 6 names `go-architecture-reader`,
  but the acceptance list does not bind to it.
- "`doctor`'s web row and probe route behave as before, proved by their existing tests" (spec:116) passes
  vacuously if the same diff edits those tests.
- Two removals change behavior, and no criterion names the test that would catch a leftover:
  - With the fallthrough gone, `Handler` answers an unregistered path with `http.NotFound` (spec:101-102).
  - With the bypass gone, a cross-origin unsafe-method request carrying a bearer token is refused, where
    `addBearerBypass` used to exempt it (`internal/serve/origin.go:88`; `origin_test.go:127` exercises it).

Fold: replace the two prose bullets with checkable ones.

- `go doc -all ./internal/serve | grep -inE 'front|driver|fallthrough|bypass|exemption'` prints nothing.
- The staged exported-unused leg reports no hit in `internal/serve`.
- The package's `go-architecture-reader` verdict carries no "unused surface" finding.
- The doctor web-row and probe tests are named and appear in the diff only for renames.
- Two tests are named, each failing against `88c535e`'s wiring: an unregistered path answers 404 with no
  session-cookie handling, and a cross-origin POST with a bearer token to a former bypass path is refused.

### C6 (minor): ruling 7 has no checkable form, and its "shrinks" bullet is vacuous

Location: spec:40-49.

Defect: bullet 2 says the removals shrink the configuration surface. They do not (see the opening section):
every retired row was already refused when set. Ruling 4 also leaves `DUBPLATE_IDENTITY_SOURCE` as a required
row whose only accepted value is `magic-link`, now that the forward-auth front moves to `IDEAS.md`. That key is
a real, cheap reduction, and the spec does not take it.

Fold: restate bullet 2 to say "removes keys a stranger could mistake for setup", and add a measurable bullet.
The stranger's required-key set for a default install is 14 names at `88c535e`, and no slice grows it without an
owner ruling. Optionally, one golden test over the registry's required names makes growth a visible diff. Route
"default the identity-source selector to `magic-link`" to 6d or slice 8 as a decision, with the reason it
stays required if it does.

### C7 (minor): the rename list misses files, including the stranger-facing example roster

Location: spec:141-146.

Defect: `streamer_username` also appears in `config/roster.example.json` (the file a stranger copies),
`web/scripts/rendered/run.ts`, `internal/backup/*_test.go`, `internal/auth/*_test.go`,
`cmd/dubplate-server/*_test.go`, and store fixtures `0004` and `0005`. The `--streamer-username` flag
(`member.go`) is covered by ruling 3 but missing from the list. `TestRosterExampleFile` would catch the example
roster. The rendered script and the backup tests would only fail at runtime or under the rendered gate.

Fold: replace the enumerated list with a criterion.
`git grep -niE 'streamer.?username' -- ':!docs' ':!_salvage'` returns nothing.

### C8 (minor): the shadow comment-rewrite list is incomplete

Location: spec:123-124.

Defect: besides `unidecode.go` and `serve.go`, the shadow command is cited as reason or precedent in five more
places:

- `Makefile:5-7`
- `scripts/deploy.sh:225`
- `cmd/dubplate-server/main.go:11,42,156`
- `internal/auth/store.go:28,120`
- `internal/config/keys.go:195-197`

`shadow.go` is the only production importer of `internal/harness/*`.
`go list -deps ./cmd/dubplate-server` lists `internal/engine/shadow`, `internal/harness/cassette`, and
`internal/harness/analysis` today. `matcherCommit` survives the deletion, because `version` prints it.

Fold: two criteria. First, `git grep -ni shadow -- cmd internal scripts Makefile` returns nothing. Second,
`go list -deps ./cmd/dubplate-server | grep internal/harness` is empty, so the server binary no longer links
harness code.

### C9 (minor): front and Navidrome text survives outside `internal/serve`

Location: spec:111-112 (scoped to `internal/serve`) and spec:129-133.

Defects:

- `internal/auth/seam.go:20-22` documents the `Seam` interface by "a forward-auth front is an alternate".
- `keys.go:133-138` tells an operator that `DUBPLATE_ACCESS_*` moved to "the forward-auth front's own config;
  the seam admits one", and that `DUBPLATE_PUBLIC_NAVIDROME_URL` is "arriving at rung 12". Both become false
  under ruling 4.
- The spec says the superseded rows "retire" without saying whether they are deleted (then classified
  "unknown") or restated as retired. `DUBPLATE_TUNNEL_TOKEN` keeps a "retired:" message as precedent.

Fold: extend the "no sentence about a front, a driver, or a fallthrough" criterion to `internal/auth` and
`internal/config`. Name the disposition: delete the reserved `streamer` class and the ping row, which never
shipped, and restate the four superseded messages as "retired: dubplate integrates with no external streamer
or authentication front" (a judgment call; deleting them is also defensible, provided it is named).

### C10 (minor): rung 16's attended gate and rollback name no fixture

Location: spec:75-79.

Defects:

- "Playback in a real music app" names no app. The client roster's first entry is the natural fixture.
- "Its rollback is the runbook's restore from backup" does not fit a first install. Before the first backup
  there is nothing to restore, and the real rollback is stopping and disabling the `serve` unit.
- The archive import invites every Navidrome user by magic link. If the bootstrapped admin and a Navidrome user
  share an address or a username, the import meets the duplicate-username refusal
  (`errDuplicateStreamerUsername`, `roster.go:109`), and the streamer spec does not say what it does.

Fold: name the app and the album fixture. State the rollback as "disable `serve`; data stays". Route
"import meets an existing member" behavior to slice 8's plan as a named case. The recommended behavior is to
match by address and keep the existing row.

### C11 (minor): command names contradict the binary

Location: spec:24, 48, 77 (`dubplate import navidrome`, `dubplate init`).

Defect: the binary is `dubplate-server`. STATUS already owes "settle `init`/`upgrade` naming against
`dubplate-server`". A stranger's setup guide depends on the answer, and the spec routes it nowhere.

Fold: add a row to the carried-items table: command naming goes to slice 8.

### C12 (minor): no ingress exists between now and slice 8

Location: spec:67 and spec:75-79.

Defect: with the tunnel gone and only tcp/22 open, nothing on the box is reachable over HTTP until rung 16.
Rung 7's uploader before/after is an owner touchpoint, and any render-and-read before slice 8 has to run
somewhere.

Fold: one sentence. Attended web checks before slice 8 run on the workstation, or through
`ssh -L` to the box. The box opens no public port before rung 16.

## Over-ceremony

### O1 (minor): two web reviewers for a label rename

Location: spec:163.

Cost: two reviewer dispatches over a diff that changes one label, one hint, and a deck string. CLAUDE.md fans
out by surface, so some review is owed. Recommendation: dispatch `daisyui-a11y-reviewer` alone, scoped to the
label and hint association and the rendered gate's result. Add `svelte-reviewer` only if the diff touches the
script block.

### O2 (minor): a separate document lane for five file edits

Location: spec:170.

Cost: one lane launch (registry entry, guard, sub-ceiling) for edits to the build spec, the streamer spec,
`CLAUDE.md`, ROADMAP, and the ledgers. Recommendation: run the streamer spec's revision 3 as its own document
task, because C1 makes it substantive, and fold the remaining four edits into the removal pass's close. Optional.
