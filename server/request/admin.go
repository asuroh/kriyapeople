package request

// AdminRequest ....
type AdminRequest struct {
	Code           string `json:"code"`
	Name           string `json:"name" validate:"required"`
	Email          string `json:"email" validate:"email"`
	Password       string `json:"password"`
	RoleID         string `json:"role_id" validate:"required"`
	ProfileImageID string `json:"profile_image_id"`
	Status         bool   `json:"status"`
}

// AdminLoginRequest ....
type AdminLoginRequest struct {
	Email    string `json:"email" validate:"email"`
	Password string `json:"password" validate:"required"`
}
