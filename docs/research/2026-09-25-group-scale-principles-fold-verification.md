# Group-scale principles spec: fold verification

Target: `docs/specs/2026-09-25-group-scale-principles-design.md`, revision 2 at
`08c812f`. Fold record: `2026-09-25-group-scale-principles-fold.md`. Reviews: the contract,
mechanics, misuse, consistency, and early prose files under this directory. The reader took no
part in the fold. Line numbers below are the spec's at `08c812f` unless another file is named.

Findings are ranked by consequence, not polish, per Anthropic's guidance that "a reviewer
prompted to find gaps will usually report some, even when the work is sound". The fold is
sound in most places. One fact it chose from a self-contradicting review is wrong, and that
fact touches production today.

**Counts: 1 blocker, 3 major, 14 minor.**

## Sources checked for this read

- Navidrome v0.63.2 source, `conf/configuration.go` (the tag `compose.yaml:27` pins):
  `viper.SetDefault("enabledownloads", true)` (line 758) and
  `viper.SetDefault("enablesharing", true)` (line 797). `viper.SetEnvPrefix("ND")` plus
  `viper.AutomaticEnv()` (lines 925-928).
- Navidrome docs, `navidrome/website` `content/en/docs/usage/configuration/options.md`:
  - "| EnableDownloads | `ND_ENABLEDOWNLOADS` | Enable the option in the UI to download
    music/albums/artists/playlists from the server | `true` |"
  - "| EnableSharing | `ND_ENABLESHARING` | Enable the Sharing feature | `true` |"
  - "When the same option is set using multiple methods, Navidrome uses the following order of
    precedence (highest to lowest): environment variables, command line arguments,
    configuration file."
- Navidrome v0.63.2 `server/subsonic/api.go:211-217`: the four share endpoints are live when
  sharing is on, and `h501` otherwise.
- OpenSubsonic `content/en/docs/Responses/error.md`: "50 | User is not authorized for the given
  operation." and "70 | The requested data was not found."
- MusicBrainz `relationships/artist-url`: artist types include `bandcamp`, `purchase for
  download`, `purchase for mail-order`, `official homepage`, and `free streaming`. No type
  means "the artist's own store".
- Repo: `config/navidrome.toml` sets neither key. `compose.yaml:29-47` sets `ND_*` environment
  variables for the same container. The tunnel ingress
  (`config/cloudflared-tunnel-config.json`) sends every path on `music.907.life` to
  `navidrome:4533`. `cmd/dubplate-server/serve.go:289-311` and `web.go:163-170` register routes
  by subtree name through `shell.Register`. `internal/serve/doc.go:46` names a fallthrough slot
  (`Shell.SetFallthrough`). `internal/serve/clientaddr.go:12-17`: the zero proxy config returns
  the TCP peer as the client address. No `restore` verb is registered yet (`main.go:14` lists it
  as rung 6), so "specified" is right for that row.
- Streamer spec header (revisions 2.2 to 2.4, dates as the spec states), `:297-312`
  (play-history shape; "A stream start is a separate event kind"), `:277` (the hashed
  modern-only app password), `:380` (the one unauthenticated carve-out). Configuration homes
  header (ratified 2026-09-23), "Profiles in the TUI" at `:150`, member list with last-seen
  address at `:171`.

## Question 1: did each blocker and major close at the cited location?

Yes for every blocker and major except the one noted, with the location checked in revision 2.

| Finding | Closed at | Verdict |
|---|---|---|
| K-B1, C-M5, U-M1, U-M2 (download limit, false copy) | :185-213, :59-61, :325-329 | Closed |
| K-M1, N-B1 (checkout) | :233-237, :128-133, conflict 7, ruling 2 | Closed |
| K-M2, N-M2, C-M8, P-B1 (`added_by`, source channels) | :169-183, :317-321, conflict 5 | Closed |
| K-M3, C-M7 (event shape) | :201-213, conflict 4, A6, A8 | Closed; window rule underdefined (m1) |
| K-M4 (buy-link coverage) | :257-276 | Closed; introduces M2 below |
| C-M1 (acceptance) | :398-418 | Closed; three mechanisms lack a row (m6) |
| C-M2, P-B2, P-B3 (status values) | :355-396 | Closed |
| C-M3 (route inventory) | :116-123, A2 | Closed; fallthrough unclassified (m4) |
| C-M4 (download id kinds) | :187-192, A5 | Closed; code choice (m2) |
| C-M6, U-M5 (member cap) | :88-111, A1 | Closed |
| C-M9 (refusals durable) | conflicts 9 and 10 | Closed as owed errata after approval |
| C-M10, N-M1, P-B3 (Navidrome period) | :239-253, A13 | **Not closed: built on a false default (B1, M1)** |
| U-M3 (shared credential) | :149-153 | Closed in part; refusal reason weak (m7) |
| U-M4 (volume view privacy) | ruling 5 | Posed; its recommended copy contradicts the body (M3) |
| N-M3 (buy-link premise) | ruling 7 | Closed |
| P-B4 (future tense on the website copy) | :50 | Closed |

