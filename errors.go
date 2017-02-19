package arangolite

type errUniqueBehavior interface {
	error
	IsErrUnique()
}

type errUnique struct{ error }

func (err errUnique) IsErrUnique() {}

func withErrUnique(err error) error {
	return errUnique{err}
}

// IsErrUnique returns true when error message contains "unique constraint violated" string.
func IsErrUnique(err error) bool {
	_, ok := err.(errUniqueBehavior)
	return ok
}

type errNotFoundBehavior interface {
	error
	IsErrNotFound()
}

type errNotFound struct{ error }

func (err errNotFound) IsErrNotFound() {}

func withErrNotFound(err error) error {
	return errNotFound{err}
}

// IsErrNotFound returns true when error message contains "not found" or "unknown collection" string.
func IsErrNotFound(err error) bool {
	_, ok := err.(errNotFoundBehavior)
	return ok
}

type errDuplicateBehavior interface {
	error
	IsErrDuplicate()
}

type errDuplicate struct{ error }

func (err errDuplicate) IsErrDuplicate() {}

func withErrDuplicate(err error) error {
	return errDuplicate{err}
}

// IsErrDuplicate returns true when error message contains "duplicate name" string.
func IsErrDuplicate(err error) bool {
	_, ok := err.(errDuplicateBehavior)
	return ok
}
