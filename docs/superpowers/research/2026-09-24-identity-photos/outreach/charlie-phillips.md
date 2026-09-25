# Charlie Phillips (Charlie Phillips Heritage Archive)

Status: blocked. The contact route failed verification: `charliephillipsarchive.com` has no MX
record (checked again 2026-09-25 at the plan fold), so mail to the archive address may bounce.
The fallback route is Nicky Akehurst (below), who represents him for talks. Geoff decides whether
to send through her. Nothing has been sent.

Checked again at the plan fold, 2026-09-25: `dig MX charliephillipsarchive.com` returns nothing.
The four V&A pages return 200.

## Contact route

- Primary: `info@charliephillipsarchive.com`, the address on the archive's contact page, which
  invites "general enquiries". Fetched 2026-09-25 (HTTP 200):
  <https://www.charliephillipsarchive.com/contact-2/>.
- **Deliverability not confirmed.** The page's own link points at
  `info@www.charliephillipsarchive.com`, and the copyrights page names the operator as
  `curator@www.charliephillipsarchive.com`
  (<https://www.charliephillipsarchive.com/copyrights/>). Both carry a stray `www.`. A DNS
  lookup on 2026-09-25 found no MX record for `charliephillipsarchive.com` (SOA at DreamHost,
  A record present). Mail may still be delivered to the A record, but a bounce is possible.
- Fallback: Nicky Akehurst, Akehurst Creative Management, `na@nickyakehurst.com`, from
  <https://nickyakehurst.com/contact/index.html> (fetched 2026-09-25). The Akehurst page for
  Phillips says "Charlie is available for talks and presentations - contact Nicky Akehurst for
  further information" (<https://nickyakehurst.com/photographers/CharliePhillips/index.html>).
  Akehurst represents him for talks; the page does not say she handles image rights.
- Current terms: the copyrights page says none of the site's content "may be downloaded,
  copied, reproduced, transmitted, stored, sold or distributed without the prior written consent
  of the copyright holder."
- Wikipedia gives his birth date as 22 November 1944 and lists no death date
  (<https://en.wikipedia.org/wiki/Charlie_Phillips_(photographer)>, raw fetched 2026-09-25).

## Images asked for

All four are prints held by the V&A, listed through the V&A collections API on 2026-09-25.
The V&A holding a print does not settle who holds copyright; the letter asks Phillips.

| Title (V&A) | Year | V&A page |
|---|---|---|
| Notting Hill Couple | 1967 | <https://collections.vam.ac.uk/item/O1225389/> |
| Cue Club Regulars | 1966 | <https://collections.vam.ac.uk/item/O1225386/> |
| Notting Hill Carnival | 1968 | <https://collections.vam.ac.uk/item/O1783696/> |
| Portobello Road | 1974 | <https://collections.vam.ac.uk/item/O1225393/> |

## Letter
**Subject:** Four of your Notting Hill photographs for a small open-source music server

Dear Mr Phillips,

My name is Geoff. I'm building dubplate, a free music server that lets a group of friends share
one well-organized library. It's open source, released under the MIT license. The sign-in page
shows a photograph from a music scene, and a credits page lists every photo. Each one is always
credited to the person who took it.

Your photographs of Notting Hill's West Indian community in the 1960s and 70s show a London I
want the people using dubplate to see. I'd love to include four of them: Notting Hill Couple
(1967), Cue Club Regulars (1966), Notting Hill Carnival (1968), and Portobello Road (1974).

Would you consider releasing those four under CC BY 4.0? CC BY lets anyone reuse and adapt a
photo, commercially too, as long as they credit you. If you agree, a line on the archive's site
naming the photos and the license would let anyone check it. I'd need files at least 1200 pixels
on the long side.

If that isn't something you'd want, would you allow them in an optional photo pack instead? The
pack is planned but has no date yet. It would be a separate download that people who run
dubplate could choose to add, on terms you set.

Thank you for the archive, and for the years of work behind it.

Geoff
