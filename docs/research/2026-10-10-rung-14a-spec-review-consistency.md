# Rung 14a spec review, lens D (consistency)

Target: `docs/superpowers/specs/2026-10-10-rung-14a-admin-api-design.md` at `9d730d7` (cited below as `14a:<line>`).
Code read at HEAD, which matches `e6fdf2b` for `internal/`, `cmd/`, and `web/` (9d730d7 is docs-only).
Scope: correctness gaps ranked by consequence, contradictions with ratified documents, and over-ceremony. Polish and
prose are out of scope. Owner rulings 1 to 3, the seed-only `DUBPLATE_ADMIN_EMAIL` ruling, and the 14a/14b split
are not reopened.

## Brief corrections the drafter made: all five verified

| Correction | Verdict |
|---|---|
| `lease.go:166-168` (kind lease, one job per kind) | correct: the `KindLease` doc comment, "admits one job of a kind at a time across every process" |
| BS:682-684 band text | correct: 682 "M is one pass under one plan", 684 "L splits into multiple passes at its own planning"; the brief's BS:684-685 reading of M was wrong, and the spec's re-band to L is the faithful one |
| `Shell.Register` slash refusal at `serve.go:120-121` | correct: `subtreeNameRefusal`'s `strings.Contains(name, "/")` case, which `Register` (`:84-86`) panics on |
| R6:909-910 audit table | correct: "The member-change audit record stays the log record rung 5 built ... CH routes no member-audit table" |
| `handlers.go:398-407` admin contact | correct: `adminContact` returns the roster's first `m.Admin` member |

## Citation spot-check (34 checked)

Accurate as cited: BS:983; PA:118; PA:205; RE:37, :38, :223; R6:183-184, :193-194, :196-197, :204-207, :319-323,
:341-342, :580-591, :587-588, :600-602, :627-636, :770-778, :809, :815-816, :838, :929-930, :1007, :1169-1172;
CH:162-165, :167-170, :172-173, :205-207, :296-298, :576-586; SS:915, :921-923, :966, :1017-1019; RP:146-148;
`ops.go:42-53`, `:86-161`; `opshttp.go:85`; `doc.go:6-19`; `identity.go:155-161`; `session.go:80-86`, `:123-125`;
`epoch.go:98-101`; `notify.go:89`; `limit.go:58-72`; `cmd/dubplate-server/auth.go:166-173`; `serve.go:359`;
`roster.go:257`, `:298-301`, `:304-305`, `:321`, `:380-386`, `:415-417`, `:426-432`; `settings.go:67`, `:76`;
`keys.go:7-9`, `:155-160`; `thresholds.go:45-50`; `evaluate.go:265`, `:293-294`, `:306`; `0004_settings.sql:10-19`;
`arch_test.go:216`; `origin.go:37-50`; `olderstore_test.go:28-33`; `post.ts:65`; `member.go:100-107`;
`musicbrainz/client.go:85-90`; HISTORY:411-416; 6d plan :833-848; prior-art survey :166; the kin-openapi
`go doc` line; go.mod versions (kin-openapi v0.149.0, go-mail v0.8.1, x/time v0.16.0, modernc sqlite v1.60.1).

Drifted or overstated citations are collected in D10.

## Findings

### D1. blocker. The mount exposes the job operations, but `serve` has no job runner and nothing in 14a builds one

