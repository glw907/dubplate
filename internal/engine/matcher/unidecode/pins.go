package unidecode

import "errors"

// Version is the Unidecode release this package's table format and
// extraction logic are pinned to.
const Version = "1.4.0"

// WheelURL is Unidecode 1.4.0's wheel, at its content-addressed
// location on PyPI's file host.
const WheelURL = "https://files.pythonhosted.org/packages/8f/b7/559f59d57d18b44c6d1250d2eeaa676e028b9c527431f5d0736478a73ba1/Unidecode-1.4.0-py3-none-any.whl"

// WheelSHA256 is Unidecode 1.4.0's wheel checksum, from PyPI's release
// metadata. ReadWheel and Build both refuse a wheel that disagrees with it.
const WheelSHA256 = "c3c7606c27503ad8d501270406e345ddb480a7b5f38827eafe4fa82a137f0021"

// tableSHA256 is the checksum of the table Build produces from the
// pinned wheel. Build and Parse both refuse any table bytes that
// disagree with it, so a load never scores against a table this
// package's own extraction did not produce.
const tableSHA256 = "c954605b805e4a21d1d9e79b4d93bc6b89684de9996dd21720b256fc58a6ce85"

// TableFileName is the file name a table built from the pinned wheel
// is stored under.
const TableFileName = "unidecode-1.4.0.tbl"

// maxWheelSize bounds how much of a wheel download or offline file
// this package reads, well above Unidecode 1.4.0's actual size
// (235,837 bytes).
const maxWheelSize = 1 << 20 // 1 MiB

// dataModuleCount is how many data modules Unidecode 1.4.0's wheel
// ships. Build refuses a wheel that extracts to any other count.
const dataModuleCount = 190

// unidecodeMaxCodepoint is Unidecode's own upper bound: it carries no
// mapping for any codepoint above it.
const unidecodeMaxCodepoint = 0xeffff

// noneSentinel marks a table entry as unmapped. No shipped entry is
// anywhere near this long, so the length field is free to reuse as the
// marker.
const noneSentinel = 0xffff

// sweepLo and sweepHi bound the digest Build computes over every
// codepoint Unicode defines, proving the built table's lookup against
// Unidecode's own output one codepoint at a time.
const (
	sweepLo = 0x000000
	sweepHi = 0x10ffff
)

// errWheelChecksum reports that a wheel's sha256 disagreed with
// WheelSHA256, whether downloaded or read from a file.
var errWheelChecksum = errors.New("unidecode: wheel checksum mismatch")

// errTableChecksum reports that a table's sha256 disagreed with the
// pinned table checksum, whether just built or read from a file.
var errTableChecksum = errors.New("unidecode: table checksum mismatch")

// errWheelTooLarge reports that a wheel download or file exceeded
// maxWheelSize.
var errWheelTooLarge = errors.New("unidecode: wheel exceeds maximum size")
