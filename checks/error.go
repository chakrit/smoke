package checks

type Error struct {
	Name string
	Err  string
}

func NewError(name, err string) *Error {
	return &Error{name, err}
}

func (e *Error) Error() string {
	if e == nil {
		return ""
	} else {
		return e.Name + ": " + e.Err
	}
}
