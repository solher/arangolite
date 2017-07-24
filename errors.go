package arangolite

type statusCodedError interface {
	error
	StatusCode() int
}

type statusCodedErrorBehavior struct {
	error
	statusCode int
}

func (e *statusCodedErrorBehavior) StatusCode() int {
	return e.statusCode
}

func withStatusCode(err error, statusCode int) error {
	return statusCodedErrorBehavior{err, statusCode}
}

// HasStatusCode returns true when one of the given error status code matches the one returned by the database.
func HasStatusCode(err error, statusCode ...int) bool {
	e, ok := err.(statusCodedError)
	if !ok {
		return false
	}
	code := e.StatusCode()
	for _, c := range statusCode {
		if code == c {
			return true
		}
	}
	return false
}

// GetStatusCode returns the status code encapsulated in the error.
func GetStatusCode(err error) (code int, ok bool) {
	e, ok := err.(statusCodedError)
	if !ok {
		return 0, false
	}
	return e.StatusCode(), true
}

type numberedError interface {
	error
	ErrorNum() int
}

type numberedErrorBehavior struct {
	error
	errorNum int
}

func (e *numberedErrorBehavior) ErrorNum() int {
	return e.errorNum
}

func withErrorNum(err error, errorNum int) error {
	return numberedErrorBehavior{err, errorNum}
}

// HasErrorNum returns true when one of the given error num matches the one returned by the database.
func HasErrorNum(err error, errorNum ...int) bool {
	e, ok := err.(numberedError)
	if !ok {
		return false
	}
	num := e.ErrorNum()
	for _, n := range errorNum {
		if num == n {
			return true
		}
	}
	return false
}

// GetErrorNum returns the database error num encapsulated in the error.
func GetErrorNum(err error) (errorNum int, ok bool) {
	e, ok := err.(numberedError)
	if !ok {
		return 0, false
	}
	return e.ErrorNum(), true
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
