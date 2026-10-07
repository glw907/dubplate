# Navidrome library creation request, observed 2026-10-07

This record serves Task 9 of the slice 2b plan. Navidrome's source was not opened. The evidence is the running app, its
network traffic, and its published docs (`multi-library.md`, `:46-57`, `:62`).

## Stack

`harness oracle-up` at the committed digests: Navidrome `docker.io/deluan/navidrome@sha256:38dc2727bfcfd5ede290f8ada114fc90368146f265ae4701ddddbcbe2a44ee52`,
reporting server version `0.64.2 (10114574)`. The container mounts only `/music`, which holds the fixture (29 tracks).
Both runs used a fresh `oracle-up` and the first admin, `admin`, whose password the command prints.

## Run 1: observation in the web UI

A Playwright script drove Chromium through the real UI and logged requests. The Playwright MCP server could not start
(no Chrome at `/opt/google/chrome/chrome`), so the repo's own Playwright 1.63.0 and its Chromium did the same work.
Steps: sign in, open `/app/#/library`, press CREATE, fill name and path, press Save. The path `/lib2` was made first with
`podman exec <container> mkdir -p /lib2`.

Login:

- Request: `POST /auth/login`, header `content-type` (plus browser defaults), body `{"username": string, "password": string}`.
- Response: 200, body `{"id": string, "isAdmin": bool, "name": string, "subsonicSalt": string, "subsonicToken": string, "token": string, "username": string}`.
  The `token` is a JWT.

Creation:

- Request: `POST /api/library`, headers `x-nd-authorization` (value `Bearer <token>`), `x-nd-client-unique-id`, `content-type`, `accept`.
- Body: `{"name": string, "path": string}`. The form also has a "Default for New Users" checkbox, left off; the UI then sent no `defaultNewUsers` field.
- Response: 200, body `{"id": string}` (the value was `"2"`).

The libraries list afterwards showed "Music Library /music" (29 songs) and "Second /lib2" (0 songs). The list is `GET /api/library?_start&_end&_sort&_order`;
an item carries `id` (number), `name`, `path`, `remotePath`, `lastScanAt`, `totalSongs`, `totalAlbums`, `totalFolders`, `defaultNewUsers`, and further totals.

## Run 2: replay with curl on a fresh oracle-up

```
B=http://127.0.0.1:<navidrome port>
curl -s -X POST $B/auth/login -H 'Content-Type: application/json' \
  -d '{"username":"admin","password":"<redacted>"}'
curl -s -w '\nstatus %{http_code}\n' -X POST $B/api/library \
  -H 'X-ND-Authorization: Bearer <redacted>' -H 'Content-Type: application/json' \
  -d '{"name":"Second","path":"/lib2"}'
curl -s "$B/rest/getMusicFolders?u=admin&t=<redacted>&s=<redacted>&v=1.16.1&c=probe&f=json"
curl -s "$B/rest/getScanStatus?u=admin&t=<redacted>&s=<redacted>&v=1.16.1&c=probe&f=json"
```

Login returned 200 with the body shape above. Creation returned:

```
{"id":"2"}
status 200
```

Eight seconds later (the new library's scan had finished):

```
getMusicFolders: {"subsonic-response":{"status":"ok","version":"1.16.1","type":"navidrome","serverVersion":"0.64.2 (10114574)","openSubsonic":true,"musicFolders":{"musicFolder":[{"id":1,"name":"Music Library"},{"id":2,"name":"Second"}]}}}
getScanStatus:   {"subsonic-response":{"status":"ok","version":"1.16.1","type":"navidrome","serverVersion":"0.64.2 (10114574)","openSubsonic":true,"scanStatus":{"scanning":false,"count":29,"folderCount":13,"lastScan":"2026-10-07T21:38:44.618999356Z","scanType":"quick","elapsedTime":618999356}}}
```

The admin saw two music folders without any grant step, as the docs say. `count` stayed 29.

## Limit

The second path holds no tracks, because the container mounts only `/music`. This record settles the request shape and the
folder listing. It does not settle how `count` sums across two libraries that both hold tracks; Task 9's live evidence
owns that. The observation is owed again at R24's next deliberate Navidrome bump.
