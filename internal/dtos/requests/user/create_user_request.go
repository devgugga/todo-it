package user

import (
	"github.com/devgugga/todo-it/internal/entities"
	"golang.org/x/crypto/bcrypt"
)

type CreateUserRequest struct {
	Name     string `json:"name" validate:"required,min=2,max=100"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6,max=50"`
	Avatar   string `json:"avatar,omitempty" validate:"omitempty,url"`
}

func (r *CreateUserRequest) ToEntity() (*entities.User, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(r.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &entities.User{
		Name:     r.Name,
		Email:    r.Email,
		Password: string(hashedPassword),
		Avatar:   r.Avatar,
	}

	user.PrepareForCreate()
	return user, nil
}
