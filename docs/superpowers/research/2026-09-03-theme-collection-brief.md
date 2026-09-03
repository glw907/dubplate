# Theme collection brief: cross-platform palette matching

Research brief, 2026-09-03. Firms up the theme roster named in the
greenfield build spec's ladder entry 7 theming ruling
(`docs/superpowers/specs/2026-09-02-greenfield-server-build.md:602-613`)
and entry 12's estate note (same file, lines 675-680): the roster
grows a matching DaisyUI theme for each streamer a driver ships,
authored as a CSS variable block from a published palette standard,
never as code (the theming-is-data constraint, same file lines
593-594). The estate pair is already ratified: custom DaisyUI
gruvbox, paired with Navidrome's built-in Gruvbox Dark. This brief
checks what's free, custom-feasible, or not expressible at all,
across the four platforms the ROADMAP names as supported-by-design
(`ROADMAP.md:38-40`): DaisyUI, Navidrome, Jellyfin, Subsonic-family.
Every claim is cited; GitHub API pulls (stars, tree listings) were
read live on 2026-09-03.

## 1. Navidrome roster, pinned v0.63.2

The full theme directory
(github.com/navidrome/navidrome/tree/v0.63.2/ui/src/themes, read via
the GitHub contents API) lists 29 themes. Six rest on a published
external palette standard; the rest are Navidrome-original basics
(`dark`, `light`, `extradark`, `ligera`, `green`, `amusic`,
`SquiddiesGlass`, `electricPurple`, `moonbaseAlpha`, `moonbaseBravo`,
`nautiline`, `nuclear`, `nutball`) or brand-adjacent (`spotify`,
imitating the Spotify app; `monokai`, a widely ported but
loosely-specified editor scheme with no single canonical hex table,
unlike the six below).

