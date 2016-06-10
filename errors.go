package arangolite

// ErrUnique throwned when error message contains "unique constraint violated" string.
type ErrUnique struct {
	s string
}

func (e *ErrUnique) Error() string {
	return e.s
}

// ErrNotFound throwned when error message contains "not found" or "unknown collection" string.
type ErrNotFound struct {
	s string
}

func (e *ErrNotFound) Error() string {
	return e.s
}

// ErrDuplicate throwned when error message contains "duplicate name" string.
type ErrDuplicate struct {
	s string
}

func (e *ErrDuplicate) Error() string {
	return e.s
}
