# Photo curation spec: fold record

Target: `docs/superpowers/specs/2026-09-25-photo-curation-design.md`, revision 1 at `8843790`,
folded to revision 2. Reviews: `2026-09-25-photo-curation-review-contract.md` (C),
`-mechanics.md` (M), `-integrity.md` (D), `-consistency.md` (K).

Fold rule: fold what prevents real harm (a wrong-licensed image in the binary, a pick landing on
the wrong image, lost work, a legally wrong claim, an unreachable finish line). Refuse a finding
whose fix costs more than the risk it removes. Nothing here edits `scripts/check.sh`,
`Makefile`, or `web/`, which lane 1 holds.

Counts: 56 findings (C 15, M 14, D 12, K 15). 54 folded, eight of them only in part with the
refused part and its reason stated (C2 group, C5, C6, C14, M5/D4, D5, D3, K15); K9 routed to
errata; M3 the one owner fork (Ruling 1). One sub-item, K15's future-website line, is refused.

## Verification done by the fold

- `git commit -- <untracked path>` fails ("pathspec did not match any file(s) known to git"),
  so D2's `git commit --only -- <paths>` alone cannot commit a new image. Probed in a scratch
  repo: `git add -- <paths>` then `git commit -- <paths>` commits only those paths and leaves
  another path staged. The spec uses that pair.
- `.gitleaks.toml` allowlists `.claude/worktrees/`, `.superpowers/`, and `web/.svelte-kit/`, so a
  new in-tree candidate directory needs its own entry. The 5c-iii plan hands "the photo column
  under Ruling 3" to rung 7 (`:1774`, `:1815`) while its Ruling 3 text (`:150`) routes the
  two-column split to the curation pass. The lane registry is a table in `docs/STATUS.md`.
- The identity spec names Lesser and the British Culture Archive as the strongest reggae leads
  (`:362-364`) and lists "outsider artists" and "DC hardcore flyers first" (`:273`, `:278`).

## Dispositions

