package auth

type Page struct {
	Error     *AuthError
	SentEmail string
	Email     string
}

type VerifyPage struct {
	Token    string
	IsSignup bool
}

func technicalError() *AuthError {
	return &AuthError{
		Type:    "Technical",
		Message: "There was a technical error. Please try again.",
	}
}

func invalidLinkError() *AuthError {
	return &AuthError{
		Type:    "Authentication",
		Message: "That sign-in link is invalid or has expired.",
	}
}
