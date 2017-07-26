package arangolite

import (
	"fmt"
)

func withMessage(err error, message string) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %s", message, err.Error())
}

type arangoError struct {
	error
	statusCode int
	errorNum   int
}

func withStatusCode(err error, statusCode int) error {
	aError, ok := err.(arangoError)
	if ok {
		aError.statusCode = statusCode
	} else {
		aError = arangoError{error: err, statusCode: statusCode}
	}
	return aError
}

func withErrorNum(err error, errorNum int) error {
	aError, ok := err.(arangoError)
	if ok {
		aError.errorNum = errorNum
	} else {
		aError = arangoError{error: err, errorNum: errorNum}
	}
	return aError
}

// HasStatusCode returns true when one of the given error status code matches the one returned by the database.
func HasStatusCode(err error, statusCode ...int) bool {
	e, ok := err.(arangoError)
	if !ok {
		return false
	}
	code := e.statusCode
	for _, c := range statusCode {
		if code == c {
			return true
		}
		break
	}
	return false
}

// GetStatusCode returns the status code encapsulated in the error.
func GetStatusCode(err error) (code int, ok bool) {
	e, ok := err.(arangoError)
	if !ok || e.statusCode == 0 {
		return 0, false
	}
	return e.statusCode, true
}

// HasErrorNum returns true when one of the given error num matches the one returned by the database.
func HasErrorNum(err error, errorNum ...int) bool {
	e, ok := err.(arangoError)
	if !ok {
		return false
	}
	num := e.errorNum
	for _, n := range errorNum {
		if num == n {
			return true
		}
		break
	}
	return false
}

// GetErrorNum returns the database error num encapsulated in the error.
func GetErrorNum(err error) (errorNum int, ok bool) {
	e, ok := err.(arangoError)
	if !ok || e.errorNum == 0 {
		return 0, false
	}
	return e.errorNum, true
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
