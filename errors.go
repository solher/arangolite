package arangolite

type errUnique interface {
	error
	IsErrUnique()
}

type errUniqueBehavior struct{}

func (err errUniqueBehavior) IsErrUnique() {}

func withErrUnique(err error) error {
	return struct {
		error
		errUniqueBehavior
	}{
		err,
		errUniqueBehavior{},
	}
}

// IsErrUnique returns true when error message contains "unique constraint violated" string.
func IsErrUnique(err error) bool {
	_, ok := err.(errUniqueBehavior)
	return ok
}

type errNotFound interface {
	error
	IsErrNotFound()
}

type errNotFoundBehavior struct{}

func (err errNotFound) IsErrNotFound() {}

func withErrNotFound(err error) error {
	return struct {
		error
		errNotFoundBehavior
	}{
		err,
		errNotFoundBehavior{},
	}
}

// IsErrNotFound returns true when error message contains "not found" or "unknown collection" string.
func IsErrNotFound(err error) bool {
	_, ok := err.(errNotFoundBehavior)
	return ok
}

type errDuplicate interface {
	error
	IsErrDuplicate()
}

type errDuplicateBehavior struct{}

func (err errDuplicate) IsErrDuplicate() {}

func withErrDuplicate(err error) error {
	return struct {
		error
		errDuplicateBehavior
	}{
		err,
		errDuplicateBehavior{},
	}
}

// IsErrDuplicate returns true when error message contains "duplicate name" string.
func IsErrDuplicate(err error) bool {
	_, ok := err.(errDuplicateBehavior)
	return ok
}
