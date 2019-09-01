package main

// Error add status method
type Error interface {
	error
	Status() int
}

// StatusError struct
type StatusError struct {
	Code int
	Err  error
}

func (se StatusError) Error() string {
	return se.Err.Error()
}

// Status code
func (se StatusError) Status() int {
	return se.Code
}
