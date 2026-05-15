package auth

type SignupForm struct {
	Email                string `form:"email"`
	ConfirmationEmail    string `form:"email-confirm"`
	FirstName            string `form:"first-name"`
	LastName             string `form:"last-name"`
	Password             string `form:"password"`
	ConfirmationPassword string `form:"password-confirm"`
	DisplayName          string `form:"display-name"`
}
