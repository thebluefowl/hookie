package queue

type Error struct {
	Err   error
	Fatal bool
}

func (e *Error) Error() string {
	return e.Err.Error()
}

func (e *Error) IsFatal() bool {
	return e.Fatal
}

func NewError(err error, isFatal bool) *Error {
	return &Error{
		Err:   err,
		Fatal: isFatal,
	}
}
