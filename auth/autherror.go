package auth

type AuthError struct {
	Type    string
	Message string
}

func (e *AuthError) Error() string {
	return e.Message
}
