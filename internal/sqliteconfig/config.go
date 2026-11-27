// Package sqliteconfig shares connection settings across the host's SQLite stores.
package sqliteconfig

import "strings"

// DSN gives each connection a bounded wait for concurrent database writers.
// The project store, lifecycle store and observation outbox share a database.
func DSN(path string) string {
	separator := "?"
	if strings.Contains(path, "?") {
		separator = "&"
	}
	return path + separator + "_pragma=busy_timeout%3D5000"
}
