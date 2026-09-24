# Rung 5, pass 5c-ii plan: prose review (revision 2.1)

Reviewed: `docs/superpowers/plans/2026-09-24-rung-5c-ii-admin-gate-roster-api-web.md` at `0bb3722`, 2446 lines. Reviewer: `prose-voice-reviewer` (Opus 5.5). Register: Google developer documentation, Go dialect.

## Deterministic floor

`tellgrader --register docs`: zero findings (0 tells per 1000 words); 19,276 words, 992 sentences, 11 tricolons, cadence CV 1.15 (tables and state lists inflate it). The docs-register measures profile did not resolve for this repo.

## Facts checked and found sound

Ceiling arithmetic (lines 356-387: 6.75M projection, 7.52M flag, 0.77M headroom, 7.48M and 8.21M split projections); 5c-iii and 5c figures (4.95M, 17.55M); deliverable count 31; model mix; HISTORY basis (`HISTORY.md:378-380`); 5c-i plan Ruling 6 and outline figures; every rung 5 spec pointer; the Go pointers in `auth.go`, `registry.go`, `config.go`, `roster.go`, `identity.go`, `handlers.go`, `prefix.go`, `serve.go`, `Makefile:52`, `check.sh:400`; RFC 6797 quotes (7.2, 8.1, 8.1.1, 8.3); `go doc` quotes; kit 2.70.3 source quotes; npm versions.

## Blockers

**P1 (lines 92-100, 106-109).** Ruling 1's "Reason" says thirteen tasks at 0.725M "puts it near 9.4M in one pass" (13 × 0.725, bare rate, no adders), while the question asks 9.4M for 5c-ii alone: one figure, two meanings. The "no" branch's "ceiling near 12.2M, flag near 9.8M" omits the adders "Token ceiling" applies (0.95M for 5c-ii's share, 0.60M for 5c-iii's); with them a one-pass projection reaches about 11M, above its own 9.8M flag, so the flag fires on plan, contradicting line 381's rule. Fold: restate the Reason around 5c-ii's 6.75M projection and re-derive the one-pass figure (about 11M projected, ceiling near 13.8M, flag near 11M).

**P2 (line 82, lines 136-139, 203-208).** Rulings 2 and 5 each give two outcomes for "no", so a "no" does not decide what is built. Ruling 2's second outcome is a fragment. Fold: one outcome per "no" (Ruling 2: no HSTS header, finding 18 open; Ruling 5: `.js` with JSDoc, `ts-conventions` claim dropped), naming the alternative as something Geoff can name in his answer.

## Warnings

**P3 (lines 141-157).** Ruling 3 asks two questions in one sentence, its "no" answers only the palette half, and "Both data files" refers to a match file the ruling never named. Fold: one question; name both theme files Task 4 writes.

**P4 (lines 165, 502).** `thresholds.go:124-128` should be `:123-126`. The `resolveIdentitySource` pointer reads as `cmd/dubplate-server/config.go`; it is `internal/config/config.go:291-296`.

**P5 (lines 82-87).** The rulings section opens with a 50-word sentence of fold history. Fold: lead with "Seven questions, each with a recommendation", the standing rulings not reopened, and a pointer to the fold record.

**P6 (lines 123-134).** "Fixed, not asked" is the longest rulings block and asks nothing; setup-colon payoff. Fold: two sentences and a pointer to R-5cii6.

**P7 (lines 235-236).** Three reasons chained by commas. Split.

**P8 (lines 175-176).** The `none`/`.css` sentence reads as though the sentinel reaches the path check. Fold: "`none` means no branding declarations and no operator file. It cannot shadow a real file, because R-5cii13 admits only paths ending in `.css`."

## Suggestions

**P9 (219-221)** split the WebKit hedge into its own sentence. **P10 (102-106)** break Ruling 1's 55-word "What yes builds" into sentences. **P11 (187-189 vs 1094-1098)** match Ruling 5's include-set list to R-5cii16, or say "for example". **P12 (392-394)** source or decompose "about 2.8M of planning". **P13 (fold record:183)** the record's "0.85M" headroom should be 0.77M. **P14 (1153)** "The gate hides the members document; the route itself is public." Leave 1226 and 1952. **P15** reflow a few over-wide source lines (319-320, 1115-1117, 1241, 2172).

## Verdict

Outside "Rulings for Geoff" the plan reads as written by a careful engineer from verified sources. Fold P1 and P2 before Geoff reads it.
