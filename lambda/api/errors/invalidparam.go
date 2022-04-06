package errors

type InvalidParametersError struct {
	Err error
}

func NewInvalidParametersError(err error) *InvalidParametersError {
	return &InvalidParametersError{Err: err}
}

func (e *InvalidParametersError) Error() string { return e.Err.Error() }
func (e *InvalidParametersError) Unwrap() error { return e.Err }
