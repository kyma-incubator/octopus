package humanerr

// Error contains a human-readable message that can be presented to a user
type Error struct {
	cause   error
	Message string
}

func NewError(cause error, msg string) *Error {
	return &Error{cause: cause, Message: msg}
}

func (e *Error) Error() string {
	return e.cause.Error()
}

func GetHumanReadableError(err error) (*Error, bool) {
	he, ok := err.(*Error)
	if !ok {
		return nil, false
	}
	return he, true
}
