package user

type UserListResponse struct {
	Users      []UserResponse `json:"users"`
	Total      int64          `json:"total"`
	Page       int64          `json:"page"`
	Limit      int64          `json:"limit"`
	TotalPages int64          `json:"total_pages"`
	HasNext    bool           `json:"has_next"`
	HasPrev    bool           `json:"has_prev"`
}
