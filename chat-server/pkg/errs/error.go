package errs

import "errors"

type Code int

const (
	CodeInvalid Code = iota + 1
	CodeUnauthorized
	CodeForbidden
	CodeNotFound
	CodeConflict
	CodeRateLimited
	CodeUnavailable
)

type Error struct {
	Code    Code
	Message string
	err     error
}

func (e *Error) Error() string {
	if e == nil {
		return ""
	}
	return e.Message
}

func (e *Error) Unwrap() error { return e.err }

func Invalid(msg string) error      { return &Error{Code: CodeInvalid, Message: msg} }
func Unauthorized(msg string) error { return &Error{Code: CodeUnauthorized, Message: msg} }
func Forbidden(msg string) error    { return &Error{Code: CodeForbidden, Message: msg} }
func NotFound(msg string) error     { return &Error{Code: CodeNotFound, Message: msg} }
func Conflict(msg string) error     { return &Error{Code: CodeConflict, Message: msg} }
func RateLimited(msg string) error  { return &Error{Code: CodeRateLimited, Message: msg} }
func Unavailable(msg string, err error) error {
	return &Error{Code: CodeUnavailable, Message: msg, err: err}
}

func As(err error) (*Error, bool) {
	var e *Error
	if errors.As(err, &e) {
		return e, true
	}
	return nil, false
}

func ClientMessage(err error) (string, bool) {
	e, ok := As(err)
	if !ok || e.Message == "" {
		return "", false
	}
	return e.Message, true
}
