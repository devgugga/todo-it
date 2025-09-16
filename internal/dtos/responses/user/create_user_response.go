package user

type UserCreatedResponse struct {
	User    UserResponse `json:"user"`
	Message string       `json:"message"`
}
