package main

// UserError carries an internal error along with a user friendly message.
type UserError struct {
	// Err is the underlying error for server logs.
	Err error
	// ErrorMessage is returned to the user.
	ErrorMessage string
}

func (e UserError) Error() string {
	if e.Err != nil {
		return e.Err.Error()
	}
	return e.ErrorMessage
}
