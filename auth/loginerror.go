package auth

type LoginError struct {
	Type    string
	Message string
}

func (e *LoginError) Error() string {
	return e.Message
}
