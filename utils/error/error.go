package error

const NOFAULT = "NoFault"

type Error struct {
	Mark string
}

func New() *Error {
	return &Error{
		Mark: NOFAULT,
	}
}