| Standard | Navidrome file(s) | Canonical spec |
|---|---|---|
| Gruvbox | `gruvboxDark.js` | [morhetz/gruvbox](https://github.com/morhetz/gruvbox); hex table at [gruvbox-contrib/color.table](https://github.com/morhetz/gruvbox-contrib/blob/master/color.table) |
| Nord | `nord.js` | [nordtheme.com/docs/colors-and-palettes](https://www.nordtheme.com/docs/colors-and-palettes) ([nordtheme/nord](https://github.com/nordtheme/nord)) |
| Dracula | `dracula.js` | [spec.draculatheme.com](https://spec.draculatheme.com/) ([draculatheme.com/spec](https://draculatheme.com/spec)) |
| Catppuccin | `catppuccinLatte.js`, `catppuccinMacchiato.js` | [catppuccin/catppuccin style-guide](https://github.com/catppuccin/catppuccin/blob/main/docs/style-guide.md) ([catppuccin/palette](https://github.com/catppuccin/palette)) |
| Rosé Pine | `rosePine.js`, `rosePineDawn.js`, `rosePineMoon.js` (all 3 variants) | [rosepinetheme.com/palette](https://rosepinetheme.com/palette/) |
| Tokyo Night | `tokyoNight.js`, `tokyoNightLight.js` | [tokyo-night/tokyo-night-vscode-theme](https://github.com/tokyo-night/tokyo-night-vscode-theme) (origin, enkia); values re-popularized and lightly diverged by [folke/tokyonight.nvim](https://github.com/folke/tokyonight.nvim) — no single site plays the role nordtheme.com or draculatheme.com play for the others (see open question 2) |

Catppuccin ships only 2 of the standard's 4 flavors (Latte, Macchiato;
missing Frappé and Mocha). Rosé Pine and Tokyo Night ship complete.

## 2. Jellyfin theming model

Built-in choice is Light/Dark only (jellyfin.org's own docs and the
awesome-jellyfin survey; no built-in palette-standard themes).
Jellyfin ships a genuine custom-CSS hook: a stylesheet field at
Dashboard > Branding (server-wide) or Settings > Display (per-user),
loaded after the default stylesheet so it can override any rule
(jellyfin.org/docs/general/clients/css-customization/). Caveats from
the same page: custom CSS affects only clients running Jellyfin Web,
externally hosted CSS assets are fetched by the client (must be
reachable from the whole client fleet), and support isn't provided
for issues the custom CSS causes.

That hook upgrades Jellyfin matching from "closest built-in" (dark
vs. light, no palette match at all) to exact: paste an `@import` and
Jellyfin repaints in the target palette, same as DaisyUI. Existing
community ports, by maintenance signal (stars / last push, read
2026-09-03):

| Standard | Port | Maintenance |
|---|---|---|
| Catppuccin | [catppuccin/jellyfin](https://github.com/catppuccin/jellyfin) (official org repo, all 4 flavors, `@import url('https://jellyfin.catppuccin.com/theme.css')`) | 127 stars, pushed 2026-08-06 |
| Nord | [theme-park.dev](https://docs.theme-park.dev/themes/jellyfin/) (`theme-park.dev/css/base/jellyfin/nord.css`) | theme.park aggregator: 3,087 stars, pushed 2026-08-28 |
| Dracula | same theme-park.dev aggregator | same |
| Gruvbox | [derektata/jellyfin-themes](https://derektata.github.io/jellyfin-themes/) ("Gruvbox Dark") | 1 star, pushed 2026-06-27 — thin, unproven |
| Tokyo Night | same derektata repo ("Tokyo," a loose match, not confirmed against either canonical source) | same 1-star repo |
| Rosé Pine | none found in the [awesome-jellyfin THEMES.md](https://github.com/awesome-jellyfin/awesome-jellyfin/blob/main/THEMES.md) survey | n/a — would be authored cold |

## 3. Subsonic family

Three distinct surfaces, not one. Original commercial Subsonic (Sindre
Mehus) is unmaintained past v6, out of scope; its community skins
(Rolleshark et al., on userstyles.org) target a codebase the roadmap
doesn't drive against.

**Airsonic-Advanced** inherits Airsonic's fixed Spring/JSP theme
system verbatim: a `themes.txt` enum of 27 named `.properties` files
(github.com/airsonic/airsonic/blob/.../theme/themes.txt) — `default`,
`midnight`, `black`, `sandstorm`, `barents`, several `hd*` variants,
etc. — none matching a published palette standard, and no live
custom-CSS field; [airsonic-advanced/airsonic-advanced](https://github.com/airsonic-advanced/airsonic-advanced)
carries the identical `theme/*.properties` directory. Adding a theme
means shipping a new `.properties` + JSP asset in a fork or upstream
PR, a code change, not data. The repo last pushed 2024-04-24 (1,403
stars) against Jellyfin's 56,504 and Navidrome's 23,320 (all read via
the GitHub API, 2026-09-03) — stale as well as closed to data-only
theming.

**gonic** ships no music-browsing web UI; its README states the web
interface is "for configuration (set up last.fm, manage users, start
scans, etc.)" only (github.com/sentriz/gonic). Anyone using gonic
experiences the library through a Subsonic client, so gonic's "look"
is whatever client they picked, not gonic itself.

**Feishin** is that client in practice for a self-hosted, desktop-
first estate (9,734 stars, pushed 2026-09-03), and already ships
nearly the entire target roster built in, no authoring required. Its
theme docs (github.com/jeffvli/feishin/blob/development/docs/CUSTOM_THEMES.md)
list 32 pre-configured themes including `gruvboxDark`/`gruvboxLight`,
`nord`, `catppuccinMocha`/`catppuccinLatte` (2/4 flavors, the
opposite two from Navidrome's), `dracula`, `tokyoNight`, and all
three `rosePine` variants. Custom themes are JSON + optional CSS
stylesheet files (desktop/Electron build only) that can `extend` a
built-in.

## 4. DaisyUI 5 side

The full built-in roster (daisyui.com/docs/themes/) is 35 themes
including `nord` and `dracula` by name — those two are free today,
no authoring. Gruvbox, Catppuccin, Rosé Pine, and Tokyo Night are not
built in; each needs a CSS variable block. Community ports, checked
live via the GitHub API (2026-09-03):

- **Catppuccin**: [catppuccin/daisyui](https://github.com/catppuccin/daisyui), the Catppuccin org's own first-party port, npm `@catppuccin/daisyui`. 274 stars, pushed 2026-09-02 (yesterday) — the most actively maintained port found in this survey, of any standard, on any platform.
- **Rosé Pine**: [rose-pine/daisyui](https://github.com/rose-pine/daisyui), also first-party (rosepinetheme.com/themes/daisyui/), all three variants. 22 stars, pushed 2025-11-05 (10 months stale but not abandoned).
- **Gruvbox, Tokyo Night**: no first-party org port found. [dachinat/daisyui-themes](https://github.com/dachinat/daisyui-themes) is a single-author collection (15 stars, pushed 2026-01-23) carrying `gruvbox.css`, `tokyonight.css`, `rosepine.css`, and `catppuccin.css` all in the current DaisyUI 5 CSS-first block shape — a usable crib source for gruvbox and tokyo-night specifically, since no better one turned up.

DaisyUI 5's theme block defines a fixed variable set (same docs
page): `--color-base-100/200/300`, `--color-base-content`,
`--color-{primary,secondary,accent,neutral,info,success,warning,error}`
and each color's paired `-content`, plus non-color
`--radius-selector/field/box`, `--size-selector/field`, `--border`,
`--depth`, `--noise`, and the `name`/`default`/`prefersdark`/
`color-scheme` block metadata. The dachinat gruvbox file (fetched
directly) confirms this is exactly the surface the ruling's "CSS
variable block" language means: a `@plugin "daisyui/theme" { name:
"gruvbox"; color-scheme: "dark"; --color-base-100: #282828;
--color-primary: #458588; … }` block, one declaration per variable,
no selectors or logic.

## (a) Cross-platform matrix

Cells: **built-in** (ships free), **custom-feasible** (a hook exists;
notes whether a ready port exists or it'd be authored cold),
**not-expressible** (no data-only path at all).

| Standard | DaisyUI | Navidrome | Jellyfin | Subsonic-family |
|---|---|---|---|---|
| Gruvbox | custom-feasible (dachinat port, §4) | **built-in** | custom-feasible (1-star port only, §2) | Airsonic-Adv: not-expressible · gonic: n/a (no UI) · Feishin: **built-in** |
| Nord | **built-in** | **built-in** | custom-feasible (theme-park.dev, 3,087★) | Airsonic-Adv: not-expressible · gonic: n/a · Feishin: **built-in** |
| Dracula | **built-in** | **built-in** | custom-feasible (theme-park.dev) | Airsonic-Adv: not-expressible · gonic: n/a · Feishin: **built-in** |
| Catppuccin | custom-feasible (official org port, 274★) | **built-in** (2/4 flavors) | custom-feasible (official org port, 127★) | Airsonic-Adv: not-expressible · gonic: n/a · Feishin: **built-in** (2/4, opposite flavors) |
| Rosé Pine | custom-feasible (official org port, 22★) | **built-in** (3/3) | custom-feasible, no port found, author cold | Airsonic-Adv: not-expressible · gonic: n/a · Feishin: **built-in** (3/3) |
| Tokyo Night | custom-feasible (dachinat port only) | **built-in** (2/2) | custom-feasible (1-star port only) | Airsonic-Adv: not-expressible · gonic: n/a · Feishin: **built-in** |

Airsonic-Advanced is not-expressible for every standard by the same
structural reason (§3): a closed `.properties` enum, not a CSS hook.

## (b) Recommended collection: lean, popularity-weighted

The matrix above is the full evidence surface; the shipped roster
stays small on purpose — ladder entry 7's tier 3 (an optional
operator CSS file served after the bundle, same spec file lines
598-601) is the designed escape hatch for anything past this set, so
"not on the default roster" never means "not possible."

**Ships at rung 7, alongside the estate gruvbox pair:**

1. **Gruvbox** — already ratified, DaisyUI custom + Navidrome
   built-in. No change.
2. **Nord** — built into DaisyUI *and* Navidrome *and* Feishin today,
   zero authoring cost; Jellyfin needs only a paste from a
   well-used aggregator (theme-park.dev, 3,087★).
3. **Dracula** — identical free-on-three-sides shape to Nord, same
   Jellyfin path.
4. **Catppuccin**, the one earned addition beyond the ladder-7 text's
   named pair. It clears a higher bar than Rosé Pine or Tokyo Night
   on both axes Geoff asked to weight: platform popularity and port
   quality. Jellyfin is the dominant surface by a wide margin
   (56,504 GitHub stars vs. Navidrome's 23,320, Feishin's 9,734,
   gonic's 2,525, Airsonic-Advanced's 1,403-and-stale-since-2024 —
   all read live, §3), and Catppuccin is the *only* standard here
   with an official, actively-maintained first-party port on
   Jellyfin (127★, pushed 2026-08-06). Its DaisyUI port
   (catppuccin/daisyui, 274★, pushed yesterday) is the freshest
   community theme found anywhere in this brief, and it's already
   built into Navidrome and Feishin. No other custom-feasible
   standard has first-party, actively-maintained coverage on more
   than two of the four platforms; Catppuccin has three.

**Explicitly not on the shipped roster, and why the escape hatch
covers each:**

- **Rosé Pine** — free on Navidrome and Feishin, has an official
  DaisyUI port, but no Jellyfin port surfaced (§2); a gap on the
  most-deployed platform disqualifies a *shipped* default even
  though the other three-way match is real. An adopter authors (or
  requests) the Jellyfin leg and drops the ready DaisyUI block in via
  the tier-3 operator CSS file; no new engine work needed.
- **Tokyo Night** — free on Navidrome and Feishin, but every DaisyUI
  and Jellyfin port found is single-author and low-star (dachinat,
  15★; derektata, 1★), and the standard lacks one canonical palette
  source (question 2). Lower confidence, lower payoff than
  Catppuccin at the same authoring cost.
- **Monokai, Spotify** (Navidrome extras) — not open palette
  standards with a canonical spec site; out of scope entirely.

## (c) Authoring guidance per custom theme

Binding constraint from the ratified ruling (build spec lines
593-598): every theme is a CSS variable block only, authored against
DaisyUI's theme-variable contract; no plugin ABI, no per-user
template override, config-rendered branding stays a separate tier
from the theme roster itself.

- **Gruvbox** (already shipped): derive from
  github.com/morhetz/gruvbox's hex table; crib the dachinat block's
  shape (§4) if the current file needs a diff-check against a second
  source.
- **Catppuccin** (new): derive hex values from the official
  style-guide (github.com/catppuccin/catppuccin/blob/main/docs/style-guide.md)
  or the `@catppuccin/palette` npm package, not from a secondary
  port, to avoid drift; crib the DaisyUI variable *shape* (not the
  colors) from catppuccin/daisyui's own source, since it's the
  freshest, most official reference for how Catppuccin maps onto
  DaisyUI's specific variable set (base/content triads, the
  paired `-content` colors, the non-color radius/size/border/depth/
  noise knobs listed in §4). Flavor choice is open (question 1).

Fill every variable DaisyUI's theme block defines (§4's list); a
partial block silently inherits from whatever DaisyUI treats as the
active base theme, which is a plausible source of an inconsistent
"pretty close but not quite" repaint if a future rung ever revisits
this file without checking for gaps.

## (d) Open questions for the rung 7 planner

1. **Catppuccin flavor.** Navidrome ships Latte (light) + Macchiato
   (dark) natively; Feishin ships Latte + Mocha (dark). Latte is the
   only flavor common to both, but it's light, and the estate default
   is Gruvbox *Dark*. Pick Macchiato (exact Navidrome dark match,
   diverges from Feishin's Mocha) or ship both dark flavors as
   separate entries, the way Rosé Pine's three variants would be.
2. **Tokyo Night's canonical source is contested**, unlike the other
   five: enkia's original repo (tokyo-night/tokyo-night-vscode-theme)
   versus folke/tokyonight.nvim's later, more widely-ported values.
   Name which one is canonical before ever authoring past the escape
   hatch, not after.
3. **Jellyfin CSS scope.** The hook has two independent install
   points, server-wide (Dashboard > Branding) and per-user
   (Settings > Display) (§2). Decide whether the Jellyfin driver
   (rung 12+) documents the server-wide field as the estate's install
   point, consistent with "coherent from uploader to player," or
   leaves it per-user.
4. **Airsonic-Advanced's ceiling.** Its theme system is closed to
   data-only matching by construction (§3), and the repo has been
   quiet since 2024-04-24. If ever built as a driver, decide up front
   whether "closest built-in, no palette match" is an acceptable
   documented limitation.
5. **Cheap operator starter files.** Rosé Pine and Tokyo Night are
   already free on Navidrome and Feishin; is it worth bundling their
   DaisyUI-side variable blocks as ready-to-drop tier-3 operator
   files (crib from rose-pine/daisyui and dachinat's ports) without
   growing rung 7's shipped roster?

## Addendum: matching the factory defaults (Geoff, 2026-09-03)

A requirement added after the survey above: the shipped roster also
carries a matching DaisyUI theme for each target streamer's factory
default, because the common adopter never re-themes the streamer and
should get coherence out of the box.

- Navidrome's default is its **Dark** theme, verified at v0.63.2:
  `conf/configuration.go` sets `defaulttheme` to `"Dark"`, and
  `ui/src/themes/useCurrentTheme.js` falls back to `DarkTheme` (the
  Auto option follows `prefers-color-scheme`). Palette source for
  extraction: `ui/src/themes/dark.js` and `dark.css.js`.
- Jellyfin-web's default is its **dark** theme (the `src/themes/dark`
  directory in jellyfin/jellyfin-web; siblings: light, appletv,
  blueradiance, purplehaze, wmc). Palette source for extraction:
  that theme directory's stylesheet.

Neither palette is a published standard, so both land as authored
DaisyUI variable blocks under the same theming-is-data rule as
gruvbox, derived from the apps' own theme sources at their pinned
versions. Final rung 7 roster: dark, light, dim, night, black, nord,
coffee, retro, gruvbox (estate default), catppuccin, navidrome-dark
match, jellyfin-dark match.
