---
name: systemd-python-subprocess
description: What a hardened dubplate unit needs before it can run beets (CPython) as a subprocess
metadata:
  type: reference
---

`dubplate-serve.service` carries the full U12a hardening set. Running beets (a
CPython program installed by uv under `/srv/music/.local`) as a subprocess of it
needed exactly one directive change: `AF_NETLINK` added to
`RestrictAddressFamilies`, because glibc's `getaddrinfo` opens a `NETLINK_ROUTE`
socket to enumerate local addresses before resolving a hostname. It grants no
egress; it only reads this host's routing table.

Everything else already held: `ProtectHome=yes` does not cover `/srv/music` (the
music user's HOME), `ProtectSystem=strict` still permits reads and exec anywhere,
and a scratch directory under `StateDir` (`/var/lib/musicbox`) is already in
`ReadWritePaths`.

Unverified on the box as of 2026-09-01: `MemoryDenyWriteExecute=yes` and
`SystemCallFilter=@system-service` against CPython plus its compiled extensions
(numpy, jellyfish, lap). Judged low risk, since none of them JIT, but the first
real box run is the proof.
