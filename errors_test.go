package arangolite

import (
	"errors"
	"fmt"
	"testing"
)

func assertEqual(t *testing.T, a interface{}, b interface{}, messages ...string) {
	var message string
	if a == b {
		return
	}
	if len(messages) == 0 {
		message = fmt.Sprintf("%v != %v", a, b)
	} else {
		message = messages[0]
	}
	t.Fatal(message)
}

func assertTrue(t *testing.T, a interface{}, message ...string) {
	assertEqual(t, true, a, message...)
}

func TestGetStatusCode(t *testing.T) {
	err := withMessage(errors.New("Error message"), "the database execution returned an error")
	err = withErrorNum(err, 1207)
	err = withStatusCode(err, 409)
	statusCode, ok := GetStatusCode(err)
	assertEqual(t, ok, true, "Failed to retrieve status code")
	assertEqual(t, statusCode, 409, "Status code does not match!")
}

func TestHasStatusCode(t *testing.T) {
	err := withMessage(errors.New("Error message"), "the database execution returned an error")
	err = withErrorNum(err, 1207)
	err = withStatusCode(err, 409)
	assertTrue(t, HasStatusCode(err, 409), "Status code does not match!")
}

func TestGetErrorNum(t *testing.T) {
	err := withMessage(errors.New("Error message"), "the database execution returned an error")
	err = withErrorNum(err, 1207)
	err = withStatusCode(err, 409)
	errNum, ok := GetErrorNum(err)
	assertEqual(t, ok, true, "Failed to retrieve error number")
	assertEqual(t, errNum, 1207, "Error code does not match!")
}

func TestHasErrorNum(t *testing.T) {
	err := withMessage(errors.New("Error message"), "the database execution returned an error")
	err = withErrorNum(err, 1207)
	err = withStatusCode(err, 409)
	assertTrue(t, HasErrorNum(err, 1207), "Error code does not match!")
}

func TestIsErrInvalidRequest(t *testing.T) {
	err := withMessage(errors.New("Error message"), "the database execution returned an error")
	err = withStatusCode(err, 400)
	assertTrue(t, IsErrInvalidRequest(err))
}

func TestIsErrUnauthorized(t *testing.T) {
	err := withMessage(errors.New("Error message"), "the database execution returned an error")
	err = withStatusCode(err, 401)
	assertTrue(t, IsErrUnauthorized(err))
}

func TestIsErrNotFound(t *testing.T) {
	err := withMessage(errors.New("Error message"), "the database execution returned an error")
	err = withErrorNum(err, 1202)
	err = withStatusCode(err, 403)
	assertTrue(t, IsErrNotFound(err))
}

func TestIsErrForbidden(t *testing.T) {
	err := withMessage(errors.New("Error message"), "the database execution returned an error")
	err = withErrorNum(err, 1202)
	err = withStatusCode(err, 403)
	assertTrue(t, IsErrForbidden(err))
}

func TestIsErrUnique(t *testing.T) {
	err := withMessage(errors.New("Error message"), "the database execution returned an error")
	err = withErrorNum(err, 1210)
	err = withStatusCode(err, 403)
	assertTrue(t, IsErrUnique(err))
}
