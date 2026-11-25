// Package doctor holds the row every health check reports to the
// doctor command. Each package that knows a fact renders its own rows
// and names them; the command only prints them in order.
package doctor

// A Row is one line of the doctor report.
type Row struct {
	// Name is the row's label, printed in the report's first column.
	Name string

	// Status is the row's text, printed after Name. It never carries
	// the value of a config.Setting registered config.Sensitive:
	// doctor's output is pasted into reports. A key outside the
	// settings family prints as its row needs, a bind address included.
	Status string

	// Failed reports whether the row makes doctor exit nonzero. A row
	// can warn in Status without failing.
	Failed bool
}