- Location: `14a:75-79` (the mount is `opshttp`'s handler over the ops seam), `14a:360-363` (AC6, AC7).
- Defect: `opshttp`'s routes are fixed to all four operations (`internal/ops/opshttp/opshttp.go:72-77`):
  `start_backup`, `job_status`, `cancel_job`, `catalog`. Mounting it in `serve` mounts the job operations. `serve`
  wires no ops implementation today (no `ops.` reference in `cmd/dubplate-server/serve.go`). R6's 6d-R6 erratum
  (R6:943-944) routes this to rung 14: "job recovery belongs to rung 14, and `serve` builds no runner at 6d." The
  spec never mentions the runner, its instance lease, or recovery at `serve`'s start. Neither does the brief or
  STATUS's carried list, so the routed item has been lost. AC6 ("A job started over `/api/v1` records the actor")
  cannot pass without a runner inside `serve`. A backup running in `serve` would also meet the serve unit's sandbox,
  which grants only `auth` and `store` (`systemd/dubplate-server-serve.service:43`). The plan would discover all of
  this mid-execution.
- Proposed fold: owner ruling 2 already decides it, so this is no fork. No 0.8 page starts a backup, and rung 16's
  timer uses the CLI verb. At 14a the HTTP projection mounts `catalog` and the seven new operations. The three job
  operations are reported in the catalog as unavailable on the cookie mount, with a reason, and R6:627-636 requires
  that asymmetry to be named in this spec. `serve`'s runner, its instance lease, and job recovery stay routed to the
  first rung whose page starts a job over HTTP (rung 9's import or rung 13). Record that routing in STATUS. Rewrite
  AC6 to prove the actor form where 14a writes it: the `settings_audit.actor` of a cookie-mount write, unchanged
  across a username change. Keep AC7 as a document-only criterion for `start_backup`'s `conflict` answer. If the
  owner wants backup over HTTP at 0.8 anyway, that is the fork, and it adds a runner task and a sandbox review.

### D2. major. Neither freshness nor the actor's `UserID` can reach the write path through the seam as designed

- Location: `14a:137-148` (re-authentication gate), `14a:186-187` (the actor's admin flag re-read inside the
  write), `14a:75-79`.
- Defect: the spec adopts `opshttp`'s design, in which the authorizer "receives no request, so a verified identity
  reaches it only through ctx" (`authorize.go:15-21`). The authorizer sees `(ctx, op)` only. Each seam method then
  receives an `ops.Actor`, which holds only an opaque wire string and a channel (`ops.go:24-27`). Two problems follow.
  - Whether `set_setting` touches a sensitive key depends on the key in the request body. The authorizer cannot see
    the body, so it cannot apply the gate. The implementation can see the key but has no session anchor.
  - Removal and demotion re-read the actor's admin flag inside the write "as the add path does". The add path keys
    that read on the actor's address (`roster.go:382`, `WHERE address = ?`). After D5 the actor's only content is
    `member:<hex UserID>`, so the implementation would have to parse its own wire string.

  The spec says freshness "comes from the verified session alone" but names no carrier. Leaving it open invites
  either a second identity read from ctx inside the implementation, which `doc.go:11-15` rules out, or wire-string
  parsing.
- Proposed fold: one clear answer, consistent with R6:583-584 ("a verified `identity.Identity` for a network caller"
  is the method parameter). `ops.NetworkActor` takes the verified identity and the session's anchor. The `Actor`
  keeps the `UserID` and the anchor in unexported fields, exposed through read-only accessors. The write path then
  re-reads the admin flag by id and applies the window to the anchor. The local actor answers "no session", and the
  gate skips it, as `14a:147-148` rules. State this in "The mount" and add it to the privileged-fields table: "a
  write's freshness: the authorizer's actor, from the verified session; never a request field".

### D3. major. The 6b-ii-a handoff's mount items are unruled, though v1 freezes them

- Location: `14a:66-90`, AC5 at `14a:357-359`, AC12 at `14a:374-376`. Handoff: `docs/HISTORY.md:2821-2839`.
- Defect: HISTORY's "To rung 14 (the mount)" lists several items the spec leaves out:
  - "a caller with no identity is rung 14's own denial kind, 401 or 404";
  - "the unauthenticated caller's response shape";
  - security L5 to L7, "audit logging of admitted operations with actor and `client_ip`";
  - whether the real authorizer lives inside `opshttp`, since the `authorizer` interface is unexported;
  - the shared-origin M2 question.

  AC5 says only that a sessionless request "reaches no operation". AC1 requires the document to list every wire code
  each operation can answer, and under the spec's own policy (`14a:49-53`) a later change to a status code is a v2
  change. The re-authentication refusal at AC12 also needs "a distinct code", which R6:589-590's closed set
  (unavailable, not found, conflict, invalid request, forbidden, canceled, internal) does not hold.
- Proposed fold: rule each item in the spec.
  - No session: 401 with the JSON error shape and a new wire code (for example `unauthenticated`), so the pages can
    send a signed-out admin to sign-in. A 404 would make a lapsed session indistinguishable from an unknown
    operation.
  - Re-authentication: a second new code (for example `reauthentication_required`).
  - Both codes are an amendment to R6:589-590's closed set. Name them as such, and the `checkStatuses` guard
    (`opshttp.go:37-44`) forces their statuses.
  - L5 to L7: one log record per admitted operation, with actor, channel, and `client_ip`.
  - The authorizer lives in `opshttp` beside the unexported interface.
  - M2: its residual is the Navidrome-era shared origin (HISTORY:3114-3119). STATUS routes it to slice 6 ("CSP-scope
    and shared-origin M2"), so the spec states that it stays there.

### D4. major. The contested-surfaces table leaves out four surfaces that both lanes write

- Location: `14a:326-335`.
- Defect: the table lists the migration number, serve registration, `web/`, and the journey vocabulary. It omits
  four more:
  - **The `internal/config` registry contract.** Hook 18 changes `config.Setting` (a closed mark type in place of
    `Sensitive bool`, plus a description), and registration then panics without them (`14a:101-102`). Rung 7
    declares its own keys (the format allowlist and the tus settings, BS:798-806; PA row `internal/config`, "every
    rung declares its keys"). A rung 7 branch that registers keys in the old shape compiles against master until
    14a merges, then fails to compile or panics at init. Its whole gate goes red at the rebase.
  - **`internal/ops`.** This covers `Seam`, the privileges table (`catalog.go:78-83`), the catalog, and the contract
    document file. The table already concedes that the journey projection is a contested enumeration, so rung 7's
    journey reads ride `/api/v1`. By the spec's own rule at `14a:82-83` they must. Both lanes then add operations to
    the same interface, the same map literal, and the same OpenAPI file.
  - **`internal/auth`.** PA's row reads "rung 5, created by it; no other lane writes it; rung 7 and rung 14 consume
    the seam". 14a writes the roster (remove, demote), the session gate, and the epoch's first caller. Rung 7
    reads the roster.
  - **`cmd/dubplate-server/serve.go`'s members wiring** beyond the `roster` line. This one is already partly covered.
- Proposed fold: add three rows.
  - **Registry contract:** 14a lands hook 18's type change in its first segment. Rung 7's plan registers every key
    with a description and mark, whichever lane merges first. If rung 7 merges first, 14a's hook-18 task marks rung
    7's keys too.
  - **`internal/ops` and the contract document:** each lane adds its operations in its own file where Go allows
    (the catalog entries), and the `Seam` interface and the document are landed by whichever lane merges first,
    with the other rebasing.
  - **`internal/auth`:** 14a writes it; rung 7 reads it only.

### D5. major. AC17's "refused" contradicts the spec's own request type and `opshttp`'s decode

- Location: `14a:84-86` ("The add operation carries no admin field") against AC17 at `14a:387` ("a body carrying an
  admin field is refused and changes nothing").
- Defect: `opshttp` decodes with a plain `json.Decoder` and no `DisallowUnknownFields` (`opshttp.go:162-163`). A
  request type with no `admin` field therefore ignores one in the body; it does not refuse it. AC5 already applies
  the ignore rule to an actor field, and R6:585-586 says "a body carrying an actor field is ignored". The roster API
  being retired does refuse (`roster.go:304-305`, `errAdminGrantForbidden`), so the behavior changes silently either
  way. An implementer meets the contradiction mid-task. That is exactly the 5b-ii fix-round class the planning-miss
  rule targets.
- Proposed fold: AC17 reads "a body carrying an admin field grants nothing: the field is ignored, as an actor field
  is (R6:585-586), and the member is added without the flag." The contract document gives `AddMemberRequest`
  `additionalProperties: false` only if the test should also catch the client sending one. The members page never
  sends one.

### D6. minor. D5 (the actor on `UserID`) supersedes ratified text without naming it

- Location: `14a:92-97`.
- Defect: R6's 6d-R2/R14 erratum (R6:913-915) ratified that `Principal` stays the username because "the username is
  the readable name in logs and the network actor." The `identity.Identity.Principal` doc comment the spec cites
  (`identity.go:209-211`) also says the principal is the name "an audit actor ... knows the member by". The spec
  cites the half of that comment that supports it. It does not name the R6 sentence it overturns or the comment
  that 14a must edit. The HISTORY 6d routed item (HISTORY:394) supports the change, so the decision stands; only
  the record is incomplete.
- Proposed fold: one sentence in D5: "This supersedes R6:914-915's 'and the network actor'; Task 2 edits
  `identity.go:209-211` and `ops.go:42-46` to match." Also note that `settings_audit.actor` will hold the hex form,
  which needs a join to read. That is acceptable, because no 14a surface reads the audit.

### D7. minor. D9's runbook claim does not match the runbook, and AC21 contradicts itself

- Location: `14a:239-243`, AC21 at `14a:396-397`.
- Defect:
  - The spec says "The runbook's bootstrap step adds `member add <DUBPLATE_ADMIN_EMAIL> --admin`" and cites the 5d
    plan (`:1399`). That plan step is stale: it uses `--streamer-username` and the retired roster leaf. The live
    runbook (`docs/server-runbook.md:164-170`) bootstraps with `admin@example.org`. Its key description at `:59-60`
    calls `DUBPLATE_ADMIN_EMAIL` "the operator's contact address" for MusicBrainz and gives it no seed role.
  - AC21 says "No code path reads `DUBPLATE_ADMIN_EMAIL` to grant or check admin", then requires `doctor` to check
    whether the address names a live admin.
  - `doctor` already prints a `DUBPLATE_ADMIN_EMAIL` row whose "ok" means the value parses (`doctor_test.go:1023`),
    and a `members` row that fails with no admin.
- Proposed fold:
  - AC21 reads "to grant or authorize admin".
  - 14a's scope names the two runbook edits: install step 10 uses the `DUBPLATE_ADMIN_EMAIL` address, and the key
    description gains its seed role.
  - The spec says which row carries the warning. Recommendation: the existing `DUBPLATE_ADMIN_EMAIL` row, which
    stays "ok" when the address parses and names a live admin, and warns when it parses but names none.
  - Cite the runbook, not the 5d plan.

### D8. minor. `reset_setting` on a running `serve` puts the default in force, not the seed

- Location: `14a:129-130`, AC11 at `14a:372-373`.
- Defect: "deletes an `admin` row, so the next start seeds the key from the environment or resolves its default."
  Between the reset and the next start, the accessor finds no row and returns `s.Default`
  (`internal/settings/accessor.go:113-115`), even when the serving environment sets the key. For an
  immediately-applied sensitive key such as `DUBPLATE_SESSION_ABSOLUTE_CAP`, the default goes live at once. The
  write answer would describe a seed that has not happened. Writing an `env-seed` row from the reset is not
  available, because R6:1002 gives that source to the seed rule alone.
- Proposed fold: the reset's answer and the settings page say "the default `<d>` is in force until the next start,
  which re-seeds `<env value>` from the environment" whenever the environment sets the key. AC11 gains the case.

### D9. minor. kin-openapi behaviors that AC1 and AC2 rely on are not quoted

- Location: `14a:41-47`, AC1 and AC2 at `14a:348-353`.
- Defect: the spec quotes only "Supports OpenAPI 3.0, OpenAPI 3.1, and OpenAPI 3.2". The same `go doc` page says
  "For OpenAPI 3.1 validation, use the JSON Schema 2020-12 validator option" (`openapi3.EnableJSONSchema2020()`),
  which `openapi3filter` takes through `Options.SchemaValidationOptions`. `Options.IncludeResponseStatus` defaults
  to false: "Set IncludeResponseStatus so ValidateResponse fails on response status not defined in OpenAPI spec."
  Without it, AC1's "every wire code it can answer" is unproven. An undocumented 401 or 409 validates green, which
  is the vacuous-pass class the planning-miss rule names.
- Proposed fold: quote both behaviors in "The contract document". AC2 states that validation runs with the
  2020-12 option and `IncludeResponseStatus`, and its proof adds "an undocumented status fails validation".

### D10. minor. Drifted or overstated citations

- `catalog.go:76-81` (`14a:78`): the privileges map is at `:78-83`. Lines 76-77 are its doc comment.
- R6:185-187 (`14a:36`) gives the first collection row to "rung 7 or rung 10". The split's condition should read
  "after the first rung that writes a collection row".
- CH:147-150 (`14a:84`) is the two-mount paragraph. The TUI sharing the admin surface is CH:139-142.
- BS:986-989 (`14a:412`): the token and service-token text is BS:985-988.
- RP:110-111 (`14a:285`) names hooks for rungs 7, 9, 13, and 14 only. It does not support "rung 6, 8 ... or a
  streamer slice". Cite SS's per-hook text instead.
- Ruling 8 (`14a:83`) is the extensibility fork ("built-ins + lifecycle hooks + the documented API; no plugin
  ABI"). It does not require every member endpoint to sit in the contract document. The 0.9 TUI reason carries the
  decision alone, so drop the ruling 8 clause or soften it to "in the spirit of".
- `roster.go:431` (`14a:235`) is `errNoAdmin`'s message text. The grant rule is `:304-305`.

### D11. minor. The prior-art table omits goose

- Location: `14a:287-297`.
- Defect: CLAUDE.md's prior-art list names goose, and the rung plan must adopt or decline each entry with a reason.
  The table covers every other entry.
- Proposed fold: add "goose | neutral | no migration at 14a".

### D12. minor. Reverse edge (14,13) is left open, though the spec's own rule answers it

- Location: `14a:81-86`. Edge: `research/2026-09-04-reverse-edge-completion.md:39` (SOFT-C, "Whether rung 13
  exposes its queue verbs over `/api/v1`, and who writes those handlers, is unwritten").
- Proposed fold: one line. Rung 13's queue verbs that have a page caller join `/api/v1` as catalog operations in the
  contract document, written by rung 13 in its own files, under the additive policy.

### D13. minor. The resolving verb's open is a fourth store mode that R6 does not have

- Location: `14a:163-168`.
- Defect: R6 defines the inspect, write, and serve modes. Write mode refuses any version but the binary's
  (R6:319-323), and serve mode is the only one that migrates. "The plan picks the open mechanism", but whatever it
  picks is a new write-capable mode over an older store. That amends R6's mode design and should be named as an
  amendment.
- Proposed fold: state the amendment. The mode never creates or migrates a store. It accepts a store at version 4
  through the binary's version. It refuses a newer store and a restore intent with the existing messages. It
  writes only the three named tables.

### D14. minor. Bump edge cases the removal must handle

- Location: `14a:203-209`, AC15 at `14a:382-384`.
- Defect:
  - `Bump` returns `ErrEpochBackward` when the stored epoch is later than now (`epoch.go:83-87`). Under "a bump
    failure rolls the removal back", a removal after a backward clock step is refused outright. Yet every
    credential anchored at or before now is already revoked by the later stored epoch, so the refusal protects
    nothing.
  - `Bump` truncates to the second and revokes everything "anchored or minted at or before now". AC15's "a sign-in
    after the re-add verifies" fails when the sign-in lands in the bump's second.
- Proposed fold: treat `ErrEpochBackward` as success inside the removal, since the stored epoch already covers it.
  AC15's fixture uses an injected clock that places the post-re-add sign-in at least one second after the bump.

### D15. minor. "Settings page" collides with the identity spec's member settings page

- Location: `14a:23`, AC20 at `14a:393`.
- Defect: the visual identity spec (`specs/2026-09-24-visual-identity-design.md:145-146`) says "Returning to 'follow
  the system' belongs to a later settings page". That page is a member's preferences page. 14a's page is
  admin-only, behind `RequireAdmin`.
- Proposed fold: name the route and the page "admin settings" (for example `routes/admin/settings`), and leave
  `settings` for the member preferences page.

## Over-ceremony

No standalone over-ceremony finding. The review reads, the hook-18 coverage of wall rows and integration keys (SS
hook 18 asks for "every registered config key"), and the older-store resolving verb (which the upgrade path needs)
are each warranted. D1's fold removes the most clock time: it keeps a `serve` runner, a recovery pass, and a sandbox
review out of 14a. Building those now, against owner ruling 2, would cost about two `auth-data` tasks (roughly 1.2M
tokens and 40 minutes at the brief's rates) for a surface with no 0.8 caller.

## Summary

Fifteen findings: one blocker, four major, and ten minor. All five brief corrections check out, and 34 citations
were spot-checked. The cited lines mostly say what the draft claims, and seven drifted or overstated ones are
collected in D10.

The top three:
- **D1 (blocker).** The cookie mount would expose `start_backup`, `job_status`, and `cancel_job`, but `serve` has no
  job runner. R6's 6d-R6 erratum routes `serve`'s runner and job recovery to rung 14, and the spec, the brief, and
  STATUS all dropped that item. Owner ruling 2 resolves it: the job operations report as unavailable on the mount,
  and the runner routes to the first rung whose page starts a job.
- **D2 (major).** The `opshttp` authorizer never sees the request body, and `ops.Actor` holds only a wire string.
  Neither the session anchor for the re-authentication gate nor the actor's `UserID` for the in-write admin re-read
  has a carrier. The fold extends `NetworkActor` to carry both.
- **D4 (major).** The contested-surfaces table leaves out the `internal/config` registry contract that hook 18
  changes under rung 7's concurrent key registrations, and `internal/ops` with the contract document, which both
  lanes extend. It also leaves out `internal/auth`, which PA says no other lane writes.
