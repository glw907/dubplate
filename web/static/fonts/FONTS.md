# Embedded fonts

Three faces ship here, within a four-woff2-file budget (a fourth face
lands later). Every face ships under an SIL Open Font License 1.1
that permits redistribution and embedding; the verbatim license texts
sit in `licenses/`. The font-delivery gate reads this file: `- File:`
names the file and `- SHA-256:` names its hash, in that order, once
per face.

## Alfa Slab One

- File: `AlfaSlabOne-Regular.woff2`
- Family: Alfa Slab One
- Font version: Version 2.000 (the `name` table's ID 5)
- Upstream: `google/fonts` `ofl/alfaslabone/AlfaSlabOne-Regular.ttf` at commit
  `ae38321d5a5ff3a2125357402dd332998e6bf41b`
- License: SIL Open Font License 1.1, `licenses/AlfaSlabOne-OFL.txt`
  (`openfontlicense.org`); the source carries the copyright line
  "Copyright 2016 The Alfa Slab One Project Authors ... with Reserved
  Font Name "Alfa Slab"."
- Conversion: `uvx --from 'fonttools[woff]==4.66.0' fonttools ttLib.woff2
  compress --no-glyf-transform -o AlfaSlabOne-Regular.woff2
  AlfaSlabOne-Regular.ttf` (fontTools 4.66.0, MIT)
- Comparison (`fonttools ttLib.woff2 decompress` against the source
  TTF): every table but `DSIG` and `head` byte-identical; `DSIG`
  absent from the output and an 8-byte stub in the source; `head`
  differs only at bytes 8 through 11 (`checkSumAdjustment`) and at
  byte 16 (flags bit 11 set, `0x3` to `0x803`)
- Reserved Font Name: the source carries the Reserved Font Name "Alfa
  Slab". The OFL FAQ (item 2.2.1) counts a change "unchanged except
  for WOFF compression" when "the original font data remains
  unchanged except for WOFF compression, and WOFF-specific metadata
  is either omitted altogether or present and includes, unaltered,
  the contents of all equivalent metadata in the original font." Both
  edits above are the WOFF2 encoder's own, not the converter's, and
  the source `DSIG` stub carries no signature to lose, so the face
  ships under its own name, "Alfa Slab One".
- Size: 39,668 bytes
- SHA-256: `3a5cd3048a1a9c06d4737ef2a05cc612690af508f5da7b8db149f182afbd13d9`

## Work Sans

- File: `WorkSans-Variable.woff2`
- Family: Work Sans
- Font version: Version 2.012 (the `name` table's ID 5)
- Upstream: `google/fonts` `ofl/worksans/WorkSans[wght].ttf` at commit
  `8b0a1d0f5983c89bc2b93f1b5fb55f9e252744b5` (upstream
  `weiweihuanghuang/Work-Sans` at `b35c81086186162164947bd39574683073d9b268` per `upstream_info.md`)
- License: SIL Open Font License 1.1, `licenses/WorkSans-OFL.txt`
  (`openfontlicense.org`); the source carries the copyright line
  "Copyright 2019 The Work Sans Project Authors"; no Reserved Font
  Name.
- Conversion: `uvx --from 'fonttools[woff]==4.66.0' fonttools ttLib.woff2
  compress --no-glyf-transform -o WorkSans-Variable.woff2
  'WorkSans[wght].ttf'` (fontTools 4.66.0, MIT)
- Comparison (`fonttools ttLib.woff2 decompress` against the source
  TTF): every table but `DSIG` and `head` byte-identical; `DSIG`
  absent from the output and an 8-byte stub in the source; `head`
  differs only at bytes 8 through 11 (`checkSumAdjustment`) and at
  byte 16 (flags bit 11 set, `0x3` to `0x803`)
- No Reserved Font Name in the source, so no rename question applies.
- Size: 138,000 bytes
- SHA-256: `4c5e33dfeb2e14a01b3ec6cf47661f80df78939dc2b12f572186d9329aa25240`

## IBM Plex Mono

- File: `IBMPlexMono-Regular.woff2`
- Family: IBM Plex Mono
- Font version: Version 2.005 (the `name` table's ID 5)
- Upstream: `@ibm/plex-mono@2.5.0` npm tarball,
  `package/fonts/complete/woff2/IBMPlexMono-Regular.woff2` (49,248
  bytes), taken as is: IBM ships Plex Mono as WOFF2 already, so no
  conversion runs on it.
- License: SIL Open Font License 1.1, `licenses/IBMPlexMono-OFL.txt`
  (`openfontlicense.org`); the source carries the copyright line
  "Copyright © 2017 IBM Corp. with Reserved Font Name "Plex""; the
  file ships under the upstream name, unmodified.
- Conversion: none; the npm tarball's own WOFF2 file is copied
  unchanged.
- Comparison: not applicable (no conversion).
- No Reserved Font Name question applies: it binds Alfa Slab One alone.
- Size: 49,248 bytes
- SHA-256: `ba204497f16b6d334cee9d1e963a831b73e3a56e1d6300a8489d18df7214b350`
