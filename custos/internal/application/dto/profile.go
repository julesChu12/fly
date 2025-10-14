package dto

// GetProfileResponse defines the response for getting user profile
type GetProfileResponse struct {
	UserID   uint   `json:"user_id"`
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar"`
	Gender   string `json:"gender"`
	Birthday string `json:"birthday,omitempty"` // Format: "2006-01-02"
	Extra    string `json:"extra,omitempty"`
}

// UpdateProfileRequest defines the request for updating user profile
type UpdateProfileRequest struct {
	Nickname string `json:"nickname" binding:"omitempty,max=64"`
	Avatar   string `json:"avatar" binding:"omitempty,max=255"`
	Gender   string `json:"gender" binding:"omitempty,oneof=male female other"`
	Birthday string `json:"birthday" binding:"omitempty"` // Format: "2006-01-02"
	Extra    string `json:"extra" binding:"omitempty"`
}

// UpdateProfileResponse defines the response for updating user profile
type UpdateProfileResponse struct {
	UserID  uint   `json:"user_id"`
	Message string `json:"message"`
}
