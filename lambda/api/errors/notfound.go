package errors

type NotFoundError struct {
	Err error
}

func NewNotFoundError(err error) *NotFoundError {
	return &NotFoundError{Err: err}
}

func (e *NotFoundError) Error() string { return e.Err.Error() }
func (e *NotFoundError) Unwrap() error { return e.Err }
