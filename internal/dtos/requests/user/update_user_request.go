package user

import "github.com/devgugga/todo-it/internal/entities"

type UpdateUserRequest struct {
	Name   string `json:"name,omitempty" validate:"omitempty,min=2,max=100"`
	Avatar string `json:"avatar,omitempty" validate:"omitempty,url"`
}

func (r *UpdateUserRequest) ApplyToEntity(user *entities.User) {
	if r.Name != "" {
		user.Name = r.Name
	}
	if r.Avatar != "" {
		user.Avatar = r.Avatar
	}
	user.PrepareForUpdate()
}
