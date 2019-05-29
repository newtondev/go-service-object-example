package entities

// Form is a registration request.
type Form struct {
	Email                string `json:"email" validate:"required,email"`
	Password             string `json:"password" validate:"gte=3,lte=16"`
	PasswordConfirmation string `json:"password_confirmation" validate:"gte=3,lte=16"`
}