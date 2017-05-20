package arangolite

type statusCodedError struct {
	error
	statusCode int
}

func withStatusCode(err error, statusCode int) error {
	return statusCodedError{err, statusCode}
}

// HasStatusCode returns true when one of the given error status code matches the one returned by the database.
func HasStatusCode(err error, statusCode ...int) bool {
	e, ok := err.(statusCodedError)
	if !ok {
		return false
	}
	for _, num := range statusCode {
		if e.statusCode == num {
			return true
		}
	}
	return false
}

type numberedError struct {
	error
	errorNum int
}

func withErrorNum(err error, errorNum int) error {
	return numberedError{err, errorNum}
}

// HasErrorNum returns true when one of the given error num matches the one returned by the database.
func HasErrorNum(err error, errorNum ...int) bool {
	e, ok := err.(numberedError)
	if !ok {
		return false
	}
	for _, num := range errorNum {
		if e.errorNum == num {
			return true
		}
	}
	return false
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
