package user

type UpdateAccountForm struct {
	Email       string `form:"email"`
	FirstName   string `form:"first-name"`
	LastName    string `form:"last-name"`
	DisplayName string `form:"display-name"`
}
