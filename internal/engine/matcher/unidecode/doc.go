// Package unidecode owns the Unidecode 1.4.0 transliteration table:
// its binary format, its pinned wheel and table checksums, a Go port
// of the extraction that turns the wheel into that format, and a
// lookup that reproduces Unidecode's replacement rules. It also
// installs the table into the state directory ([Install]) and renders
// doctor's row for it ([DoctorRows]). Beyond the standard library it
// imports internal/atomicfile and internal/doctor only, so it carries
// no data of its own; every table [Build] produces or [Parse] accepts
// is verified against a checksum pinned in this package's source.
package unidecode
