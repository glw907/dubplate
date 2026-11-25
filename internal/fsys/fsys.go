// Package fsys holds the file system reads the server makes of a
// directory: the space free on its file system, and whether it is a
// mount point. The reads are built for Linux; elsewhere both refuse
// with [errors.ErrUnsupported].
package fsys
