"""Author the vendored template FLAC fixture, once.

Run manually, from anywhere, as ``python3 gen_template_flac.py``: it
(re)writes ``template.flac`` next to itself and is never invoked again
on its own, by anything. This script is the fixture's provenance
record, not a build step; ``make derive`` and the repository gate never
run it (see corpus/README.md's file-tags fixture section for the
committed output's byte size and the exact command that produced it).

The fixture is a 42-byte, valid, silent FLAC container carrying no
tags at all: the four-byte ``fLaC`` signature followed by one
last-block STREAMINFO metadata block (mono, 44.1kHz, 16-bit, zero
samples). That is the minimum ``mediafile``'s reader needs to
dispatch a file as FLAC: it delegates to mutagen's own parse and
checks ``type(mgfile).__name__ == "FLAC"``
(``mediafile/__init__.py:174``), which only requires the signature and
a well-formed STREAMINFO block, nothing more.

No ``ffmpeg``, no ``flac`` binary, and no audio codec at all: mutagen
only needs the metadata container to be well-formed to open it, and
mutagen is already a pinned transitive dependency of mediafile
(``tools/venv/requirements.txt``, "via mediafile"). Every subject case's
own derivation copies these committed bytes into a fresh scratch path
and writes real tags into that copy; this script authors the shared,
tag-free starting point those copies begin from.
"""

from __future__ import annotations

import os
import struct

from mutagen.flac import StreamInfo

# STREAMINFO's metadata block type code, per the FLAC format; 0x80 ORs
# in the "this is the last metadata block" flag, which it is here: the
# template carries no other block and no audio frame.
_STREAMINFO_LAST_BLOCK_HEADER_BYTE = 0x80


def streaminfo_payload() -> bytes:
    """Build STREAMINFO's 34-byte payload for a silent, tag-free stream.

    ``StreamInfo.__new__`` bypasses ``StreamInfo.__init__``, which only
    knows how to parse an existing block's bytes; the fields below are
    set directly instead, mirroring how a fresh block is authored with
    no bytes yet to parse. ``total_samples=0`` is what makes the
    container silent: a decoder reads zero samples and never needs an
    audio frame to follow.
    """
    info = StreamInfo.__new__(StreamInfo)
    info.min_blocksize = 0
    info.max_blocksize = 0
    info.min_framesize = 0
    info.max_framesize = 0
    info.sample_rate = 44100
    info.channels = 1
    info.bits_per_sample = 16
    info.total_samples = 0
    info.md5_signature = 0
    return info.write()


def main() -> None:
    payload = streaminfo_payload()
    # A metadata block header is one byte (last-block flag | block type)
    # plus a three-byte big-endian length; STREAMINFO's type code is 0.
    header = bytes([_STREAMINFO_LAST_BLOCK_HEADER_BYTE])
    header += struct.pack(">I", len(payload))[1:]
    data = b"fLaC" + header + payload

    out = os.path.join(os.path.dirname(os.path.abspath(__file__)), "template.flac")
    with open(out, "wb") as f:
        f.write(data)
    print(f"wrote {out} ({len(data)} bytes)")


if __name__ == "__main__":
    main()
