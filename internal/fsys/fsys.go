// Package fsys holds the file system reads that more than one package
// makes of a directory: the space free on its file system and the
// device it sits on. The reads are built for Linux; elsewhere both
// refuse with [errors.ErrUnsupported].
package fsys
