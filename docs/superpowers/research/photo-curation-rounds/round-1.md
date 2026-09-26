# Photo curation round 1

Plan: `docs/superpowers/plans/2026-09-25-photo-curation-round-one.md`. Spec: revision 2.3.

## Task 4 checkpoint (search closed, 2026-09-25)

Seven search agents ran in parallel. `check.mjs sheet` over the candidate home wrote 114
candidates with no cross-agent duplicate.

| Agent | Kept | Main drop reasons |
|---|---|---|
| research-a | 38 | one too-small |
| research-b | 27 | 11 license (PDM marks, PD-self), 5 too-small, 11 redundant |
| research-c | 18 | ~22 too-small, ~13 license (implausible own-work claims) |
| hiphop | 21 | 6 too-small, 4 license, 3 PD-self (no basis yet) |
| flyers | 7 | 17 license, 16 too-small |
| research-d1 | 2 | 28 too-small, 40 off-topic |
| research-d2 | 1 | 28 too-small, 41 off-topic |

By family: reggae 22, punk 25, hip-hop 22, everything else 45. By primary subject: performers
53, clubs and venues 20, crowds and dancers 14, streets and neighborhoods 12, sound systems 7,
recording studios 6, record stores 1, pressing plants and cutting lathes 1. By region: North
America 39, Europe 24, no place 23, UK and Ireland 12, Caribbean 8, Latin America 5, one each
MENA, Asia and Pacific, Africa.

Findings:

- The unverified Flickr harvest yielded 3 of 140 rows; later rounds skip keyword harvests.
- The DC Punk Archive's Dig DC flyers are all marked In Copyright; the outreach letter is the
  only route.
- Beth Lesser's archive is 479px web copies; only her letter can reach it.
- Round-two leads: hiphop's 8 vetted, undownloaded candidates (its `manifest.json`);
  research-a's unpursued Skatalites and Notting Hill 1989 frames; the HABS 9:30 Club interior
  (year unknown); reggae record stores and sound systems; reggae's share.

Conductor corrections owed before the sitting:

- Drop the Sun Ra poster (flyers): its embedded 1973 photo has no verified license.
- Add a `self-dedication` public-domain basis (Commons `{{PD-self}}`) and recover the dropped
  PD-self images (1520 Sedgwick Avenue, two Bronx 2009 Kool Herc and Coke La Rock frames, two
  CBGB exteriors).
- Add the `off-topic` reason code; research-d1 and d2 logged off-topic drops as `passed`.
- Add the `sonidero` and `rebetiko` scenes to the everything-else family.
- Tighten `SOURCE_ID_RE`: `commons:<digits>`, `flickr:<digits>`, `archive:<name>:<key>` with
  dots allowed in the name.
- The spot-check tests every public-domain and flyer record, with particular attention to the
  1978 to 1989 no-notice-unregistered basis (the Black Flag flyer is likely Raymond
  Pettibon's), the Gottlieb jazz photos' basis (a donor dedication, not government work), and
  the 1976 Marley press photo's no-notice claim.

Operational lessons for later rounds (to the plan's Global constraints): wrap every download in
`timeout -k 10 150` as well as curl's own `--max-time`, which the sandbox did not always
enforce; kill only PIDs an agent recorded itself, never by pattern (one agent's `pkill -9 -f`
on the shared lock killed every agent's queued downloads); never have a search agent start
downloads it cannot finish within the round.

Spend: roughly 4.8M of the 5.5M ceiling at this checkpoint, the 4.4M flag crossed during the
search.

## Resume after the power loss (2026-09-25)

The conductor session ended mid-question when the laptop lost power. A fresh session resumed from
this log. The sorting page is published at `https://claude.ai/artifact/7CewFiyc6pE1SG1a4pfQPE`
(Task 2's close owed this line). Every Task 6 dispatch names it.

Owner rulings (Geoff, 2026-09-25):

- Ceiling raised to 7.0M; the 80% flag moves to 5.6M.
- Text-only flyers count as public domain under the threshold of originality (Commons
  `{{PD-text}}`), limited to purely typographic flyers with no drawing, logo art, or photo. The
  spot-check views each such flyer to confirm. This recovers the three 1984 to 1988 West Coast
  hip-hop flyers.
- The sorting page shows each photo in the context of the login page: the silver-print
  treatment and film-edge caption in the login page's photo column, beside the form, at desktop
  and phone widths. The crop to that column is the judgment Geoff makes while sorting.
- The sorting page gains a one-at-a-time review mode with yes, no, and maybe (Geoff: "cycle
  through them with a yes/no/maybe for each"). Conductor ruling: a maybe is treated like an
  unsorted candidate at close, staying in `.candidates/` for the next round, never rejected.
