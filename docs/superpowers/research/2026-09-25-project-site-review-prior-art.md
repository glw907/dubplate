# Project site spec review: comparable project sites

Reviewer lens: prior art from comparable self-hosted project sites. Target:
`docs/superpowers/specs/2026-09-25-project-site-design.md` at `e91f528`. All sources were checked
on 2026-09-25. The Paperless-ngx docs site returned a Cloudflare challenge, so its claims cite the
docs' Markdown source on GitHub, which is what the site builds from.

## Verdict

The page map is sound in outline. Three of its choices are already better than most of the
comparable sites:

- The operator and member readers are split. Funkwhale does the same, and most of the others do
  not.
- "Why an album was turned away" is a generated, member-facing troubleshooting page. None of the
  sites surveyed has an equivalent.
- The comparison page is cited and dated.

The map lacks four things that nearly every comparable site carries and that a stranger running a
server for a group would miss:

- Where to get help, and how to report a security problem.
- Which apps a member can use.
- Moving an existing library in, and getting it back out.
- Securing the server and knowing what leaves it.

Everything else is minor.

## What the comparable sites carry

| Project | Landing shape and "who it's for" | Install path, requirements | FAQ | Screenshots, demo | Apps list | Security, privacy | Help, community | Contributing, roadmap, changes | Comparison |
|---|---|---|---|---|---|---|---|---|---|
| Navidrome ([site](https://www.navidrome.org/)) | "Your Personal Streaming Service", feature blocks, install CTA | [Installation](https://www.navidrome.org/docs/installation/) per OS plus Docker, community packages, managed hosting. The only stated requirement is ffmpeg, with no RAM or CPU figures | [FAQ](https://www.navidrome.org/docs/faq/), 7 questions, including "How can I upload music to Navidrome?" and "Where are the logs?" | Demo in the top nav | [Apps](https://www.navidrome.org/apps/): 89 apps, filterable by platform and by Subsonic, OpenSubsonic, or Navidrome API | [Security Considerations](https://www.navidrome.org/docs/usage/admin/security/) (reverse proxy, TLS, bind address); Data Collection and Backup pages under Administration ([docs index](https://www.navidrome.org/docs/)) | Reddit, Discord, forum | GitHub PRs invited. No roadmap or changelog on the site | None. The FAQ points to "alternative music servers" for folder browsing |
| gonic ([README](https://github.com/sentriz/gonic)) | README only: "Free-software subsonic server API implementation, supporting its many clients" | Source, Docker, systemd, packages. Full config table | None | Screenshots | "Tested clients" list | None | IRC #gonic | GitHub | None |
| Jellyfin ([site](https://jellyfin.org/)) | Tagline, "What is Jellyfin", screenshots, four values (including "Privacy Focused") | [Docs](https://jellyfin.org/docs/): Quick Start, Installation, Post-Install Setup ([wizard, Networking, Transcoding](https://jellyfin.org/docs/general/post-install/setup-wizard)) | [FAQ](https://jellyfin.org/docs/general/faq/): "Why fork Emby?", "how do I get help?", "Why is my media not showing up?" | Four screenshots | [Clients](https://jellyfin.org/docs/general/clients/), with inclusion criteria | Values statement only | "Getting Help" docs section, forum | Contribute page, blog | Lineage (Emby) explained in FAQ and About |
| Funkwhale ([site](https://www.funkwhale.audio/)) | Audience segments (music lovers, artists, community) | Three paths: join a pod, self-host, hosted providers | None on the landing page | None noted | Subsonic app support noted | None noted | Help, contact | Contribute, donate; docs carry "Changes" ([docs](https://docs.funkwhale.audio/)) | None |
| Ampache ([site](https://ampache.org/)) | One-sentence description, four feature blocks, a large screenshot gallery | Download, Docker, Wiki | In the wiki | Demo in nav, extensive screenshots | API page | None noted | Footer links | Wiki | None |
| Airsonic-Advanced ([README](https://github.com/kagemomiji/airsonic-advanced)) | Fork lineage (Subsonic, Libresonic, Airsonic) and "differences from Airsonic" | Java version requirements per release line | None | None | None | None | GitHub Discussions | GitHub | Lineage and differences, stated factually |
| Immich ([site](https://immich.app/), [docs](https://docs.immich.app/overview/welcome)) | Nav hub; docs open with Quick start, Comparison, Support, Help Me! | [Quick start](https://docs.immich.app/overview/quick-start) ends at the first admin, the mobile app, and a backup warning. [Requirements](https://docs.immich.app/install/requirements): RAM and CPU minimums, filesystem, "never network shares" | FAQ linked in nav | None noted | Mobile apps | Privacy Policy in nav | Discord, Reddit, "Help Me!" | Roadmap linked in nav | [Comparison](https://docs.immich.app/overview/comparison): one neutral paragraph linking an external table |
| Paperless-ngx ([docs index source](https://raw.githubusercontent.com/paperless-ngx/paperless-ngx/main/docs/index.md)) | One sentence, Get started and Demo buttons, a screenshot, features, "a history" of its predecessors | setup.md; Raspberry Pi notes | [faq.md](https://raw.githubusercontent.com/paperless-ngx/paperless-ngx/main/docs/faq.md), including "Let's say I want to switch tools in a year. Can I easily move to other systems?" and "is my data private?" | Screenshot gallery, public demo | Scanner compatibility list (wiki) | "Your data ... is never transmitted or shared" | Support section: Discussions, Matrix, how to file bugs and feature requests | Contributing section; changelog.md and migration-v3.md in the docs | Lineage only |
| Home Assistant ([site](https://www.home-assistant.io/)) | Tagline, values ("All your smart home data stays local"), community size | [Installation](https://www.home-assistant.io/installation/) helps the reader choose: a recommended path, hardware per path, a method comparison table. [Getting started](https://www.home-assistant.io/getting-started/) runs from install and onboarding to "Join the community" | Yes | Live demo | Integrations | [Security](https://www.home-assistant.io/security/): disclosure process, response time, scope, past advisories | Forum, chat, newsletter | Release-notes blog; a public roadmap | None |
| Audiobookshelf ([site](https://www.audiobookshelf.org/)) | "Self-hosted audiobook and podcast server" | Docs, Docker | [FAQ](https://www.audiobookshelf.org/faq): "Where can I get help?", "How should I report security issues?", "Is the project dead?" | Showcase | Mobile apps | Security reporting answered in the FAQ | Discord | Contributing page | None |

Patterns that recur across the survey:

- Every project names a place to get help, from Navidrome to Audiobookshelf.
- Every project with a streaming API lists or tests apps: Navidrome, gonic, and Jellyfin.
- Stranger-first projects put requirements before install and end getting-started at the first
  real use: Immich and Home Assistant.
- The older projects tell their lineage and their differences from predecessors plainly, without
  running down anyone: Airsonic-Advanced, Paperless-ngx, and Jellyfin.

## Findings

Findings are ranked by consequence to a stranger's success.

### P1 (major): no place to get help or report a problem

**Location:** `2026-09-25-project-site-design.md:50-62` (page table), `:62` (About).

**Gap:** No page says where an operator goes when setup fails, how to report a bug, or how to
report a security problem. Every surveyed project answers these:

- Navidrome links Reddit, Discord, and a forum.
- Immich has "Help Me!" and Jellyfin has "Getting Help".
- Paperless-ngx has a Support section that splits bugs from feature requests.
- The Audiobookshelf FAQ answers "Where can I get help?" and "How should I report security
  issues?".
- Home Assistant has a full security disclosure page.

A stranger whose `dubplate init` fails, with no channel and no troubleshooting page, has nowhere
to go. For an internet-facing server that holds members' email addresses, a missing disclosure
address is a real risk.

**Fold:** Add a "Get help" page (site-owned). It names the channel or channels, says what to
include in a report, and gives a security contact with a stated scope. Home Assistant's scope is
"latest stable versions" only.

**OWNER FORK: what support a solo project offers.** It depends on the `oss-pair` ruling.

- (a) Public issue tracker plus a security email.
- (b) Security email only, with the site saying plainly that support is best-effort.
- (c) A community channel as well.

Recommendation: (b) at 0.8 if the source stays private, and (a) once it is open. Either way, a
security contact exists on day one.

### P2 (major): members have no list of apps that work

**Location:** `:57` (For members: "listening with an OpenSubsonic app").

**Gap:** The first question a member asks is "which app do I install on my phone?". Three
surveyed projects answer it with a list:

- Navidrome's [Apps](https://www.navidrome.org/apps/) page lists 89 apps by platform and API
  flavor.
- gonic keeps a "Tested clients" list.
- Jellyfin's [Clients](https://jellyfin.org/docs/general/clients/) page states its inclusion
  criteria.

dubplate adds a second reason. Members sign in by magic link, so the page must say how an
OpenSubsonic app authenticates. The page cannot ship without that answer, whatever 0.8 decides.

**Fold:** Add a "Listening apps" section to For members, or a sibling page. List a small set of
apps actually tested against 0.8, by platform, each with the date tested and the app's own link,
plus the sign-in step for apps. Keep it short and tested, the way gonic does. Don't try to match
Navidrome's catalogue: an untested list rots and hands a stranger a failure.

### P3 (major): no path for bringing an existing library in, or taking it back out

**Location:** `:55` (Set up a server), `:56` (Run your server).

**Gap:** ROADMAP (`packaging`, line 192) moves "the Navidrome import" into 0.8. A stranger most
likely arrives with a library already: files on disk, or a Navidrome or beets setup. The page map
has no page for that path. The surveyed projects cover both directions:

- Immich's docs open with "If you're new here and came from other ... alternatives"
  ([comparison](https://docs.immich.app/overview/comparison)).
- Paperless-ngx carries `migration-v3.md` and the FAQ entry "Let's say I want to switch tools in
  a year. Can I easily move to other systems?"
  ([faq.md](https://raw.githubusercontent.com/paperless-ngx/paperless-ngx/main/docs/faq.md)).
- Airsonic-Advanced documents database migration between release lines
  ([README](https://github.com/kagemomiji/airsonic-advanced)).

The exit question is dubplate's strength: plain FLAC in a consistent layout. Saying so builds the
deciding reader's trust at no cost.

**Fold:** Add a dubplate-owned "Bring your library" guide covering the Navidrome import and a
folder of existing albums, including what gets turned away and why (link the reject page). Add a
short "Leaving dubplate" section (what is on disk and in what layout) to Run your server or to
the is/isn't page.

### P4 (major): securing the server and what leaves it are unassigned

**Location:** `:55-56`, `:62`.

**Gap:** Navidrome has a
[Security Considerations](https://www.navidrome.org/docs/usage/admin/security/) page (reverse
proxy, TLS, bind address) and a Data Collection page ([docs index](https://www.navidrome.org/docs/)).
Several other sites state what they send outward:

- Jellyfin lists "Privacy Focused" among its four values ([site](https://jellyfin.org/)).
- Paperless-ngx says data "is never transmitted or shared in any way, unless you explicitly
  choose to do so" ([index.md](https://raw.githubusercontent.com/paperless-ngx/paperless-ngx/main/docs/index.md)).
- Home Assistant says "All your smart home data stays local" ([site](https://www.home-assistant.io/)).

dubplate runs a TLS listener (0.8), holds member emails, sends mail over SMTP, and queries
MusicBrainz. An operator who is answerable to their group needs one place that covers:

- What is exposed and how to harden it.
- Which outbound calls happen: MusicBrainz, cover art, mail, the backup remote.
- That nothing else leaves the server.

**Fold:** Add a "Security and privacy" section to Run your server (dubplate-owned, so it tracks
the code). The outbound-call list is a candidate for generation, like the config reference. Link
it from About and from P1's security contact.

### P5 (minor, but high-cost if missed): backups are left to "Run", not setup

**Location:** `:55` versus `:56`.

**Gap:** Immich's [quick start](https://docs.immich.app/overview/quick-start) puts its backup
warning inside the install path: "You must setup manual backups of the images and videos...". On
dubplate the server's store is the library's record ("Facts agents get wrong" in CLAUDE.md). A
group that loses the box before anyone reads "Run your server" loses the library.

**Fold:** End Set up a server with configuring and testing a backup, so the path reads first
admin, first member, first backup. Restore stays in Run.

### P6 (minor): requirements are folded into setup

**Location:** `:55` ("What you need").

**Gap:** Immich ([requirements](https://docs.immich.app/install/requirements)) and Home Assistant
([installation](https://www.home-assistant.io/installation/)) put requirements where a deciding
reader finds them before committing: RAM and CPU minimums, filesystem constraints, hardware per
path. The dubplate list is longer than most, since the stranger has to line these up before
starting:

- A VPS, with arm64 or amd64.
- A domain and TLS.
- An SMTP provider.
- A backup remote.
- Storage sized for FLAC.

Folding them into setup is acceptable as long as "What you need" is the guide's first,
linkable section and the Home page and the is/isn't page link to it.

**Fold:** State that placement and linking in the spec, and have the section give storage sizing
in albums per gigabyte, not bare numbers.

### P7 (minor): release notes are site-owned, which breaks the spec's own drift rule

**Location:** `:61`, `:102-106`.

**Gap:** The spec vendors guides so they can't drift from a tag, but release notes, including
upgrade steps and breaking changes, are hand-written site posts. Paperless-ngx keeps
`changelog.md` and `migration-v3.md` in the docs tree beside the code
([docs listing](https://github.com/paperless-ngx/paperless-ngx/tree/main/docs)). Funkwhale's
docs carry "Changes" ([docs](https://docs.funkwhale.audio/)). Home Assistant publishes a
release-notes blog ([site](https://www.home-assistant.io/)). The upgrade steps are what
`dubplate upgrade` users need, and they belong to the tag.

**Fold:** Keep the changelog and upgrade notes in dubplate's `docs/guide/`, vendored at the pin.
The site post is the announcement and links to them.

### P8 (minor): no "what's coming" line

**Location:** `:53`.

**Gap:** Immich and Home Assistant link a roadmap. For dubplate, the fit-deciding facts are
future ones: FLAC-only at 0.8, other formats at 1.0, the local client at 0.9. A reader with an
MP3 library needs to know that the answer is "not yet", not "never". The standing rule against
dates allows a list without dates.

**OWNER FORK:**

- (a) One undated "Coming later" list on the is/isn't page.
- (b) A roadmap page.
- (c) Nothing public.

Recommendation: (a). It answers the fit question without making a promise.

### P9 (minor): no FAQ

**Location:** page table.

**Gap:** Navidrome, Jellyfin, Immich, Paperless-ngx, Audiobookshelf, and Home Assistant all carry
an FAQ. The spec's is/isn't page and comparison page absorb most of what an FAQ does. The rest
(logs, email not arriving, "can I avoid running a server?", which Navidrome's
[FAQ](https://www.navidrome.org/docs/faq/) answers with managed hosting) fits in P1's Get help
page as a short troubleshooting section.

**Fold:** None beyond P1. A standalone FAQ at 0.8 would be padding. The spec is right to leave
it out.

### P10 (minor): screenshots aren't named

**Location:** `:52` ("the look"), `:147` (no demo server).

**Gap:** Every landing page surveyed except Immich's and gonic's shows the product. Jellyfin
shows four screenshots, and Ampache and Paperless-ngx show galleries. Excluding a demo server is
right for a private-by-design product; Navidrome, Ampache, Paperless-ngx, and Home Assistant run
public demos, but dubplate's "isn't a public music site" line makes a demo a contradiction.
"The look" is vague.

**Fold:** Name the screenshots Home carries: the library, an album page, and the uploader
turning an album away. The last one sells the uploader, which none of the peers has.

### Choices where the spec already matches or beats prior art

- **Split by reader** (`:30-35`). This matches Funkwhale's user and admin docs split
  ([docs](https://docs.funkwhale.audio/)) and beats Navidrome's single tree.
- **Generated references** (`:79-83`). gonic keeps its config table by hand in its README. A
  generated reference is strictly better.
- **Reject page** (`:58`). No peer has a member-facing "why was this refused" page. Navidrome's
  FAQ answer to "How can I upload music?" shows the gap it fills.
- **Comparison page** (`:70-77`). This is more than any peer does. Immich links to an external
  table, and Jellyfin, Airsonic-Advanced, and Paperless-ngx only tell their lineage. The rules
  (own terms, cited, dated, no quality claims) are the right guard. One suggestion: open the page
  with a one-sentence lineage and credit line in the Airsonic-Advanced and Paperless-ngx manner,
  since dubplate grew up on Navidrome.
- **Contributing.** It is absent, correctly, until `oss-pair` rules.

## Candidates for "How dubplate compares"

Each description is quoted from the project's own site, checked 2026-09-25. The "fits better
when" lines state purposes only, never quality.

1. **Navidrome.** "Navidrome can be used as a standalone server, that allows you to browse and
   listen to your music collection using a web browser. It can also work as a lightweight
   Subsonic-API compatible server" ([overview](https://www.navidrome.org/docs/overview/)).
   *Fits better when* you already keep your music organized on disk yourself and want a
   streaming server over it. Navidrome says it "does not write to your music folder or the
   files by design" ([FAQ](https://www.navidrome.org/docs/faq/)). It also fits better when you
   want the widest tested app list ([Apps](https://www.navidrome.org/apps/)) or managed hosting.
   Credit it by name as the group's server before 0.8.
2. **gonic.** "Free-software subsonic server API implementation, supporting its many clients"
   ([README](https://github.com/sentriz/gonic)). *Fits better when* you want a minimal
   Subsonic server with browsing by folder, podcasts, or jukebox mode, over a library you
   organize yourself.
3. **Jellyfin.** "The volunteer-built media solution that puts you in control of your media"
   ([site](https://jellyfin.org/)). *Fits better when* one server should hold the household's
   movies, shows, photos, and books alongside the music.
4. **Funkwhale.** "a self-hosted audio player and publication platform. It enables users to
   build libraries of existing content and publish their own"
   ([docs](https://docs.funkwhale.audio/)). *Fits better when* you want to publish your own
   audio or podcasts, or to share across federated servers. dubplate's "isn't a network of
   servers" line is the other side of this.
5. **Lyrion Music Server.** "open-source software which can control and serve (stream) music to
   a wide range of physical and virtual audio players called Squeezeboxes"
   ([site](https://lyrion.org/)). *Fits better when* the goal is whole-house playback on
   Squeezebox-family or other networked players.
6. **Koel** (optional). "Music streaming solution that works." ([site](https://koel.dev/)).
   *Fits better when* one person wants a web-first personal streamer. It has a paid "Koel Plus"
   tier, which the page should state neutrally if it lists Koel.

Recommendation for the Phase A plan: take 1 through 5. Koel and Ampache overlap Navidrome's
"fits better" case, and they add re-verification cost without helping a reader choose.
