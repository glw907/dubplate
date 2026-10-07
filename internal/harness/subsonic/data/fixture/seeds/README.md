# Fixture seeds

Every file here is generated, never downloaded. `../manifest.json`
pins each by sha256; the fixture builder refuses a seed whose hash
differs. The audio seeds carry the `flac` encoder's own
`VORBIS_COMMENT` block, which `tagio.Write` keeps, so a built file
never carries the building binary's version.

Tools: ffmpeg 9.0.2, flac 1.5.0.

Audio (0.4 s, 16 kHz, mono, 16-bit sine; `$F` is the frequency):

    ffmpeg -f lavfi -i "sine=frequency=$F:duration=0.4:sample_rate=16000" -ac 1 -sample_fmt s16 tone-$F.wav
    flac -s -f --best -o tone-$F.flac tone-$F.wav

`tone-220.flac`, `tone-330.flac`, and `tone-440.flac` use 220, 330,
and 440 Hz.

Images:

    ffmpeg -f lavfi -i "testsrc=size=96x96:rate=1:duration=1" -frames:v 1 -q:v 5 cover-a.jpg
    ffmpeg -f lavfi -i "smptebars=size=96x96:rate=1:duration=1" -frames:v 1 -q:v 5 cover-b.jpg
    ffmpeg -f lavfi -i "testsrc2=size=64x64:rate=1:duration=1" -frames:v 1 embedded-a.png
