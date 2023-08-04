package webapp

// Facade to handle database operations.

type Database interface {
	// Attempt to open the database.  Fails if already open, or previously failed.
	Open()
}

const (
	DatabaseStateNew = iota
	DatabaseStateOpen
	DatabaseStateClosed
	DatabaseStateFailed
)
