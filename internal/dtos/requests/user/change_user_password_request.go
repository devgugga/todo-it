package user

import "golang.org/x/crypto/bcrypt"

type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" validate:"required"`
	NewPassword     string `json:"new_password" validate:"required,min=6,max=50"`
	ConfirmPassword string `json:"confirm_password" validate:"required,eqfield=NewPassword"`
}

func (r *ChangePasswordRequest) ValidateCurrentPassword(hashedPassword string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(r.CurrentPassword))
}

func (r *ChangePasswordRequest) GetHashedNewPassword() (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(r.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}