## Findings

### B1 (blocker). Navidrome's sharing is on by default in the pinned 0.63.2, so the "Before 0.8" premise and the "no public links" line are false today

- **Location:** spec :245-249 ("Navidrome's public share links are off by default, and nothing
  keeps them off today. The Navidrome configuration sets `EnableSharing = false` explicitly");
  :56-57 (the "isn't" line "It has no public links"); fold record :233-234 (erratum 9, "owed by
  the next code lane").
- **Defect:** Navidrome 0.63.2 sets `enablesharing` to `true` (source line 797, docs row
  above). The repo's `config/navidrome.toml` and `compose.yaml` set neither the key nor
  `ND_ENABLESHARING`, and the tunnel routes every path, `/share/*` included, to Navidrome. So
  the production box has the share feature on now, and a member can mint a public link that
  serves audio. The mechanics review carried both facts ("`conf/configuration.go:994,1035`
  default both settings to true today" and "Sharing shipped disabled by default in 0.49.0"),
  and the fold took the stale half. The present-tense "sets `EnableSharing = false`
  explicitly" also describes a change that has not landed.
- **Fix:** rewrite :245-247 as "Navidrome 0.63.2 turns sharing on by default (its
  configuration docs: `EnableSharing`, default `true`). The Navidrome configuration pins
  `EnableSharing = false`, with a gate leg (A13) ...". Move erratum 9 out of "owed by the next
  code lane" and route it to the conductor now as a repo change plus a deploy, under the
  box's quiet-box rule, with a read-only check for shares that already exist. Until it lands,
  the "isn't" line is false for the running server; say so in STATUS, not on the website.

### M1 (major). The A13 gate leg reads only the config file, and Navidrome's environment variables outrank it

- **Location:** spec :246-247 and A13 (:417); ruling 3's "Off" option (:493-494) has the same
  shape for `EnableDownloads`.
- **Defect:** Navidrome's documented precedence is "environment variables, command line
  arguments, configuration file". `compose.yaml` already feeds this container `ND_*` variables
  (`ND_SCANNER_SCHEDULE`, `ND_PASSWORDENCRYPTIONKEY`, and others). A leg that checks
  `navidrome.toml` for `EnableSharing = false` passes while `ND_ENABLESHARING: "true"` in
  `compose.yaml` or `env/dubplate.env.template` turns sharing back on. That is a vacuous pass.
- **Fix:** A13's fixture becomes "`config/navidrome.toml`, `compose.yaml`'s Navidrome
  environment, and `env/dubplate.env.template`"; passes when the file sets `EnableSharing =
  false` and neither other file names `ND_ENABLESHARING`; fails when the key is absent, true,
  or overridden by the environment. Quote the precedence line in the "Before 0.8" section.
  Apply the same wording to ruling 3's "Off" option. Rung 12's `doctor` row should read the
  effective value (`getUser`'s `shareRole`, which Navidrome sets from `EnableSharing`), not
  the file.

### M2 (major). The buy-link source order and display order contradict each other, and one allowed type does not exist in MusicBrainz

- **Location:** spec :263-274, A14 (:418).
- **Defect:** "Where the links come from, in order" reads artist links only as the last
  fallback, "labeled as the artist's rather than this album's". "What order they appear in"
  then puts "the artist's own store and Bandcamp" first. If artist links render only when the
  release and its siblings have none, they never sit beside release links, so that ordering is
  either moot or means something else. The likely intent is a release-level link whose host is
  Bandcamp or the artist's shop. MusicBrainz has no "own store" type: the artist-url types are
  `bandcamp`, `purchase for download`, and `purchase for mail-order`, and a release's
  Bandcamp page is filed as `purchase for download`. So "artist-direct first" needs a host
  rule the spec does not state, and A14's "in the stated order" has two readings. This is the
  planning-miss class the repo rule names: a state the check can meet with no named report.
- **Fix:** state two rules. (1) Sources: take the first of release, sibling releases, artist
  that yields at least one allowed link; artist links carry the "the artist's" label. (2)
  Order within the rendered set: a link whose host is `bandcamp.com` (or a subdomain) or
  equals the host of the artist's `official homepage` sorts first, then `purchase for
  mail-order`, then other `purchase for download`. Allowed types: release `purchase for
  download` and `purchase for mail-order`; artist `bandcamp`, `purchase for download`, and
  `purchase for mail-order`. A14's fixture adds a release-level Bandcamp link to prove the
  host rule.

### M3 (major). Ruling 5's member-page sentence contradicts the body's play-history disclosure

- **Location:** ruling 5 (:504-505), "The admin can see how much data your devices download,
  not what you play."; body :223-225, "Play history ... does record songs per member. An admin
  with shell or backup access can read it. Each member's own page says so in plain words."
- **Defect:** the recommended copy tells a member the admin cannot see what they play. The
  body says the admin can, through shell or backup access, and requires the page to say so.
  The mechanics review's m9 asked for exactly this disclosure. Geoff would ratify a privacy
  statement the spec itself calls false.
- **Fix:** the recommended sentence becomes: "In the app, the admin sees how much data your
  devices download, not what you play. The server keeps a history of what you play for your
  recently played and most played lists, and whoever runs the server can read it."

### Minor

- **m1, :201-211 and A6 (:410). The meter's window rules are undefined for a rolling window.**
  "Counted once per credential, track, and window" has no meaning when the window has no
  edges, and "a Range continuation adds bytes only" does not say which row holds those bytes,
  given that conflict 4 puts bytes on serve events. A6's "counts and bytes are exact" then has
  two readings. Fix: "Distinct tracks are the distinct (credential, track) pairs with any
  served byte in the trailing 24 hours. A request with a nonzero Range offset adds bytes to
  the credential's byte counter and records no new start."
- **m2, :187-190 and A5. Error 70 misstates the refusal.** OpenSubsonic's 70 is "The requested
  data was not found", which is untrue for an artist or playlist id that exists. Code 50,
  "User is not authorized for the given operation", is what Navidrome 0.63.2 answers for
  disabled downloads, and the mechanics review recommended it. Fix: answer error 50 with a
  message naming the song-or-album rule. Neither choice was quoted in the spec, so quote the
  chosen line.
- **m3, :149-153. Distinct addresses are vacuous in the default proxy mode.** The zero
  `TrustedProxyConfig` returns the TCP peer (`clientaddr.go:12-17`), which behind the tunnel is
  cloudflared, so every credential shows one address. Fix: name `serve.ClientIP` as the source
  and say the count is shown only in trusted-proxy mode, with a `doctor` row otherwise.
- **m4, :116-123 and A2. The route inventory has no class for the fallthrough.** From rung 12
  the shell's fallthrough (`Shell.SetFallthrough`) forwards `/rest/*` to Navidrome, which serves
  audio under "the streamer's own" authentication (build spec :560-563). It is none of public,
  session, or admin, and "None of the public rows serves audio" silently excludes it. Fix: add
  a fourth class, "delegated (the streamer's own auth)", and a row for the fallthrough.
- **m5, ruling 3 (:488-494). The framing settles this fork.** "Off" adds friction for honest
  members (clients that sync through `download` lose offline sync) and none for an abuser,
  who uses `stream`. That is the test the spec used to drop the album limit (:14-15). All
  three lenses recommended "on". Fix: take "on" as a spec decision and report it, per the
  standing no-ceremony rule.
- **m6, :398-418. Three new mechanisms have no acceptance row.** The admin credential list's
  distinct-address count, `musicBrainzId` in album responses, and rung 12's sharing `doctor`
  row. Fix: add a row for each, or state that A7, A3, and A13 cover them.
- **m7, fold record :244-246. U-M3's refusal rests on a premise the meter removes.** The fold
  refuses a concurrent-streams measure because "the direct stream path is stateless". The meter
  (:207-209) already wraps every `stream` and `download` response to count bytes written, which
  is where an in-flight counter would live. The refusal can stand on cost against signal, but
  the stated reason is stale. Fix: restate the reason in the fold record.
- **m8, ruling 2 (:483-487). Not an independent fork.** Its recommendation is "follow ruling
  1", and each option is bound to one answer of ruling 1. Fix: fold it into ruling 1 as a
  stated consequence, leaving six forks.
- **m9, ruling 4 (:495-501). The notice threshold lacks evidence for a small library.** "Half
  the library in a day is well above any listening day" fails for a new server with a few
  hundred tracks, where one long day or one sync crosses it. Fix: a floor, such as "half the
  library or 1,000 tracks, whichever is larger", or state the founding library's track count
  as the evidence.
- **m10, :61. The "isn't" copy assumes ruling 5's answer.** "The admin can see how much each
  member's devices take" is false under ruling 5's "exceptions only" option. Fix: "the admin
  can see how much music is downloaded", or mark the line as following ruling 5. "Take" is
  also vague; use "download".
- **m11, :50. "Each album" overclaims.** About nine in ten releases carry no link of their own,
  and some artists carry only streaming links, which the purchase filter drops. Fix: "Album
  pages will link to places to buy the music, where MusicBrainz knows one."
- **m12, :264-265. The artist fallback's coverage is overstated.** "Every credited artist
  carried one" counts "free streaming" links (the mechanics sample: "free streaming" or
  "purchase for download"), which the spec then refuses to render. Fix: "every credited artist
  carried a buy or stream link; how many carry a purchase link is unmeasured."
- **m13, :327. "Symfonium's documented response is to retry or to switch endpoints".** The
  switch is a user toggle the docs tell the user to check, not the client's response. Fix:
  "Symfonium retries, and its docs tell the user to switch off the download endpoint."
- **m14, :49 and :59. Website wording.** "Where it came from" reads as verified provenance;
  it is the member's own answer and can be unrecorded. "From 0.8" names a version a website
  reader can't place. Fix: "who added each album and what they said about where it came
  from"; publish the dated lines only once 0.8 ships, or drop the qualifier then. Separately,
  :197 runs to 122 characters; rewrap it.

## Question 2: contradictions the fold introduced

- Is/isn't against the body: "no public links" (:56) against the real Navidrome default (B1);
  :61 against ruling 5's open options (m10). The rest holds: the download scope, federation,
  fetching, and legal lines each match their section.
- Acceptance against mechanisms: A13 against Navidrome's precedence (M1); A14's order against
  the source order (M2); A6's exactness against the undefined window (m1). A1 to A5, A7 to A12
  match their mechanisms.
- Rulings against the body: ruling 5's copy against :223-225 (M3). Ruling 7's premise (album
  pages at rung 7) matches the visual identity spec as the fold verified. Rulings 1, 4, and 6
  match the body.

## Question 3: new mechanisms stated from memory

| Mechanism | Sourced or marked? | Note |
|---|---|---|
| Store-backed rolling meter | Marked "new", landing 0.8 slice 5, A6 | A design decision; window rule undefined (m1) |
| Error 70 for other `download` id kinds | Marked "new", A5; code meaning not quoted | Code choice questionable (m2) |
| Credential cap's distinct-address count | Marked "new", 0.8 slice 6 | Address source exists but is vacuous by default (m3); no row (m6) |
| `EnableSharing` pin and gate leg | Key name right; default stated from the wrong source | B1, M1 |
| Buy-link fallback order | Sample quoted from the mechanics review | Order contradicts sources; type names partly invented (M2) |
| Route-inventory test | Marked "new"; feasible, since routes register by subtree name | Fallthrough class missing (m4) |

Navidrome key names and defaults, verified against the docs row and the v0.63.2 source:
`EnableSharing` / `ND_ENABLESHARING`, default `true`; `EnableDownloads` / `ND_ENABLEDOWNLOADS`,
default `true`. The spec's key names are right. Its sharing default is wrong. Its downloads
default ("which is its default", :251) is right.

## Question 4: can Geoff rule the seven forks?

- Ruling 1 (whole library on a member's devices): a genuine product fork, evidence adequate.
- Ruling 2 (checkout scope): derivative of ruling 1 (m8).
- Ruling 3 (Navidrome downloads): settled by the framing (m5). Its "Off" option also needs
  M1's environment check.
- Ruling 4 (defaults): a genuine taste call; the notice default needs m9's floor or the
  library size.
- Ruling 5 (volume view): genuine, but fix M3's copy first so Geoff does not ratify a false
  sentence.
- Ruling 6 (source required): genuine, evidence adequate.
- Ruling 7 (buy links timing): genuine priority call; rulable as posed.

No settled decision in the body is actually a fork. The spec's own decisions (dropping the
limit, narrowing `download`, dropping shares, the wishlist's web-only scope, link order) each
rest on the owner's framing or a prior ruling.

## Question 5: is the is/isn't section true and website-ready?

Not yet. It is plain and stranger-readable, and it no longer overclaims about downloads. Three
lines need a change before a website uses it: :56 is false for the running server until the
pin deploys (B1), :61 depends on ruling 5 (m10), and :50 and :49 overclaim mildly (m11, m14).
With those fixes it can go on the website as it stands.
