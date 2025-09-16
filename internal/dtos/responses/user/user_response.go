package user

import (
	"time"

	"github.com/devgugga/todo-it/internal/entities"
)

type UserResponse struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Avatar    string    `json:"avatar,omitempty"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (r *UserResponse) FromEntity(user *entities.User) {
	r.ID = user.ID.Hex()
	r.Name = user.Name
	r.Email = user.Email
	r.Avatar = user.Avatar
	r.IsActive = user.IsActive
	r.CreatedAt = user.CreatedAt
	r.UpdatedAt = user.UpdatedAt
}

func NewUserResponse(user *entities.User) *UserResponse {
	response := &UserResponse{}
	response.FromEntity(user)
	return response
}
