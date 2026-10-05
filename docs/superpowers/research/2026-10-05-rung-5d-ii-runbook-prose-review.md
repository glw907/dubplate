# Rung 5d-ii Task 7: the runbook's prose-voice review

`prose-voice-reviewer` (Opus) over `docs/server-runbook.md` at `5ca9fc1`, 2026-10-05, against the
Google developer documentation standard. tellgrader: 3835 words, cadence CV 0.74, one
bold-lead-bullets hit (:138), one tricolon. Flags, exit codes (0/1/69/75/130), and defaults (20, 60,
5, 8080, 127.0.0.1) were checked against `--help` and match. No real IP, hostname, or personal
address appears; examples use RFC 5737/3849 and example.org.

Conductor ruling: fold all fifteen; none refused.

## Blockers (facts and contradictions)

1. :9-10 contradicts :489. The intro says every command ran against a scratch state dir; :489 says
   the box's systemd has not been shown to accept the property list, and neither the `systemd-run`
   blocks nor `deploy.sh --apply` can run against a scratch dir. Rewrite: "The `dubplate-server`
   commands were run against a scratch state directory, and their flags match `--help`. The
   `systemd-run` and `deploy.sh` commands have not yet run on a fresh host."
2. :57-60 vs :64-76 and :81-82. The text says the block carries the serve unit's properties but it
   omits `RestrictAddressFamilies`, so step 6's "without `RestrictAddressFamilies`" removes nothing.
   Either add `-p 'RestrictAddressFamilies=AF_INET AF_INET6 AF_UNIX'` to step 5's block and have step 6
   drop that line, or say in step 5 that the block leaves it off and why.
3. :20-28, :444-462. `deploy.sh` is this estate's script: it defaults HOST to the maintainers' server,
   reads the workstation age store, renders the incumbent's `/etc/dubplate/env`, brings up compose,
   and installs beets. The template it renders carries the estate's public URL and site name, and the
   runbook sends strangers to it at :25. Say plainly that `deploy.sh` deploys the maintainers' estate,
   and give strangers the manual path: copy the binary and unit, write `server.env` by hand, and run
   `install -d -o music -g music -m 0700` for each leaf. Flag the template's estate values as values a
   stranger replaces.
4. :24-36. "Each step depends on the one before it" is false of `deploy.sh`'s own order (leaves, then
   `unidecode install`, then the swap, then `server.env`; deploy.sh:396-562); the table install needs
   no env file. Rewrite: "Run the steps in this order. Steps 5 and later need the ones before them."
5. :473-475 says the staged install carries "the shadow unit's hardening" and differs only by
   `Environment=`. deploy.sh:487-497 shows four differences: `Environment=`, a wider
   `ReadWritePaths`, AF_INET/AF_INET6 for the download, and no `UMask`. Name all four, or say "under
   `systemd-run` with the property list `--check` prints".
6. Estate leakage. :235-236 ("Hetzner among them", "the route this project verified on its own box");
   :350 ("rung 12", an internal ladder citation; rewrite "It arrives with the streamer driver in a
   later release"); the shadow unit and timer (:59, :467), `/etc/dubplate/env`, and "the compose
   stack" (:512, :522-523) are incumbent-estate artifacts a stranger lacks. Keep "Some VPS providers
   block outbound 465 by default; check from the host." Mark the third rollback copy "(only on a host
   that also runs the legacy stack)."
7. :461-462. "Mint each one first" is estate jargon. Rewrite: "Add each missing name to your secrets
   store, then run `--check` again."

## Warnings (structure)

8. :79-80. "Swap the last line" is wrong; the verb spans :75-76. Write "Replace the `member add ...`
   arguments with `member list`." Drop `doctor` there; it duplicates step 6.
9. :138-173 and the same reflex at :244-261, :319-346, :372-387, :395-440: bold leads open most
   bullets. Make the boundary rows `####` subheads or a definition table; keep bold only where it
   labels a field; make the backup prerequisites a numbered checklist.
10. :16, :105, :192, :265, :369, :391, :444: the "**Brief.**" label opens every section. Delete the
    label and keep the sentence.
11. :259. Name the per-address limit: "per-address limits
    (`DUBPLATE_LINK_REQUESTS_PER_ADDRESS_PER_HOUR`, default 5)".

## Suggestions (voice)

12. :82-83. "The reason, in one sentence:" is a setup-colon payoff. Rewrite: "The `bind addresses` row
    reads the interface table over a netlink socket, which the restriction denies."
13. :301. "This is the field's normal answer, not an invention of this project." is a contrast frame.
    Rewrite: "Other self-hosted servers take the same approach."
14. :171. Cut "That limit is the whole point of the row"; start at "The row dials the server
    directly...".
15. :4-6. Replace the section inventory with "It covers install, `doctor`, mail, the trust boundary,
    backup, and deploy."
