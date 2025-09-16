package user

import (
	"time"

	"github.com/devgugga/todo-it/internal/entities"
)

type UserProfileResponse struct {
	UserResponse
	TodosCount     int64      `json:"todos_count"`
	CompletedTodos int64      `json:"completed_todos"`
	PendingTodos   int64      `json:"pending_todos"`
	LastLoginAt    *time.Time `json:"last_login_at,omitempty"`
}

func NewUserProfileResponse(user *entities.User, todosCount, completedTodos, pendingTodos int64) *UserProfileResponse {
	return &UserProfileResponse{
		UserResponse: UserResponse{
			ID:        user.ID.Hex(),
			Name:      user.Name,
			Email:     user.Email,
			Avatar:    user.Avatar,
			IsActive:  user.IsActive,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
		},
		TodosCount:     todosCount,
		CompletedTodos: completedTodos,
		PendingTodos:   pendingTodos,
	}
}
