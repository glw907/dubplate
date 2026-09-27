// Package unidecode owns the Unidecode 1.4.0 transliteration table:
// its binary format, its pinned wheel and table checksums, a Go port
// of the extraction that turns the wheel into that format, and a
// lookup that reproduces Unidecode's own replacement rules. It imports
// only the standard library, so it carries no data of its own; every
// table [Build] produces or [Parse] accepts is verified against a
// checksum pinned in this package's source.
package unidecode