| IDs | Disposition | Where / reason |
|---|---|---|
| C1 | Folded | Pool size and balance: plus or minus 3 point bands, waivers in `photos/meta/waivers.json`, waiver proposals after two empty targeted rounds, bundled into one checkpoint question; **Done** paragraph. The pool-floor number is not a fork: Geoff set the ~300 target, and done is his declaration at a clean balance report. The band width is taken, since an exact share is unreachable and a waiver covers taste. |
| C2, M4, D7, K7 | Folded in part | Storage: `tools/photos/check.mjs`, integrity and balance modes, fixtures proving it fails. D7's and K7(b)'s gate leg in this pass is refused for now: `scripts/check.sh` is lane 1's contested surface, so the leg is a queued edit landing after lane 1 finishes (Process), and the conductor runs integrity before every photo commit until then. Balance cannot gate every commit because the pool fills over many commits. |
| C3, K12 | Folded | Family from primary scene; floor counts first-listed subject only; region cap over records with a place; "a landmark" moved out of the subject list. |
| C4, K13 | Folded | Posed at or below 15% of picked photographs, waivable. Taken, not a fork: the identity spec already rules "scenes beat posed portraits", and a waiver lets Geoff overrule the number. |
| C5 | Folded | One image per named artist, keyed by MBID or normalized name; listed in Changes. The closed `artists` list is refused: a normalized-name key catches the same duplicates without a conductor-maintained list. |
| C6 | Folded in part | Landmark key is Wikidata ID or normalized name, with waivers. The closed landmark list in `vocab.json` is refused: a slipped second image of one landmark is a balance nit Geoff sees on the sorting page, not worth a per-round list edit. |
| C7, D10, K4 | Folded | `artwork` field, required for flyer and zine cover, with the faithful-scan basis. |
| C8 | Folded | Picked license limited to shipped classes by the integrity check; `public_domain` shape, required only for `LicenseRef-public-domain`. The basis values live in a closed list in `vocab.json` rather than in the spec, so no US term rule is stated from memory. |
| C9, D6 | Folded | Sample size (every PD and flyer record, 20% of the rest, at least five); round logs; `file_url`, `archive_url`, `retrieved_at`; `verified` names who verified. Grants: letter asks for a public posting; an email-only grant commits a grant record (D6 option a, taken as settled: the record holds only what the grantor wants credited, and the repository is the durable record). |
| C10, K8, D9 | Folded | Non-picked records move to `photos/meta/` (rejected as dedupe keys plus a reason code, pack as full records); `screen` is stripped at commit; pair check covers picked only; status absent only in the candidate home. |
| C11, M1, K2 | Folded | Licensing: per-version basis (4.0 section 2(a)(4); 2.0 and 3.0 by the CC format-change guidance), CREDITS notes the resize, `source_title` for 2.0 and 3.0; the Changes bullet now says the file is modified. |
| C12, K15 (dimensions) | Folded | `width`, `height` of the stored file; crop judgment left to Geoff's sort. Source dimensions are not recorded: Ruling 1's floor is checked at screening. |
| C13 | Folded | Year precision: `exact`, `circa`, `decade`. |
| C14 | Folded | The final-night clause stays as guidance with the `error` surface marking it; error and empty floors of three each. `closed-venue` subject refused: the surfaces facet already records the use. |
| C15 | Folded | Agents set `pack` at screening; the page does not show pack candidates. |
| M2 | Folded | Licensing: raw `curl` evidence under a descriptive UA, `action=raw` on Commons, license link on Flickr; the spot-check uses the same channel. |
| M3 | Owner fork | Ruling 1 (size floor below 1600px). Upscaling (option c) is refused as the review recommends. |
| M5, D4 | Folded | `source_id`, `origin_id`; per-agent candidate subdirectories; no overwrite; conductor mints the slug at pick time with a uniqueness check; slug pattern; slugs never renamed; the check reports unparseable JSON. D4's temp-file-and-rename step is refused: per-agent directories already rule out a writer collision, and the check catches a truncated file. |
| D5 | Folded in part | Upstream-most copy governs, and conflicting copies reject (Licensing); `source_id` and `origin_id` catch the Commons-Flickr transfer. The perceptual hash and near-duplicate pass are refused: they only catch duplicates that are both already in the pool, the laundering case D5 names comes from an original outside it (caught by the "own work" rule and the upstream-most rule), and Geoff sees near-duplicates in one grid. |
| M6, D1, K15 (scratch) | Folded | `photos/.candidates/`, gitignored, gitleaks allowance landed by the conductor before round one; 25-file asset batches; roughly 600px thumbnails plus full files; prune after each round; shared database documents keyed by `source_id`. |
| D1 (hash rule) | Folded | Round close step 2: a keep applies only to a file matching `source_sha256`. |
| D2, K15 (lane registry) | Folded | Process: document lane in STATUS's registry; `git add` plus `git commit -- <paths>` (corrected, see Verification); `git show --stat` check; retry on `index.lock`. |
| D3 | Folded in part | Round close step 1: closed rounds, snapshot read, explicit reopen; database keyed by `source_id`. The separate reconciliation line is refused: applying the closed snapshot leaves no window for drift, and the round log records counts. |
| M7 | Folded | ajv dropped; the Node check reads `vocab.json` directly, so the strict-mode and meta-file failures M7 probed do not arise, and no stale `ajv-cli` dependency enters. |
| M8, K3 | Folded | Closed license list in `vocab.json`, SPDX IDs where they exist, `LicenseRef-` otherwise, each mapped to its legal code; `license_url` becomes derived. |
| M9, D11 | Folded | Encode command pinned (ImageMagick 7.1.2, cwebp 1.6.0, `-q 80 -m 6`); `sha256` written by the encoding step; size estimate corrected to about 48MB from M9's measured 164KB mean. |
| M10, D11 (GPS) | Folded | `-metadata icc` keeps tone and drops GPS; the `notice` field carries any embedded copyright notice into CREDITS, meeting the keep-notices duty without shipping location EXIF. |
| M11 | Folded | Artists: agents record names; one serialized MBID step at no more than one request per second. |
| M12 | Folded | Licensing: grant-time evidence for Flickr-origin files; a live-page change is logged, not a wrong quote. |
| M13 | Folded | Place: current ISO 3166-1 code plus an optional historical country. |
| M14 | Folded | Storage: rung 7 notes (package inside `photos/`, embed by pattern, leaf package). |
| D8 | Folded | `withdrawn` status, integrity refusal, pre-release source recheck. D8's history fork is taken as settled at option (a), as the review recommends: the binary is the distribution act, and Git LFS or R2 adds infrastructure for about 300 files. |
| D12 | Folded | Categories: a vocabulary edit rewrites affected records in the same commit, and the integrity check proves the set. |
| K1 | Folded | Rung 7 named as the consumer; errata owed below. |
| K5 | Folded | Categories: film-edge slot mapping; Changes bullet on `subject` becoming `title`. |
| K6 | Folded | Outreach: the British Culture Archive joins the first letters. |
| K9 | Errata owed | ROADMAP, below. |
| K10 | Folded | Changes list: ruling 7's core, hip-hop's era, the manifest shape, the new counted rules; "outsider artists" restored to the scene list; DC hardcore flyers first restored to Process. Hip-hop's "every era" is not raised as a fork: it is listed in Changes, so approving the spec approves it. |
| K11 | Folded | Contact sheet references replaced with the sorting page. |
| K14 | Folded | Outreach: never "non-commercial"; "free and open-source under the MIT license", no publication date. |
| K15 (future website) | Refused | The stable-slug rule stands on its own merits, and the website line records intent from the drafting conversation without building anything for it. |

## Errata owed (ratified documents, not edited by this fold)

- **Visual identity spec:** the one-line pointer to this spec; ruling 7's core (hip-hop joins);
  Curation rules (`:269-279`: target, hip-hop era); Licensing's "manifest entry" and field list
  (`:289-291`: one record per image, `subject` now `title`); Photography's "the shipped file stays
  unmodified" (`:248-249`: the file is resized and re-encoded, CREDITS says so).
- **ROADMAP `photo-curation` (`:178-185`):** target about 300, not 100 or more; "rung 7 consumes
  it", not "rung 7 and 5c-iii"; owner touchpoints are the sorting sittings and reading and
  sending letters, not only outreach approvals.
- **5c-iii plan (`:150`):** Ruling 3's "routed to the curation pass" becomes "routed to rung 7,
  which consumes the curation pass's records", matching the plan's own handoff at `:1774` and
  `:1815`.

## Queued repo edits for the conductor

- Before round one: `/photos/.candidates/` in `.gitignore`, and `'''photos/\.candidates/.*'''` in
  the `.gitleaks.toml` allowlist with a comment in the file's existing style.
- After lane 1 finishes, in its batch: a `scripts/check.sh` leg running
  `node tools/photos/check.mjs` in integrity mode.
