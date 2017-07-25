package arangolite

type causer interface {
	Cause() error
}

type causerBehavior struct {
	cause error
}

func (e *causerBehavior) Cause() error {
	return e.cause
}

type statusCoder interface {
	StatusCode() int
}

type statusCoderBehavior struct {
	statusCode int
}

func (e *statusCoderBehavior) StatusCode() int {
	return e.statusCode
}

func withStatusCode(err error, statusCode int) error {
	return &struct {
		error
		*causerBehavior
		*statusCoderBehavior
	}{
		err,
		&causerBehavior{cause: err},
		&statusCoderBehavior{statusCode: statusCode},
	}
}

// HasStatusCode returns true when one of the given error status code matches the one returned by the database.
func HasStatusCode(err error, statusCode ...int) bool {
	for err != nil {
		e, ok := err.(statusCoder)
		if !ok {
			if cause, ok := err.(causer); ok {
				err = cause.Cause()
				continue
			} else {
				break
			}
		}
		code := e.StatusCode()
		for _, c := range statusCode {
			if code == c {
				return true
			}
		}
		break
	}
	return false
}

// GetStatusCode returns the status code encapsulated in the error.
func GetStatusCode(err error) (code int, ok bool) {
	for err != nil {
		e, ok := err.(statusCoder)
		if !ok {
			if cause, ok := err.(causer); ok {
				err = cause.Cause()
				continue
			} else {
				break
			}
		}
		return e.StatusCode(), true
	}
	return 0, false
}

type errorNumbered interface {
	ErrorNum() int
}

type errorNumberedBehavior struct {
	errorNum int
}

func (e *errorNumberedBehavior) ErrorNum() int {
	return e.errorNum
}

func withErrorNum(err error, errorNum int) error {
	return &struct {
		error
		*causerBehavior
		*errorNumberedBehavior
	}{
		err,
		&causerBehavior{cause: err},
		&errorNumberedBehavior{errorNum: errorNum},
	}
}

// HasErrorNum returns true when one of the given error num matches the one returned by the database.
func HasErrorNum(err error, errorNum ...int) bool {
	for err != nil {
		e, ok := err.(errorNumbered)
		if !ok {
			if cause, ok := err.(causer); ok {
				err = cause.Cause()
				continue
			} else {
				break
			}
		}
		num := e.ErrorNum()
		for _, n := range errorNum {
			if num == n {
				return true
			}
		}
		break
	}
	return false
}

// GetErrorNum returns the database error num encapsulated in the error.
func GetErrorNum(err error) (errorNum int, ok bool) {
	for err != nil {
		e, ok := err.(errorNumbered)
		if !ok {
			if cause, ok := err.(causer); ok {
				err = cause.Cause()
				continue
			} else {
				break
			}
		}
		return e.ErrorNum(), true
	}
	return 0, false
}

// IsErrInvalidRequest returns true when the database returns a 400.
func IsErrInvalidRequest(err error) bool {
	return HasStatusCode(err, 400)
}

// IsErrUnauthorized returns true when the database returns a 401.
func IsErrUnauthorized(err error) bool {
	return HasStatusCode(err, 401)
}

// IsErrForbidden returns true when the database returns a 403.
func IsErrForbidden(err error) bool {
	return HasStatusCode(err, 403)
}

// IsErrUnique returns true when the error num is a 1210 - ERROR_ARANGO_UNIQUE_CONSTRAINT_VIOLATED.
func IsErrUnique(err error) bool {
	return HasErrorNum(err, 1210)
}

// IsErrNotFound returns true when the database returns a 404 or when the error num is:
// 1202 - ERROR_ARANGO_DOCUMENT_NOT_FOUND
// 1203 - ERROR_ARANGO_COLLECTION_NOT_FOUND
func IsErrNotFound(err error) bool {
	return HasStatusCode(err, 404) || HasErrorNum(err, 1202, 1203)
}
