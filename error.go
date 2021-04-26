package gopipe

// PipeError will return errors on the process
type PipeError struct {
	Step string
	Err  error
}

func (p PipeError) Error() string {
	return p.Err.Error()
}
