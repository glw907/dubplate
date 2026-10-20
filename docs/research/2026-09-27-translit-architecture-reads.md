# Transliteration pass: architecture reads and fold task 3c (2026-09-27)

Three `go-architecture-reader` grades (Opus, no plan in context) ran at HEAD `aeea780`, after
the simplifier round, with conductor rulings inline. Task 3c takes every FOLD item.
Everything marked ROUTE goes to the architecture-cleanup list.

## cmd/dubplate-server: workmanlike (unidecode.go alone would grade sound with nits)

It confirms the install verb refuses before any write, that ownership checks precede the
already-installed shortcut, the atomic write, and the load-before-OpenStore order.

FOLD (Task 3c):
- One "unidecode install:" wrap at the RunE entry point. The helpers stop prefixing by hand
  (17 sites). The `--from` read wraps once with "--from %s: %w", so the Build error and
  ErrWheelTooLarge name the path. Drop the duplicated path in the load error (`unidecode.go:74`).
- The HTTP client becomes a package-level var seam (`var unidecodeClient`), swapped under
  t.Cleanup like geteuid. `newRootCmd` goes back to no parameters, `newTestRootCmd` goes, and the
  timeout const moves to `unidecode.go` without the copied byte count.
- `readUnidecodeWheelFile` calls the new `unidecode.ReadWheel`. It opens first and uses
  `f.Stat()` for the IsDir check.
- Rename `installUnidecodeTransliterator` to `setUnidecodeTransliterator`, and drop the four
  disclaimers (`unidecode.go:56-57`, `:69-70`, `:87`; `doctor.go:133-135`).
- Use Mkdir instead of MkdirAll for the fresh leaf, with a one-line umask reason
  (`unidecode.go:198-200`). Cut geteuid's comment to one line.
- Test discipline:
  - `unidecode_test.go:498-526` claims "every refusal" but drives one. Make it table-driven over
    the refusals it names, or delete it.
  - `:618-621` has a doc/fixture mismatch.
  - `loadCachedUnidecodeWheel` returns the path.
  - `:599` uses the literal "unidecode".

ROUTE: the owner, leaf-shape, and uid-name checks exist in four copies with two refusal wordings
(`store/leaf.go`, `unidecode.go`, `member.go:196-223`, `backup/restore.go:594-602`). Give them
one shared internal package.

## internal/engine/matcher/unidecode: sound with nits (density 0.216 vs archive/zip 0.258)
No reachable correctness defect; spine right (wheel checksum before zip parse, table checksum before decode, LimitReader(max+1)).
FOLD (Task 3c):
- lex.go:196-205 \xHH >= 0x80 appends a raw byte (Python 3 str gives U+0080 -> C2 80): refuse or utf8.AppendRune. Latent, masked by the pin.
- table.go:42 `r < 0x80` admits negative runes; use uint32(r) < 0x80.
- extract.go:64-77 a near-miss name is opened and drained before refusal, so an Open/CRC failure masks the name refusal; check the name first.
- fetch.go:66 LoadCachedTable is a test fixture in the production API; move to a unidecodetest subpackage (LoadCachedTable(tb testing.TB) *Table, tb.Fatal) per net/http/httptest.
- ReadWheel(r io.Reader) ([]byte, error) as the one bounded-read-then-checksum; cmd/dubplate-server/unidecode.go:256-261 and cmd/harness/fetchunidecode.go:66-69 call it; MaxWheelSize, ErrWheelTooLarge, ErrTableChecksum unexport (ErrWheelChecksum stays only if a production caller branches).
ROUTE to cleanup: Table's four header fields read only by encode and a test (table.go:17-22, 84-93, 238-244); verify() helper for the checksum compare (fetch.go:43-47, extract.go:22-26, 37-41, table.go:30-34); "1.4.0" literal x4 (derive TableFileName and the wheel name); punctuationKind and expect's want param (lex.go:103-115, 243-270); position prefix x6 (lex.go); LoadCachedTable/loadCachedWheel shared body; drain vs read branch (extract.go:71-93); file split (build.go for Build and sweepDigest); dead branches and `hex` shadowing (extract.go:128-139); TestFetchRefusesNon200Status asserts only err != nil (fetch_test.go:71-83); helper label (extract_test.go:59); rebuttal comment x3; lex.go:21 says byte offset.
Best: fetch.go:35-41 LimitReader(max+1); extract.go:21-43 with table.go:25-35 both-ends pinning; extract.go:156-176 sweepDigest over Lookup.

## internal/engine/matcher: workmanlike (density 0.81 vs regexp 0.29)
The transliteration seam is the right shape.
FOLD (this pass's own lines): translit.go:8-13 cites time/tzdata as precedent, which is wrong (it never panics); cite crypto.RegisterHash/Hash.New or nothing; drop translit.go:17-19's reinstall sentence.
ROUTE to cleanup (pre-existing package): Distance.Update dead (penalty.go:197-208); ~15 exports with no outside caller (Transliterate, Levenshtein, DistanceForAlbum, AlbumPair, AssignItems, AlbumPairsFor, ParseRecommendation, ParseAction, SD* tables, BeetsSDPatterns, SDPatternsNoBracketDiscount, ErrInvalidTunable, Tunables.StringDistPatterns); LegalQuietFallbacks/LegalNoneRecActions exported mutable slices (verdict.go:86-98); sdForeignRegexps test-only mutex cache (strdist.go:161-193); component Name/Key test-only, three spellings (track.go:9-13, score.go:12-16); compileOptions duplication and compilePanicAsError recover (album_media.go:41-51, album_country.go:36-46, tunables.go:495-512); quiet_fallback/none_rec_action checks twice with drifted citations (tunables.go:481-490, verdict.go:184-205); recNames table; 21 one-function files -> album.go and track.go, Levenshtein to strdist.go, quiet.go to tunables.go; stale self line-citations (quiet.go:14-15, tunables.go:498); ~30 process citations; audience-arguing comments (penalty.go:348-351, 290-293; tunables.go:326-329; strdist.go:77-80; compat.go:7-16, 24-28); likely[T] generic (likelies.go:51-60).
Best: strdist.go:48-54; penalty.go:223-231; verdict.go:9-19 with :166.

FOLD (Task 3c) from the unidecode read also: ReadWheel as the one bounded read then checksum, used by Fetch, the --from path, and cmd/harness's wheelChecksumMatches; unexport MaxWheelSize, ErrWheelTooLarge, ErrTableChecksum. The simplifier review (`aeea780`, accept) agreed on the unidecodetest move; give its error the "unidecode:" prefix.
