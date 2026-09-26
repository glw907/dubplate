// Package settings reads and seeds the store-homed settings a package
// declares through [config.Registry.AddSetting].
//
// A setting's value lives in the store's settings table. The
// environment supplies only a seed: a value the seed rule writes into
// the table when no administrator has set the key, and removes again
// when the environment stops naming it. A key the table holds no row
// for reads the registry's current default, so a changed default
// reaches every store that never overrode it.
//
// An [Accessor] serves typed reads to a running process. It caches the
// table and reloads it whenever the store's settings generation moves,
// so a write by another process is seen on the next read.
//
// [Evaluate] runs the seed rule against a store without writing:
// every setting's outcome, the refusals among them, and the rows no
// registered setting claims. The seed rule's refusals are an
// administrator's value the environment contradicts, a value that does
// not parse under its key's kind, a value in force that falls outside
// its bounds, and a correlated rule the values in force break.
package settings
