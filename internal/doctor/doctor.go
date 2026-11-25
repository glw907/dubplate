// Package doctor holds the row every health check reports to the
// doctor command. Each package that knows a fact renders its own rows
// and names them; the command only prints them in order.
package doctor

// A Row is one line of the doctor report.
type Row struct {
	// Name is the row's label, printed in the report's first column.
	Name string

	// Status is the row's text, printed after Name. It never carries
	// the value of a config.Setting registered config.Sensitive, since
	// doctor's output is pasted into reports. The rule binds the
	// settings family only (record 0018).
	Status string

	// Failed reports whether the row makes doctor exit nonzero. A row
	// can warn in Status without failing.
	Failed bool
}
