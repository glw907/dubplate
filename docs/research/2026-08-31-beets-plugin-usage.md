# Beets plugin adoption survey (2026-08-31, for the dubplate-organizer initiative)

Sonnet dispatch: GitHub code search over public beets configs. 277
candidate files found, 90 sampled, 87 with a usable plugins: list (84
unique repos). YAML-aware regex parse handling inline, block-list, and
folded styles; commented-out plugins excluded. Raw tally beside this file
(2026-08-31-beets-plugin-tally.json).

## Frequency (n=87)

fetchart 89.7% | embedart 67.8% | lastgenre 55.2% | info 40.2% |
duplicates 35.6% | lyrics 35.6% | scrub 35.6% | chroma 34.5% |
replaygain 34.5% | edit 31.0% | musicbrainz 31.0% | missing 31.0% |
fromfilename 29.9% | mbsync 28.7% | convert 26.4% | inline 21.8% |
discogs 18.4% | ftintitle 18.4% | mpdupdate 16.1% | zero 14.9% |
badfiles 13.8% | smartplaylist 11.5% | play 10.3% | deezer 9.2% |
web 9.2% | the 9.2%. Long tail below ~7%: ~40 plugins at 1-3 configs
each (bandcamp, spotify, keyfinder, acousticbrainz, filetote, ...).

## Corroboration and caveat

Beets' own docs foreground art fetching as the common early addition and
call chroma "tricky to install, try without it first" - both match the
sample. No survey-style community threads exist to cross-check; the docs
plus this sample are the evidence. Sampling bias: dotfiles publishers
are power users (8-15 plugins per config here), so absolute counts
overstate the casual user; the RANKING (art > genre > hygiene > niche
sources) is the durable signal, and out-of-box coverage for casual
users is higher than these percentages suggest.

## Conclusion (batteries scope)

Ship in the box: art fetch+embed, genre, duplicates, scrub-style tag
hygiene, fingerprinting, ReplayGain, MusicBrainz sync - the 28-90% band,
all stable-dependency (MusicBrainz, Cover Art Archive, ffmpeg,
chromaprint, pure logic). Surface as options, not defaults: convert,
discogs, ftintitle, zero (13-26%). Refuse permanently: the sub-10% tail
- niche metadata sources, scrapers (lyrics sites), and personal
workflow glue - which is exactly the fragile-API territory the plugin
firewall exists for upstream.
